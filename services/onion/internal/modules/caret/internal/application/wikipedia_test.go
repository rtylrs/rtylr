package application

import (
	"context"
	"testing"
)

func TestWikipediaSearchesThenSummarises(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		"https://fr.wikipedia.org/w/rest.php/v1/search/page":         `{"pages":[{"key":"Beyrouth","title":"Beyrouth"},{"key":"Beyrouth_(film)","title":"Beyrouth (film)","description":"film"}]}`,
		"https://fr.wikipedia.org/api/rest_v1/page/summary/Beyrouth": `{"title":"Beyrouth","description":"capitale du Liban","extract":"Beyrouth est la capitale du Liban.","type":"standard","content_urls":{"desktop":{"page":"https://fr.wikipedia.org/wiki/Beyrouth"}}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "knowledge.wikipedia", map[string]any{"query": "Beyrouth", "language": "fr"})
	if err != nil {
		t.Fatal(err)
	}
	if output["summary"] != "Beyrouth est la capitale du Liban." || len(output["other_matches"].([]map[string]string)) != 1 {
		t.Fatalf("output = %#v", output)
	}
	if sources := output["sources"].([]map[string]string); sources[0]["url"] != "https://fr.wikipedia.org/wiki/Beyrouth" {
		t.Fatalf("sources = %#v", sources)
	}
}
