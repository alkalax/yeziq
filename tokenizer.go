package main

import (
	"log"
	"math"
	"regexp"
	"strings"
)

var sample string = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

var sample2 string = "asdfsadf sadfsadf. asdfsadf. asdfsadfsdf? asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?asdfsadf sadfsadf. asdfsadf. asdfsadfsdf? asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?"

func tokenize(text string) []Token {
	re := regexp.MustCompile(`[,.?!]?\s+`)
	words := re.Split(text, -1)
	delims := re.FindAllString(text, -1)

	tokens := []Token{}
	for i, word := range words {
		if i < len(words)-1 {
			tokens = append(tokens, Token{word: word})
		} else if strings.ContainsRune(",.?!", rune(word[len(word)-1])) {
			// Edge case for delimiter at the end of text
			tokens = append(tokens, Token{word: word[0 : len(word)-1]})
			tokens = append(tokens, Token{word: word[len(word)-1:], delim: true})
		}
		if i < len(delims) {
			tokens = append(tokens, Token{word: delims[i], delim: true})
		}
	}

	return tokens
}

func (tf *TokenField) switchFocusVertically(currentIndex int, up bool) int {
	currToken := 0
	for i, token := range tf.tokens {
		if token.index == currentIndex {
			currToken = i
			break
		}
	}

	focusedToken := tf.tokens[currToken]
	lastLine := tf.tokens[len(tf.tokens)-1].line
	if (up && focusedToken.line == 0) || (!up && focusedToken.line == lastLine) {
		return currentIndex
	}

	newLine := focusedToken.line
	if up {
		newLine--
	} else {
		newLine++
	}

	anchorIndex := (focusedToken.start + focusedToken.end) / 2

	candidate := 0
	for i, token := range tf.tokens {
		if token.line == newLine && !token.delim {
			candidate = i
			break
		}
	}

	for {
		if candidate >= len(tf.tokens) {
			last := len(tf.tokens) - 1
			if tf.tokens[last].delim {
				return last - 1
			} else {
				return last
			}
		}

		candidateToken := tf.tokens[candidate]
		if candidateToken.line == focusedToken.line {
			prev := candidate - 1
			for tf.tokens[prev].delim {
				prev--
			}
			return prev
		}

		if candidateToken.end >= anchorIndex {
			lineDiff := int(math.Abs(float64(focusedToken.line) - float64(tf.tokens[candidate].line)))
			if lineDiff != 1 {
				// Edge case when going through empty space at the end of the line
				for {
					lineDiff = int(math.Abs(float64(focusedToken.line) - float64(tf.tokens[candidate].line)))
					if lineDiff == 1 && !tf.tokens[candidate].delim {
						break
					}
					candidate--
				}
			}

			return candidate
		}

		candidate++
	}
}

func (tf *TokenField) renderTokens(focusedToken int) string {
	var netLineLength int = tf.width - 2*tf.horizontalPadding
	var sbTokenField strings.Builder

	line := 0
	index := 0
	renderedIndex := 0
	var sbLinePlain strings.Builder // Tracks plain text for layout decisions
	var sbLine strings.Builder      // Tracks actual rendered output
	for i := 0; i < len(tf.tokens); i += 2 {
		log.Println("========================================")
		log.Println("Word:", tf.tokens[i].word)

		lineWithWord := sbLinePlain.String() + tf.tokens[i].word
		if i+1 < len(tf.tokens) {
			lineWithWord += tf.tokens[i+1].word
		}

		log.Printf("Index %d, lineww: %s\n", index, lineWithWord)

		if len(lineWithWord) > netLineLength {
			log.Printf("%d > %d, resetting line buffer\n", len(lineWithWord), netLineLength)
			sbTokenField.WriteString(sbLine.String())
			sbTokenField.WriteRune('\n')
			sbLine.Reset()
			sbLinePlain.Reset()
			line++
			index = 0
		}

		tf.tokens[i].start = index
		tf.tokens[i].end = tf.tokens[i].start + len(tf.tokens[i].word)
		tf.tokens[i].line = line
		if i+1 < len(tf.tokens) {
			tf.tokens[i+1].line = line
		}
		tf.tokens[i].index = renderedIndex
		log.Println(tf.tokens[i])

		renderedWord := tf.tokens[i].word
		if focusedToken == i {
			renderedWord = defaultStyles().focusedToken.Render(renderedWord)
		}

		sbLine.WriteString(renderedWord)
		sbLinePlain.WriteString(tf.tokens[i].word)
		if i+1 < len(tf.tokens) {
			sbLine.WriteString(tf.tokens[i+1].word)
			sbLinePlain.WriteString(tf.tokens[i+1].word)
		}

		index = tf.tokens[i].end
		if i+1 < len(tf.tokens) {
			tf.tokens[i].end += len(tf.tokens[i+1].word) - 1
			index += len(tf.tokens[i+1].word)
		}
		renderedIndex += 2
		log.Println("========================================")
	}

	if sbLine.Len() > 0 {
		sbTokenField.WriteString(sbLine.String())
	}

	return sbTokenField.String()
}
