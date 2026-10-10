package application

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var (
	dictionaryEndpoint = "https://api.dictionaryapi.dev/api/v2/entries/en/"
	wiktionaryEndpoint = "https://en.wiktionary.org/api/rest_v1/page/definition/"
)

// defineWord defines a word from Wiktionary, which covers many languages,
// falling back to an English dictionary that also lists synonyms.
func (e *AssistantToolExecutor) defineWord(ctx context.Context, input map[string]any) (map[string]any, error) {
	word := strings.TrimSpace(inputText(input, "word", "term", "query"))
	if word == "" || len([]rune(word)) > 60 {
		return nil, errors.New("give one word or short phrase")
	}
	language := strings.ToLower(inputText(input, "language", "lang"))
	if !wikipediaLanguage.MatchString(language) {
		language = "en"
	}
	if out, err := e.wiktionaryDefinition(ctx, word, language); err == nil {
		return out, nil
	} else if language != "en" {
		return nil, err
	}
	return e.englishDictionary(ctx, strings.ToLower(word))
}

func (e *AssistantToolExecutor) wiktionaryDefinition(ctx context.Context, word, language string) (map[string]any, error) {
	var entries map[string][]struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Language     string `json:"language"`
		Definitions  []struct {
			Definition string   `json:"definition"`
			Examples   []string `json:"examples"`
		} `json:"definitions"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), wiktionaryEndpoint+url.PathEscape(word), &entries); err != nil {
		return nil, err
	}
	meanings := []map[string]any{}
	for _, entry := range entries[language] {
		definitions := []map[string]string{}
		for _, definition := range entry.Definitions {
			text := strings.Join(strings.Fields(plainSearchText(definition.Definition)), " ")
			if text == "" || len(definitions) == 3 {
				continue
			}
			item := map[string]string{"definition": text}
			if len(definition.Examples) > 0 {
				item["example"] = strings.Join(strings.Fields(plainSearchText(definition.Examples[0])), " ")
			}
			definitions = append(definitions, item)
		}
		if len(definitions) > 0 {
			meanings = append(meanings, map[string]any{"part_of_speech": strings.ToLower(entry.PartOfSpeech), "definitions": definitions})
		}
	}
	if len(meanings) == 0 {
		return nil, errors.New("the dictionary has no entry for that word in that language")
	}
	link := "https://en.wiktionary.org/wiki/" + url.PathEscape(word)
	return map[string]any{
		"word": word, "language": language, "meanings": meanings,
		"sources": []map[string]string{{"title": word + " - Wiktionary", "url": link}},
	}, nil
}

func (e *AssistantToolExecutor) englishDictionary(ctx context.Context, word string) (map[string]any, error) {
	var entries []struct {
		Word     string `json:"word"`
		Phonetic string `json:"phonetic"`
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
				Example    string `json:"example"`
			} `json:"definitions"`
			Synonyms []string `json:"synonyms"`
			Antonyms []string `json:"antonyms"`
		} `json:"meanings"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), dictionaryEndpoint+url.PathEscape(word), &entries); err != nil {
		if errors.Is(err, errNotFound) {
			return nil, errors.New("the English dictionary has no entry for that word")
		}
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("the English dictionary has no entry for that word")
	}
	meanings := []map[string]any{}
	for _, entry := range entries {
		for _, meaning := range entry.Meanings {
			definitions := []map[string]string{}
			for i, definition := range meaning.Definitions {
				if i == 3 {
					break
				}
				item := map[string]string{"definition": definition.Definition}
				if definition.Example != "" {
					item["example"] = definition.Example
				}
				definitions = append(definitions, item)
			}
			item := map[string]any{"part_of_speech": meaning.PartOfSpeech, "definitions": definitions}
			if len(meaning.Synonyms) > 0 {
				item["synonyms"] = meaning.Synonyms[:min(len(meaning.Synonyms), 6)]
			}
			if len(meaning.Antonyms) > 0 {
				item["antonyms"] = meaning.Antonyms[:min(len(meaning.Antonyms), 6)]
			}
			meanings = append(meanings, item)
		}
	}
	return map[string]any{"word": entries[0].Word, "phonetic": entries[0].Phonetic, "meanings": meanings}, nil
}
