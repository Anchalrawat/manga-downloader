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
	Pink   = lipgloss.Color("#FF5F87") // Neon Pink
	Purple = lipgloss.Color("#AF87FF") // Soft Purple
	Cyan   = lipgloss.Color("#00D7FF") // Bright Cyan
	Green  = lipgloss.Color("#00FFAF") // Spring Green
	Red    = lipgloss.Color("#FF5F5F") // Error Red

	// Common Styles
	DocStyle = lipgloss.NewStyle().Padding(1, 2)

	// Text Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Background).
			Background(Purple).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	HeadingStyle = lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true).
			MarginBottom(1)

	SubtleStyle = lipgloss.NewStyle().Foreground(Subtle)

	// Container Styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(1)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1)

	// Input Styles
	InputPromptStyle = lipgloss.NewStyle().Foreground(Pink).Bold(true)
	InputTextStyle   = lipgloss.NewStyle().Foreground(Cyan)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().Foreground(Pink)

	// Stats Styles
	HeroStyle = lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true).
			Align(lipgloss.Center)

	StatLabelStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Width(12)

	StatValueStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// List Styles
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(Pink)
	PaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)
