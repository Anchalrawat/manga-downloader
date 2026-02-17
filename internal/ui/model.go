package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"go-manga-ripper/internal/domain"
)

type Status int

const (
	StatusInput Status = iota
	StatusFetching
	StatusSelection
	StatusDownloading
	StatusDone
	StatusError
)

type Model struct {
	State     Status
	TextInput textinput.Model
	Spinner   spinner.Model
	Progress  progress.Model
	List      list.Model
	Viewport  viewport.Model

	// Logs for the dashboard
	Logs []string

	Manga *domain.MangaDetails
	Err   error

	// Selection state
	Selected map[int]struct{} // Key is Chapter.ID

	// Download state
	TotalChapters int
	DoneChapters  int
	CurrentStatus string
	StartTime     time.Time

	// Window size
	Width  int
	Height int
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Paste URL here..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50
	ti.PromptStyle = InputPromptStyle
	ti.TextStyle = InputTextStyle

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = SpinnerStyle

	prog := progress.New(
		progress.WithGradient(string(Pink), string(Cyan)),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Setup List
	delegate := ChapterDelegate{Selected: make(map[int]struct{})}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Chapters"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle
	l.Styles.PaginationStyle = PaginationStyle
	l.Styles.HelpStyle = HelpStyle

	return Model{
		State:     StatusInput,
		TextInput: ti,
		Spinner:   s,
		Progress:  prog,
		List:      l,
		Selected:  delegate.Selected, // Share the map reference
		Logs:      []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.Spinner.Tick)
}
