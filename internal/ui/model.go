package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"mangadl/internal/domain"
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
	State       Status
	TextInput   textinput.Model
	FilterInput textinput.Model
	Spinner     spinner.Model
	Progress    progress.Model
	Viewport    viewport.Model

	// Logs for the dashboard

	Logs []string

	Manga *domain.MangaDetails
	Err   error

	// Selection state
	Selected         map[int]struct{} // Key is Chapter.ID
	FilteredChapters []domain.Chapter
	SelectionCursor  int
	SelectionOffset  int
	SelectionColumns int
	SelectionRows    int

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

	fi := textinput.New()
	fi.Placeholder = "Filter chapters..."
	fi.CharLimit = 50
	fi.Width = 30
	// Style similar to main input but smaller?
	fi.Prompt = "/ "

	return Model{
		State:       StatusInput,
		TextInput:   ti,
		FilterInput: fi,
		Spinner:     s,
		Progress:    prog,
		Selected:    make(map[int]struct{}),
		Logs:        []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.Spinner.Tick)
}
