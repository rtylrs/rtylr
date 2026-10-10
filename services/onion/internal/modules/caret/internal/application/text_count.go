package application

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

// countText counts exactly what models miscount: characters, words,
// sentences, lines, and whether text fits common length limits.
func (e *AssistantToolExecutor) countText(_ context.Context, input map[string]any) (map[string]any, error) {
	text := inputText(input, "text", "content", "value")
	if text == "" {
		return nil, errors.New("give the text to count")
	}
	words := len(strings.FieldsFunc(text, func(r rune) bool { return unicode.IsSpace(r) }))
	sentences := 0
	inSentence := false
	for _, r := range text {
		switch {
		case r == '.' || r == '!' || r == '?' || r == '؟' || r == '。':
			if inSentence {
				sentences++
			}
			inSentence = false
		case !unicode.IsSpace(r):
			inSentence = true
		}
	}
	if inSentence {
		sentences++
	}
	characters := utf8.RuneCountInString(text)
	out := map[string]any{
		"characters": characters, "characters_without_spaces": utf8.RuneCountInString(strings.Join(strings.Fields(text), "")),
		"words": words, "sentences": sentences, "lines": strings.Count(text, "\n") + 1,
		"reading_time_minutes": math.Max(1, math.Round(float64(words)/230)),
		"fits":                 map[string]bool{"sms_160": characters <= 160, "x_post_280": characters <= 280, "meta_description_160": characters <= 160, "instagram_caption_2200": characters <= 2200},
	}
	if word := inputText(input, "find", "word", "term"); word != "" {
		out["occurrences_of_"+strings.ToLower(word)] = strings.Count(strings.ToLower(text), strings.ToLower(word))
	}
	return out, nil
}
