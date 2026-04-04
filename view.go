package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (tf *TokenField) View(width, height, focusedToken int) string {
	tf.width = width - 2
	tf.height = height - 2

	return defaultStyles().tokenField.
		Width(tf.width).
		Height(tf.height).
		Padding(tf.verticalPadding, tf.horizontalPadding).
		Render(tf.renderTokens(focusedToken))
}

func (tf *TokenField) ViewModal(selected int) string {
	translations, err := getTranslations(tf.tokens[selected].word)
	var renderedTranslations string
	if err != nil {
		renderedTranslations = err.Error()
	} else {
		renderedTranslations = strings.Join(translations, "\n")
	}
	return tf.getSentence(selected) + "\n\n" + renderedTranslations
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
			defaultStyles().modal.
				Width(m.width/3).
				Height(m.height/3).
				Render(m.tokenField.ViewModal(m.index)),
		)
	default:
		return ""
	}
}
