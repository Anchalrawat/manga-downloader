package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	var content string

	if m.Err != nil {
		content = m.viewError()
	} else {
		switch m.State {
		case StatusInput:
			content = m.viewInput()
		case StatusFetching:
			content = m.viewFetching()
		case StatusSelection:
			content = m.viewSelection()
		case StatusDownloading:
			content = m.viewDownloading()
		case StatusDone:
			content = m.viewDone()
		default:
			content = "Unknown state"
		}
	}

	// App Frame
	// header := HeaderStyle.Render(" MANGADL v1.0 ")

	// Status Line (Right aligned)
	statusText := "IDLE"
	statusColor := Dim
	switch m.State {
	case StatusInput:
		statusText = "WAITING FOR INPUT"
	case StatusFetching:
		statusText = "FETCHING DATA"
		statusColor = Orange
	case StatusSelection:
		statusText = "SELECTION"
		statusColor = Purple
	case StatusDownloading:
		statusText = "DOWNLOADING"
		statusColor = Green
	case StatusDone:
		statusText = "COMPLETED"
		statusColor = Pink
	case StatusError:
		statusText = "ERROR"
		statusColor = Red
	}

	status := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText)

	gap := m.Width - lipgloss.Width(status) - 2
	if gap < 0 {
		gap = 0
	}
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, strings.Repeat(" ", gap), status)

	footer := FooterStyle.Render(" Ctrl+C: Quit • Esc: Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		headerBar,
		DocStyle.Render(content),
		footer,
	)
}

func (m Model) viewInput() string {
	var logo string
	// Only show logo if we have enough height
	if m.Height > 20 {
		logoContent := `
███╗   ███╗ █████╗ ███╗   ██╗ ██████╗  █████╗ ██████╗ ██╗
████╗ ████║██╔══██╗████╗  ██║██╔════╝ ██╔══██╗██╔══██╗██║
██╔████╔██║███████║██╔██╗ ██║██║  ███╗███████║██║  ██║██║
██║╚██╔╝██║██╔══██║██║╚██╗██║██║   ██║██╔══██║██║  ██║██║
██║ ╚═╝ ██║██║  ██║██║ ╚████║╚██████╔╝██║  ██║██████╔╝███████╗
╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝`
		logo = LogoStyle.Render(logoContent)
	}

	// Dynamic width for input box
	inputWidth := 60
	if m.Width < 64 {
		inputWidth = m.Width - 4
	}

	input := InputBoxStyle.Width(inputWidth).Render(
		lipgloss.JoinVertical(lipgloss.Center,
			InputPromptStyle.Render("ENTER MANGA URL"),
			m.TextInput.View(),
		),
	)

	tips := TipsStyle.Render("Supported Sites: MangaKatana\nExample: https://mangakatana.com/manga/one-piece.20")

	// Hide tips if height is very constrained
	if m.Height < 15 {
		tips = ""
	}

	return lipgloss.Place(m.Width, max(0, m.Height-5), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			logo,
			"",
			input,
			tips,
		),
	)
}

func (m Model) viewFetching() string {
	return lipgloss.Place(m.Width, max(0, m.Height-5), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			m.Spinner.View(),
			"  Fetching Metadata...",
		),
	)
}

