package main

import "github.com/charmbracelet/lipgloss"

func (tf *TokenField) View(width, height, focusedToken int) string {
	tf.width = width - 2
	tf.height = height - 2

	return lipgloss.NewStyle().
		Width(tf.width).
		Height(tf.height).
		Padding(tf.verticalPadding, tf.horizontalPadding).
		Border(lipgloss.NormalBorder()).
		Render(tf.renderTokens(focusedToken))
}

func (tf *TokenField) ViewModal(selected int) string {
	//return tf.tokens[selected].word
	return tf.getSentence(selected)
}

func (m *Model) View() string {
	switch m.viewState {
	case TextView:
		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Bottom,
			m.tokenField.View(m.width/2, m.height*7/8, m.index),
		)
	case ModalView:
		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Width(m.width/3).Height(m.height/3).
				Border(lipgloss.NormalBorder()).
				Align(lipgloss.Center, lipgloss.Center).
				Render(m.tokenField.ViewModal(m.index)),
		)
	default:
		return ""
	}
}
