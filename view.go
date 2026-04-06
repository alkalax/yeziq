package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (tf *TokenField) View(width, height, focusedToken int, multiselect bool, multistart int) string {
	tf.width = width - 2
	tf.height = height - 2

	return defaultStyles().tokenField.
		Width(tf.width).
		Height(tf.height).
		Padding(tf.verticalPadding, tf.horizontalPadding).
		Render(tf.renderTokens(focusedToken, multiselect, multistart))
}

func (tf *TokenField) ViewModal(selected int, multiselect bool, multistart int) string {
	translations, err := getTranslations(tf.getWordSelection(selected, multiselect, multistart), DeepL)
	var renderedTranslations string
	if err != nil {
		renderedTranslations = err.Error()
	} else {
		renderedTranslations = strings.Join(translations, "\n")
	}

	var sb strings.Builder
	sb.WriteString(tf.getSentence(selected))
	sb.WriteString("\n---\n")
	if !multiselect && tf.tokens[selected].translation == "" {
		sb.WriteString(defaultStyles().modalNoTranslation.Render("\nNo translation selected.\n"))
		sb.WriteString("\n---\n")
	}
	sb.WriteString(renderedTranslations)

	return sb.String()
}

func (m *Model) View() string {
	switch m.viewState {
	case TextView:
		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Bottom,
			m.tokenField.View(m.width/2, m.height*7/8, m.index, m.multiselect, m.start),
		)
	case ModalView:
		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Center,
			defaultStyles().modal.
				Width(m.width/3).
				Height(m.height/3).
				Render(m.tokenField.ViewModal(m.index, m.multiselect, m.start)),
		)
	default:
		return ""
	}
}
