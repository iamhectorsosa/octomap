package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iamhectorsosa/octomap/internal/config"
	"github.com/iamhectorsosa/octomap/internal/model"
)

func main() {
	cfg, err := config.New(os.Args)
	if err != nil {
		fmt.Printf("Error handling arguments, %v", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model.New(cfg)).Run(); err != nil {
		fmt.Println("Error starting Bubble Tea program:", err)
		os.Exit(1)
	}
}
