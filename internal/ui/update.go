package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"mangadl/internal/domain"
	"mangadl/internal/downloader"
	"mangadl/internal/scraper"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global Quit
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Contextual Key Handling
		switch m.State {
		case StatusInput:
			if msg.Type == tea.KeyEnter {
				if m.TextInput.Value() != "" {
					m.State = StatusFetching
					return m, fetchMangaCmd(m.TextInput.Value())
				}
			}
			if msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}

		case StatusSelection:
			// Handle Filter Input
			if m.FilterInput.Focused() {
				switch msg.Type {
				case tea.KeyEnter, tea.KeyEsc:
					m.FilterInput.Blur()
					return m, nil
				}

				var cmd tea.Cmd
				m.FilterInput, cmd = m.FilterInput.Update(msg)
				m.updateFilteredChapters()
				return m, cmd
			}

			switch msg.String() {
			case "/":
				m.FilterInput.Focus()
				return m, textinput.Blink

			case "enter":
				// Start Download
				chapters := m.getSelectedChapters()
				if len(chapters) > 0 {
					m.State = StatusDownloading
					m.TotalChapters = len(chapters)
					m.DoneChapters = 0
					m.StartTime = time.Now()
					m.addLog("Initializing download sequence...")
					return m, startDownload(chapters, m.Manga.Title)
				}

			case " ":
				if len(m.FilteredChapters) > 0 {
					idx := m.SelectionCursor
					if idx >= 0 && idx < len(m.FilteredChapters) {
						c := m.FilteredChapters[idx]
						if _, exists := m.Selected[c.ID]; exists {
							delete(m.Selected, c.ID)
						} else {
							m.Selected[c.ID] = struct{}{}
						}
					}
				}

			case "a":
				// Toggle All (visible)
				allSelected := true
				for _, c := range m.FilteredChapters {
					if _, ok := m.Selected[c.ID]; !ok {
						allSelected = false
						break
					}
				}

				if allSelected {
					for _, c := range m.FilteredChapters {
						delete(m.Selected, c.ID)
					}
				} else {
					for _, c := range m.FilteredChapters {
						m.Selected[c.ID] = struct{}{}
					}
				}

			// Grid Navigation
			case "up", "k":
				m.moveCursor(-m.SelectionColumns)
			case "down", "j":
				m.moveCursor(m.SelectionColumns)
			case "left", "h":
				m.moveCursor(-1)
			case "right", "l":
				m.moveCursor(1)
			}

		case StatusDone:
			if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc || msg.String() == "q" {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Recalculate layout
		m.recalcLayout()

		// Resize Progress
		targetWidth := (m.Width / 2) - 10
		if targetWidth < 20 {
			targetWidth = 20
		}
		m.Progress.Width = targetWidth

	case MangaFetchedMsg:
		m.Manga = msg
		m.State = StatusSelection
		m.FilteredChapters = m.Manga.Chapters
		m.SelectionCursor = 0
		m.recalcLayout()

		// Default: Select All
		m.Selected = make(map[int]struct{})
		for _, c := range m.Manga.Chapters {
			m.Selected[c.ID] = struct{}{}
		}

	case ErrMsg:
		m.Err = msg
		m.State = StatusError

	case ProgressMsg:
		if msg.Done == -1 {
			m.DoneChapters++
		} else if msg.Done >= 0 {
			m.DoneChapters = msg.Done
		}
		// If msg.Done == -2, do not change DoneChapters

		if msg.Message != "" {
			m.CurrentStatus = msg.Message
			m.addLog(msg.Message)
		}

		pct := float64(m.DoneChapters) / float64(m.TotalChapters)
		if pct > 1.0 {
			pct = 1.0
		}
		if m.TotalChapters == 0 {
			pct = 0
		}

		cmd = m.Progress.SetPercent(pct)
		return m, tea.Batch(cmd, waitForDownloadMsg)

	case DownloadCompleteMsg:
		m.State = StatusDone
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Route updates
	var cmds []tea.Cmd

	if m.State == StatusInput {
		m.TextInput, cmd = m.TextInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.State == StatusDownloading {
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Update progress bar (animations, resizing, etc)
		var pCmd tea.Cmd
		var pModel tea.Model
		pModel, pCmd = m.Progress.Update(msg)
		m.Progress = pModel.(progress.Model)
		cmds = append(cmds, pCmd)
	}

	return m, tea.Batch(cmds...)
}

func fetchMangaCmd(url string) tea.Cmd {
	return func() tea.Msg {
		details, err := scraper.FetchMangaDetails(url)
		if err != nil {
			return ErrMsg(err)
		}
		return MangaFetchedMsg(details)
	}
}

var downloadChan chan ProgressMsg

func startDownload(chapters []domain.Chapter, title string) tea.Cmd {
	downloadChan = make(chan ProgressMsg, 100)

	go func() {
		mangaDir := downloader.SanitizeFilename(title)
		total := len(chapters)

		var wg sync.WaitGroup
		sem := make(chan struct{}, 10) // Concurrency limit

		for _, chapter := range chapters {
			wg.Add(1)
			go func(ch domain.Chapter) {
				defer wg.Done()
				sem <- struct{}{}

				select {
				case downloadChan <- ProgressMsg{Done: -2, Total: total, Message: fmt.Sprintf("Started: %s", ch.Name)}:
				default:
				}

				err := downloader.DownloadChapter(ch.URL, ch.Name, mangaDir)
				<-sem

				msg := fmt.Sprintf("Finished: %s", ch.Name)
				if err != nil {
					msg = fmt.Sprintf("Failed: %s (%v)", ch.Name, err)
				}

				downloadChan <- ProgressMsg{Done: -1, Total: total, Message: msg}
			}(chapter)
		}

		wg.Wait()
		close(downloadChan)
	}()

	return waitForDownloadMsg
}

func waitForDownloadMsg() tea.Msg {
	msg, ok := <-downloadChan
	if !ok {
		return DownloadCompleteMsg{}
	}
	return msg
}

func (m *Model) getSelectedChapters() []domain.Chapter {
	var chaps []domain.Chapter
	for _, c := range m.Manga.Chapters {
		if _, ok := m.Selected[c.ID]; ok {
			chaps = append(chaps, c)
		}
	}
	return chaps
}

func (m *Model) addLog(msg string) {
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] %s", ts, msg)
	m.Logs = append(m.Logs, entry)

	maxLogs := 20
	if len(m.Logs) > maxLogs {
		m.Logs = m.Logs[len(m.Logs)-maxLogs:]
	}
}

