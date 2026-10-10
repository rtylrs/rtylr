package application

import (
	"context"
	"testing"
)

func TestDefineWordKeepsDefinitionsAndExamples(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
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

func TestCountryFactsAcceptsCodesAndPrefersExactNames(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		countriesEndpoint + "alpha/LB":    `{"name":{"common":"Lebanon","official":"Lebanese Republic"},"cca2":"LB","capital":["Beirut"],"currencies":{"LBP":{"name":"Lebanese pound","symbol":"ل.ل"}},"languages":{"ara":"Arabic","fra":"French"},"population":5490000,"timezones":["UTC+02:00"],"idd":{"root":"+9","suffixes":["61"]}}`,
		countriesEndpoint + "name/Guinea": `[{"name":{"common":"Papua New Guinea"},"cca2":"PG"},{"name":{"common":"Guinea"},"cca2":"GN"}]`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "knowledge.country", map[string]any{"country": "LB"})
	if err != nil {
		t.Fatal(err)
	}
	if output["capital"] != "Beirut" || output["dialing_code"] != "+961" || output["currencies"].([]string)[0] != "LBP (Lebanese pound, ل.ل)" {
		t.Fatalf("output = %#v", output)
	}
	output, _, err = executor.ExecuteCaretTool(context.Background(), "knowledge.country", map[string]any{"country": "Guinea"})
	if err != nil || output["country_code"] != "GN" {
		t.Fatalf("exact match = %#v, %v", output, err)
	}
}
