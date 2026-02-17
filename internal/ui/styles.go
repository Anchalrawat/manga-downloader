package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Terminal.shop inspired palette
	Background = lipgloss.Color("#000000")
	Foreground = lipgloss.Color("#EAEAEA")
	Subtle     = lipgloss.Color("#666666")
	Dim        = lipgloss.Color("#333333")
	// Accents
	Primary = lipgloss.Color("#5B8AF0") // Calm Blue
	Accent  = lipgloss.Color("#42E66C") // Fresh Mint

	// Mapped Palette
	Pink   = Primary                   // Maps to Blue
	Purple = Primary                   // Maps to Blue
	Cyan   = Accent                    // Maps to Mint
	Green  = Accent                    // Maps to Mint
	Red    = lipgloss.Color("#FF5555") // Soft Error Red
	Orange = Accent                    // Maps to Mint

	// Layout Styles
	DocStyle = lipgloss.NewStyle().Padding(1, 2)

	// App Frame
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Background).
			Background(Purple).
			Bold(true).
			Padding(0, 1)

	SubtleStyle = lipgloss.NewStyle().Foreground(Subtle)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			PaddingTop(1)

	// Dashboard / Input
	LogoStyle = lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true).
			MarginBottom(1)

	InputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1).
			Align(lipgloss.Center)

	InputPromptStyle = lipgloss.NewStyle().Foreground(Pink).Bold(true)
	InputTextStyle   = lipgloss.NewStyle().Foreground(Cyan)

	TipsStyle = lipgloss.NewStyle().
			Foreground(Dim).
			Italic(true).
			MarginTop(1)

	// List / Selection
	TitleStyle = lipgloss.NewStyle().
			Foreground(Background).
			Background(Purple).
			Bold(true).
			Padding(0, 1)

	ItemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(Pink)
	CheckedStyle      = lipgloss.NewStyle().Foreground(Green).Bold(true)
	UncheckedStyle    = lipgloss.NewStyle().Foreground(Dim)

	PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	// Download Monitor
	ProgressContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Dim).
				Padding(0, 1).
				MarginBottom(1)

	WorkerGridStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1).
			Height(12) // Fixed height for grid

	LogContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Dim).
				Padding(0, 1)

	StatLabelStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Width(10)

	StatValueStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	SpinnerStyle = lipgloss.NewStyle().Foreground(Pink)
)
