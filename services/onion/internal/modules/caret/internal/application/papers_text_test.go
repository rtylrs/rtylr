package application

import (
	"context"
	"testing"
)

func TestSearchPapersRebuildsAbstracts(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		papersEndpoint: `{"results":[{"display_name":"Loyalty works","publication_year":2020,"cited_by_count":42,"doi":"https://doi.org/10.1/x","authorships":[{"author":{"display_name":"A"}},{"author":{"display_name":"B"}}],"abstract_inverted_index":{"Points":[0],"raise":[1],"visits":[2]}}]}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "knowledge.papers", map[string]any{"query": "loyalty programs"})
	if err != nil {
		t.Fatal(err)
	}
	paper := output["papers"].([]map[string]any)[0]
	if paper["abstract"] != "Points raise visits" || paper["authors"] != "A, B" || len(output["sources"].([]map[string]string)) != 1 {
		t.Fatalf("paper = %#v", paper)
	}
}

func TestCountText(t *testing.T) {
	e := &AssistantToolExecutor{}
	output, _, err := e.ExecuteCaretTool(context.Background(), "text.count", map[string]any{"text": "Fresh bread daily. Come early!\nOpen 7am", "find": "bread"})
	if err != nil {
		t.Fatal(err)
	}
	if output["words"] != 7 || output["sentences"] != 3 || output["lines"] != 2 || output["characters"] != 39 || output["occurrences_of_bread"] != 1 {
		t.Fatalf("output = %#v", output)
	}
}
