package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"go-manga-ripper/internal/domain"
	"go-manga-ripper/internal/downloader"
	"go-manga-ripper/internal/scraper"
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
			if m.List.FilterState() == list.Filtering {
				break // Let list handle inputs
			}

			switch msg.String() {
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
				if i, ok := m.List.SelectedItem().(domain.Chapter); ok {
					if _, exists := m.Selected[i.ID]; exists {
						delete(m.Selected, i.ID)
					} else {
						m.Selected[i.ID] = struct{}{}
					}
				}

			case "a":
				// Toggle All
				if len(m.Selected) == len(m.Manga.Chapters) {
					m.Selected = make(map[int]struct{})
				} else {
					for _, c := range m.Manga.Chapters {
						m.Selected[c.ID] = struct{}{}
					}
				}
			}

		case StatusDone:
			if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc || msg.String() == "q" {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Resize List
		h, v := DocStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v-2)

		// Resize Progress
		targetWidth := (m.Width / 2) - 10
		if targetWidth < 20 {
			targetWidth = 20
		}
		m.Progress.Width = targetWidth

		// Resize Viewport
		m.Viewport.Width = msg.Width - 10
		m.Viewport.Height = msg.Height - 15

	case MangaFetchedMsg:
		m.Manga = msg
		m.State = StatusSelection

		items := make([]list.Item, len(m.Manga.Chapters))
		for i, c := range m.Manga.Chapters {
			c.ID = i // Ensure ID matches index
			items[i] = c
			// Default: Select All
			m.Selected[c.ID] = struct{}{}
		}
		m.List.SetItems(items)

		// Update delegate reference
		d := ChapterDelegate{Selected: m.Selected}
		m.List.SetDelegate(d)

	case ErrMsg:
		m.Err = msg
		m.State = StatusError

	case ProgressMsg:
		if msg.Done == -1 {
			m.DoneChapters++
		} else {
			m.DoneChapters = msg.Done
		}

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

	if m.State == StatusSelection {
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.State == StatusDownloading {
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)
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
		completed := 0

		var wg sync.WaitGroup
		sem := make(chan struct{}, 10) // Concurrency limit

		for _, chapter := range chapters {
			wg.Add(1)
			go func(ch domain.Chapter) {
				defer wg.Done()
				sem <- struct{}{}

				select {
				case downloadChan <- ProgressMsg{Done: completed, Total: total, Message: fmt.Sprintf("Started: %s", ch.Name)}:
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
