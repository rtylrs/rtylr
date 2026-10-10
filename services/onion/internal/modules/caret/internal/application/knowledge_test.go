package application

import (
	"context"
	"testing"
)

func TestDefineWordKeepsDefinitionsAndExamples(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		wiktionaryEndpoint + "Terse": `{}`,
		dictionaryEndpoint + "terse": `[{"word":"terse","phonetic":"/tɜːs/","meanings":[{"partOfSpeech":"adjective","definitions":[{"definition":"Brief.","example":"a terse reply"}],"synonyms":["curt"]}]}]`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "knowledge.define", map[string]any{"word": "Terse"})
	if err != nil {
		t.Fatal(err)
	}
	meaning := output["meanings"].([]map[string]any)[0]
	if meaning["part_of_speech"] != "adjective" || meaning["definitions"].([]map[string]string)[0]["example"] != "a terse reply" {
		t.Fatalf("meaning = %#v", meaning)
	}
}

func TestDefineWordUsesWiktionaryInOtherLanguages(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		wiktionaryEndpoint + "maison": `{"fr":[{"partOfSpeech":"Noun","definitions":[{"definition":"<a href=\"/wiki/house\">House</a>, home.","examples":["<i>une maison</i>"]}]}]}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "knowledge.define", map[string]any{"word": "maison", "language": "fr"})
	if err != nil {
		t.Fatal(err)
	}
	definition := output["meanings"].([]map[string]any)[0]["definitions"].([]map[string]string)[0]
	if definition["definition"] != "House, home." || definition["example"] != "une maison" {
		t.Fatalf("definition = %#v", definition)
	}
}

func TestCountryFactsFromTheEmbeddedTable(t *testing.T) {
	e := &AssistantToolExecutor{}
	for _, query := range []string{"LB", "lebanon", "LBN", "Lebanese Republic"} {
		output, _, err := e.ExecuteCaretTool(context.Background(), "knowledge.country", map[string]any{"country": query})
		if err != nil {
			t.Fatalf("%q: %v", query, err)
		}
		if output["capital"] != "Beirut" || output["dialing_code"] != "+961" || output["country_code"] != "LB" {
			t.Fatalf("%q = %#v", query, output)
		}
	}
	output, _, err := e.ExecuteCaretTool(context.Background(), "knowledge.country", map[string]any{"country": "Guinea"})
	if err != nil || output["country_code"] != "GN" {
		t.Fatalf("exact name must win over partial matches: %#v %v", output, err)
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "knowledge.country", map[string]any{"country": "Atlantis"}); err == nil {
		t.Fatal("unknown country accepted")
	}
}
