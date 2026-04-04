package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const translateUrl = "http://127.0.0.1:5000/translate"

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
