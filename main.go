package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"mangadl/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
