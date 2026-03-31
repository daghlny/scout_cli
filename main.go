package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/daghlny/scout_cli/pkg/tui"
)

func main() {
	aiMode := flag.String("ai", "smart", "AI mode: smart or llm")
	flag.Parse()

	app := tui.NewApp(*aiMode)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
