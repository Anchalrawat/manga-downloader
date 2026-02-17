package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Err != nil {
		return m.viewError()
	}

	switch m.State {
	case StatusInput:
		return m.viewInput()
	case StatusFetching:
		return m.viewFetching()
	case StatusSelection:
		return DocStyle.Render(m.List.View())
	case StatusDownloading:
		return m.viewDownloading()
	case StatusDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

func (m Model) viewInput() string {
	logo := `
███╗   ███╗ █████╗ ███╗   ██╗ ██████╗  █████╗ 
████╗ ████║██╔══██╗████╗  ██║██╔════╝ ██╔══██╗
██╔████╔██║███████║██╔██╗ ██║██║  ███╗███████║
██║╚██╔╝██║██╔══██║██║╚██╗██║██║   ██║██╔══██║
██║ ╚═╝ ██║██║  ██║██║ ╚████║╚██████╔╝██║  ██║
╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝
`
	logo = lipgloss.NewStyle().Foreground(Pink).Render(logo)

	title := TitleStyle.Render("RIPPER V2")
	prompt := BoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			HeadingStyle.Render("Enter MangaKatana URL"),
			m.TextInput.View(),
		),
	)

	footer := SubtleStyle.Render("Press Esc to quit")

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			logo,
			title,
			"",
			prompt,
			"",
			footer,
		),
	)
}

func (m Model) viewFetching() string {
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			m.Spinner.View(),
			"  Fetching Metadata...",
		),
	)
}

func (m Model) viewDownloading() string {
	totalWidth := m.Width - 4

	pct := float64(m.DoneChapters) / float64(m.TotalChapters)
	if m.TotalChapters == 0 {
		pct = 0
	}

	pctStr := fmt.Sprintf("%.0f%%", pct*100)
	heroBlock := HeroStyle.
		Width(totalWidth).
		Height(3).
		Render(pctStr)

	barWidth := totalWidth - 10
	if barWidth < 10 {
		barWidth = 10
	}
	m.Progress.Width = barWidth

	barBlock := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(totalWidth).
		Render(m.Progress.View())

	elapsed := time.Since(m.StartTime).Round(time.Second)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		StatLabelStyle.Render("ELAPSED"),
		StatValueStyle.Render(elapsed.String()),
	)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		StatLabelStyle.Render("PROGRESS"),
		StatValueStyle.Render(fmt.Sprintf("%d/%d", m.DoneChapters, m.TotalChapters)),
		"   ", // spacer
		StatLabelStyle.Render("STATUS"),
		StatValueStyle.Foreground(Green).Render("Active"),
	)

	statsBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Dim).
		Padding(1, 2).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Left, row1, row2),
		)

	var logLines []string
	logCount := len(m.Logs)
	for i := 0; i < 3; i++ {
		if logCount > i {
			idx := logCount - 1 - i
			line := m.Logs[idx]
			style := lipgloss.NewStyle().Foreground(Foreground)
			if i == 1 {
				style = style.Foreground(Subtle)
			}
			if i == 2 {
				style = style.Foreground(Dim)
			}
			logLines = append(logLines, style.Render(line))
		}
	}

	logContent := ""
	if len(logLines) > 0 {
		// Just showing them in order
		logContent = strings.Join(logLines, "\n")
	}

	logBlock := lipgloss.NewStyle().
		MarginTop(1).
		Width(totalWidth).
		Render(logContent)

	ui := lipgloss.JoinVertical(lipgloss.Center,
		heroBlock,
		barBlock,
		"", // spacer
		statsBlock,
		logBlock,
	)

	return DocStyle.Render(ui)
}

func (m Model) viewDone() string {
	box := FocusedBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			HeadingStyle.Render("DOWNLOAD COMPLETE"),
			"",
			lipgloss.NewStyle().Foreground(Green).Render("✓ All tasks finished successfully."),
			"",
			"Files saved to ./output/",
			"",
			SubtleStyle.Render("Press Enter to quit"),
		),
	)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) viewError() string {
	box := BoxStyle.BorderForeground(Red).Render(
		lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(Red).Bold(true).Render("ERROR"),
			"",
			m.Err.Error(),
			"",
			SubtleStyle.Render("Press Ctrl+C to quit"),
		),
	)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
}