func (m Model) viewSelection() string {
	// 1. Filter Input (at top if active)
	var filterView string
	if m.FilterInput.Focused() || m.FilterInput.Value() != "" {
		filterView = lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(Pink).Bold(true).Render("FILTER: "),
			m.FilterInput.View(),
		)
		filterView = lipgloss.NewStyle().MarginBottom(1).Render(filterView)
	}

	// 2. Grid Content
	if len(m.FilteredChapters) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			filterView,
			"No chapters found.",
		)
	}

	// Determine visible range
	startRow := m.SelectionOffset

	// Determine available rows
	// Height - Header(1) - Footer(1) - DocPadding(2) - FilterHeight
	availHeight := m.Height - 4
	if filterView != "" {
		availHeight -= 2 // Margin + height
	}

	if availHeight < 1 {
		availHeight = 1
	}

	endRow := startRow + availHeight

	startIndex := startRow * m.SelectionColumns
	endIndex := endRow * m.SelectionColumns

	if startIndex >= len(m.FilteredChapters) {
		startIndex = len(m.FilteredChapters)
	}
	if startIndex < 0 {
		startIndex = 0
	}

	if endIndex > len(m.FilteredChapters) {
		endIndex = len(m.FilteredChapters)
	}

	var rows []string
	var currentRow []string

	// Calculate column width
	colWidth := 30
	if m.SelectionColumns > 0 {
		// (Width - Padding) / Columns
		colWidth = (m.Width - 4) / m.SelectionColumns
	}

	for i := startIndex; i < endIndex; i++ {
		c := m.FilteredChapters[i]

		// Render Item
		isSelected := false
		if _, ok := m.Selected[c.ID]; ok {
			isSelected = true
		}

		isCursor := (i == m.SelectionCursor)

		// Styles
		check := "[ ]"
		if isSelected {
			check = "[x]"
		}

		style := lipgloss.NewStyle().Width(colWidth).PaddingRight(1)

		checkStyle := UncheckedStyle
		nameStyle := lipgloss.NewStyle().Foreground(Dim)

		if isSelected {
			checkStyle = CheckedStyle
			nameStyle = lipgloss.NewStyle().Foreground(Foreground)
		}

		if isCursor {
			nameStyle = nameStyle.Copy().Foreground(Pink).Bold(true)
			checkStyle = checkStyle.Copy().Foreground(Pink)
		}

		// Truncate name if too long
		maxNameLen := colWidth - 6 // [x] + padding space
		if maxNameLen < 5 {
			maxNameLen = 5
		}

		name := c.Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		itemStr := fmt.Sprintf("%s %s", checkStyle.Render(check), nameStyle.Render(name))

		// If cursor is on this item, we can also underline or change bg?
		// For now text color change is enough.

		currentRow = append(currentRow, style.Render(itemStr))

		if len(currentRow) == m.SelectionColumns {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	// Flush remaining row
	if len(currentRow) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	gridContent := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(lipgloss.Left,
		filterView,
		gridContent,
	)
}

func (m Model) viewDownloading() string {
	// Top: Progress
	m.Progress.Width = max(0, m.Width-10)
	// Use View() as state is updated in Update()
	progView := m.Progress.View()

	stats := lipgloss.JoinHorizontal(lipgloss.Top,
		StatLabelStyle.Render("PROGRESS"),
		StatValueStyle.Render(fmt.Sprintf("%d/%d", m.DoneChapters, m.TotalChapters)),
		"   ",
		StatLabelStyle.Render("ELAPSED"),
		StatValueStyle.Render(time.Since(m.StartTime).Round(time.Second).String()),
	)

	topBlock := ProgressContainerStyle.Width(max(0, m.Width-4)).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			stats,
			"",
			progView,
		),
	)

	// Calculate available height for logs
	// Total available content height (from View) is m.Height - 5 (header/footer)
	// But lipgloss.Place adds padding/alignment, so we just use the raw calculation here.
	// Let's assume we want to fill the remaining vertical space.

	// We need to account for the chrome (header/footer) which is handled in the main View() function
	// The main View() passes m.Width and m.Height to us implicitly via 'm'.
	// But View() wraps this content in DocStyle.Render(content). DocStyle has Padding(1, 2).
	// So effective content area is m.Height - 5 (header/footer) - 2 (vertical padding) = m.Height - 7.

	availableHeight := m.Height - 7
	topHeight := lipgloss.Height(topBlock)
	logHeight := availableHeight - topHeight - 1 // -1 for spacing

	// Middle: Logs
	var logBlock string

	if logHeight > 2 {
		var logLines []string

		// Calculate available lines for log content (subtracting border and header line)
		// Border: 2
		// Header "LOGS": 1
		// Total overhead: 3
		contentLines := logHeight - 3
		if contentLines < 0 {
			contentLines = 0
		}

		// Show last N logs
		start := 0
		if len(m.Logs) > contentLines {
			start = len(m.Logs) - contentLines
		}

		for i := start; i < len(m.Logs); i++ {
			logLines = append(logLines, m.Logs[i])
		}
		logContent := strings.Join(logLines, "\n")

		logBlock = LogContainerStyle.
			Width(max(0, m.Width-4)).
			Height(logHeight).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					lipgloss.NewStyle().Foreground(Subtle).Render("LOGS"),
					logContent,
				),
			)
	} else {
		// Not enough space for logs
		logBlock = ""
	}

	// Safety: ensure we don't render something that exceeds available height
	// This might happen if logBlock calc was slightly off due to margins.
	// But lipgloss should handle it mostly.

	// JoinVertical might add a newline if logBlock is empty string? No.
	// But if it's not empty, it adds a newline between blocks?

	if logBlock == "" {
		return topBlock
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		topBlock,
		logBlock,
	)
}

func (m Model) viewDone() string {
	boxWidth := 60
	if m.Width < 64 {
		boxWidth = m.Width - 4
	}

	box := InputBoxStyle.
		Width(boxWidth).
		BorderForeground(Green).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(Green).Bold(true).Render("DOWNLOAD COMPLETE"),
				"",
				"Files saved to ./output/",
				"",
				SubtleStyle.Render("Press Enter to quit"),
			),
		)

	return lipgloss.Place(m.Width, max(0, m.Height-5), lipgloss.Center, lipgloss.Center, box)
}

func (m Model) viewError() string {
	boxWidth := 60
	if m.Width < 64 {
		boxWidth = m.Width - 4
	}

	box := InputBoxStyle.
		Width(boxWidth).
		BorderForeground(Red).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(Red).Bold(true).Render("ERROR"),
				"",
				m.Err.Error(),
				"",
				SubtleStyle.Render("Press Ctrl+C to quit"),
			),
		)
	return lipgloss.Place(m.Width, max(0, m.Height-5), lipgloss.Center, lipgloss.Center, box)
}
