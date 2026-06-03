package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yakushevhk/macupdate/tui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			fmt.Println("MacUpdate — update all apps on macOS")
			fmt.Println()
			fmt.Println("Usage: macbee")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  --version, -v   print version")
			fmt.Println("  --help, -h      show this help")
			fmt.Println()
			fmt.Println("Keys:")
			fmt.Println("  ↑↓    navigate")
			fmt.Println("  enter open app")
			fmt.Println("  u     scan all apps")
			fmt.Println("  space select/deselect")
			fmt.Println("  U     update selected")
			fmt.Println("  f     cycle filter")
			fmt.Println("  /     search")
			fmt.Println("  q     quit")
			return
		case "--version", "-v":
			fmt.Println(version)
			return
		}
	}

	m := tui.New()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
