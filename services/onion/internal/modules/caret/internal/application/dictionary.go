package application

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var dictionaryEndpoint = "https://api.dictionaryapi.dev/api/v2/entries/en/"

func init() {
	registerGeneralTool("knowledge.define", (*AssistantToolExecutor).defineWord)
}

func (e *AssistantToolExecutor) defineWord(ctx context.Context, input map[string]any) (map[string]any, error) {
	word := strings.ToLower(inputText(input, "word", "term", "query"))
	if word == "" || len([]rune(word)) > 60 {
		return nil, errors.New("give one English word or short phrase")
	}
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