func (m *Model) recalcLayout() {
	if m.Width == 0 {
		return
	}

	// Layout params
	availableWidth := m.Width - 4 // DocStyle padding
	minItemWidth := 35            // Slight increase for better readability

	cols := availableWidth / minItemWidth
	if cols < 1 {
		cols = 1
	}

	m.SelectionColumns = cols

	// Ensure cursor is valid
	if len(m.FilteredChapters) > 0 {
		if m.SelectionCursor >= len(m.FilteredChapters) {
			m.SelectionCursor = len(m.FilteredChapters) - 1
		}
		if m.SelectionCursor < 0 {
			m.SelectionCursor = 0
		}
	}

	// Recalculate offset to ensure cursor is visible
	m.moveCursor(0)
}

func (m *Model) updateFilteredChapters() {
	if m.FilterInput.Value() == "" {
		m.FilteredChapters = m.Manga.Chapters
	} else {
		filter := strings.ToLower(m.FilterInput.Value())
		var res []domain.Chapter
		for _, c := range m.Manga.Chapters {
			if strings.Contains(strings.ToLower(c.Name), filter) {
				res = append(res, c)
			}
		}
		m.FilteredChapters = res
	}
	m.SelectionCursor = 0
	m.recalcLayout()
}

func (m *Model) moveCursor(delta int) {
	if len(m.FilteredChapters) == 0 {
		return
	}

	newCursor := m.SelectionCursor + delta
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= len(m.FilteredChapters) {
		newCursor = len(m.FilteredChapters) - 1
	}
	m.SelectionCursor = newCursor

	// Handle scrolling
	// Calculate visible rows
	// Header (1) + Footer (1) + DocPadding (2) = 4
	// Filter bar takes space if visible
	// Let's assume ~5 lines of overhead

	overhead := 5
	// If filter is shown (it's always shown in grid view header?)

	visibleRows := m.Height - overhead
	if visibleRows < 1 {
		visibleRows = 1
	}

	currentRow := m.SelectionCursor / m.SelectionColumns

	if currentRow < m.SelectionOffset {
		m.SelectionOffset = currentRow
	}

	if currentRow >= m.SelectionOffset+visibleRows {
		m.SelectionOffset = currentRow - visibleRows + 1
	}
}
