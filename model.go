package main

import tea "github.com/charmbracelet/bubbletea"

type ViewState int

const (
	TextView ViewState = iota
	ModalView
)

type Model struct {
	viewState  ViewState
	tokenField TokenField
	width      int
	height     int
	index      int
}

type TokenField struct {
	tokens            []Token
	width             int
	height            int
	horizontalPadding int
	verticalPadding   int
}

type Token struct {
	delim bool
	word  string
	start int
	end   int
	line  int
	index int
}

func initialModel() *Model {
	return &Model{
		tokenField: TokenField{
			tokens:            tokenize(sample),
			horizontalPadding: 1,
		},
		viewState: TextView,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch m.viewState {
		case TextView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "l", "right":
				if m.index < len(m.tokenField.tokens)-2 {
					m.index += 2
				}
			case "h", "left":
				if m.index > 1 {
					m.index -= 2
				}
			case "j", "down":
				m.index = m.tokenField.switchFocusVertically(m.index, false)
			case "k", "up":
				m.index = m.tokenField.switchFocusVertically(m.index, true)
			case " ":
				m.viewState = ModalView
			}
		case ModalView:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "q":
				m.viewState = TextView
			}
		}
	}

	return m, nil
}
