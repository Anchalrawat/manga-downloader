package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go-manga-ripper/internal/domain"
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

	str := fmt.Sprintf("%d. %s", c.ID+1, c.Name)

	// Check state
	checked := false
	if _, ok := d.Selected[c.ID]; ok {
		checked = true
	}

	fn := ItemStyle.Render

	// Icon selection
	icon := "[ ]"
	if checked {
		icon = "[✓]"
	}

	// Cursor selection
	if index == m.Index() {
		fn = func(s ...string) string {
			return SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	// Apply colors to the icon/text based on state
	if checked {
		icon = lipgloss.NewStyle().Foreground(Cyan).Render(icon)
		str = lipgloss.NewStyle().Foreground(Foreground).Render(str)
	} else {
		icon = lipgloss.NewStyle().Foreground(Subtle).Render(icon)
		str = lipgloss.NewStyle().Foreground(Subtle).Render(str)
	}

	if index == m.Index() {
		str = lipgloss.NewStyle().Bold(true).Foreground(Pink).Render(str)
	}

	fmt.Fprint(w, fn(icon, str))
}
