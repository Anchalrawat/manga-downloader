package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"mangadl/internal/domain"
)

type ChapterDelegate struct {
	Selected map[int]struct{}
}

func (d ChapterDelegate) Height() int                             { return 1 }
func (d ChapterDelegate) Spacing() int                            { return 0 }
func (d ChapterDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ChapterDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	c, ok := listItem.(domain.Chapter)
	if !ok {
		return
	}

	selected := false
	if _, ok := d.Selected[c.ID]; ok {
		selected = true
	}

	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}

	check := "[ ]"
	if selected {
		check = "[x]"
	}

	// Styles
	if index != m.Index() {
		// Hide cursor logic if needed, but we use blank
	}

	checkStyle := UncheckedStyle
	textStyle := lipgloss.NewStyle().Foreground(Dim)

	if selected {
		checkStyle = CheckedStyle
		textStyle = lipgloss.NewStyle().Foreground(Foreground)
	}

	if index == m.Index() {
		textStyle = textStyle.Copy().Foreground(Pink).Bold(true)
	}

	// Render
	// format: "> [x] Chapter Name"
	fmt.Fprintf(w, "%s %s %s",
		lipgloss.NewStyle().Foreground(Pink).Render(cursor),
		checkStyle.Render(check),
		textStyle.Render(c.Name),
	)
}
