package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var sample string = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

var sample2 string = "asdfsadf sadfsadf. asdfsadf. asdfsadfsdf? asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?asdfsadf sadfsadf. asdfsadf. asdfsadfsdf? asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?asdfsadf sadfsadf. asdfsadf. asdfsadfsdf?"

var sample3 string = "En un lugar de la Mancha, de cuyo nombre no quiero acordarme, no ha mucho tiempo que vivía un hidalgo de los de lanza en astillero, adarga antigua, rocín flaco y galgo corredor. Una olla de algo más vaca que carnero, salpicón las más noches, duelos y quebrantos los sábados, lantejas los viernes, algún palomino de añadidura los domingos, consumían las tres partes de su hacienda."

type Translator int

const (
	LibreTranslate Translator = iota
	DeepL
)

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

func (tf *TokenField) renderTokens(focusedToken int, multiselect bool, multistart int) string {
	var netLineLength int = tf.width - 2*tf.horizontalPadding
	var sbTokenField strings.Builder

	multiend := focusedToken
	if multiselect && focusedToken < multistart {
		tmp := multistart
		multistart = focusedToken
		multiend = tmp
	}

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
		if multiselect && multistart <= i && i <= multiend {
			renderedWord = defaultStyles().multiSelectToken.Render(renderedWord)
		} else if !multiselect && focusedToken == i {
			renderedWord = defaultStyles().focusedToken.Render(renderedWord)
		} else {
			renderedWord = defaultStyles().normalToken.Render(renderedWord)
		}

		sbLine.WriteString(renderedWord)
		sbLinePlain.WriteString(tf.tokens[i].word)
		if i+1 < len(tf.tokens) {
			delim := tf.tokens[i+1].word
			if multiselect && multistart <= i && i < multiend {
				sbLine.WriteString(defaultStyles().multiSelectToken.Render(delim))
			} else {
				sbLine.WriteString(defaultStyles().normalToken.Render(delim))
			}
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

func (tf *TokenField) getSentence(selected int) string {
	index := -1
	for i, token := range tf.tokens {
		if token.index == selected {
			index = i
			break
		}
	}
	start := index
	for start >= 0 && !strings.ContainsRune(".?!", rune(tf.tokens[start].word[0])) {
		start--
	}
	start++

	end := index
	for end <= len(tf.tokens)-1 && !strings.ContainsRune(".?!", rune(tf.tokens[end].word[0])) {
		end++
	}
	if end <= len(tf.tokens)-1 {
		end++
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		renderedToken := tf.tokens[i].word
		if i == index {
			renderedToken = defaultStyles().focusedToken.Render(renderedToken)
		}
		sb.WriteString(renderedToken)
	}

	return sb.String()
}

func renderTranslations(text string) string {
	var sb strings.Builder
	for _, translator := range []Translator{DeepL, LibreTranslate} {

		sb.WriteRune('\n')
		translatorStyle := defaultStyles().modalTranslator
		switch translator {
		case LibreTranslate:
			sb.WriteString(translatorStyle.Render("LibreTranslate"))
		case DeepL:
			sb.WriteString(translatorStyle.Render("DeepL"))
		default:
			sb.WriteString(translatorStyle.Render("unrecognized translator"))
		}
		sb.WriteString(":\n\n")

		translations, err := getTranslations(text, translator)
		if err != nil {
			sb.WriteString(err.Error())
			sb.WriteRune('\n')
		} else {
			sb.WriteString(strings.Join(translations, ", "))
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

func getTranslations(text string, translator Translator) ([]string, error) {
	switch translator {
	case LibreTranslate:
		return getTranslationsLibreTranslate(text)
	case DeepL:
		return getTranslationsDeepL(text)
	default:
		return nil, errors.New("unrecognized translator")
	}
}

func getTranslationsLibreTranslate(text string) ([]string, error) {
	const translateUrl = "http://127.0.0.1:5000/translate"

	type reqObj struct {
		Query        string `json:"q"`
		Source       string `json:"source"`
		Target       string `json:"target"`
		Alternatives int    `json:"alternatives"`
		Format       string `json:"format"`
	}
	payload := reqObj{
		Query:        text,
		Source:       "es",
		Target:       "en",
		Format:       "text",
		Alternatives: 4,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, translateUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(errBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respObj struct {
		Alternatives   []string `json:"alternatives"`
		TranslatedText string   `json:"translatedText"`
	}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return nil, err
	}

	translations := []string{respObj.TranslatedText}
	for _, alt := range respObj.Alternatives {
		translations = append(translations, alt)
	}

	return translations, nil
}

func getTranslationsDeepL(text string) ([]string, error) {
	const translateUrl = "https://api-free.deepl.com/v2/translate"

	type reqObj struct {
		Text   []string `json:"text"`
		Source string   `json:"source_lang"`
		Target string   `json:"target_lang"`
	}
	payload := reqObj{
		Text:   []string{text},
		Source: "ES",
		Target: "EN",
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, translateUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+os.Getenv("API_KEY"))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(errBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respObj struct {
		Translations []struct {
			Text   string `json:"text"`
			Source string `json:"detected_source_language"`
		} `json:"translations"`
	}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return nil, err
	}

	translations := []string{}
	for _, translation := range respObj.Translations {
		translations = append(translations, translation.Text)
	}

	return translations, nil
}

func (tf *TokenField) getWordSelection(selected int, multiselect bool, multistart int) string {
	if !multiselect {
		return tf.tokens[selected].word
	}

	multiend := selected
	if multiselect && selected < multistart {
		tmp := multistart
		multistart = selected
		multiend = tmp
	}

	var sb strings.Builder
	for i := multistart; i <= multiend; i++ {
		sb.WriteString(tf.tokens[i].word)
	}

	return strings.Trim(sb.String(), " ")
}
