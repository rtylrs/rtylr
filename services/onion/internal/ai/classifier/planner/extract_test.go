package planner

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractEntities(t *testing.T) {
	doc := "# New patient intake\n" +
		"Patient: Jane Doe\n" +
		"- Invoice number: INV-2026-0042\n" +
		"* Total: 1,240.50\n" +
		"Phone number: (503) 555-0142\n" +
		"\n" +
		"Empty:\n" +
		"NoColonLine\n" +
		"Nested: Label: value\n" +
		"Phone number: (503) 555-0142\n"
	got := ExtractEntities(doc, 0)
	want := []Entity{
		{Label: "Patient", Value: "Jane Doe"},
		{Label: "Invoice number", Value: "INV-2026-0042"},
		{Label: "Total", Value: "1,240.50"},
		{Label: "Phone number", Value: "(503) 555-0142"},
		{Label: "Nested", Value: "Label: value"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entities, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entity %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestExtractEntitiesLimit(t *testing.T) {
	doc := ""
	for i := 0; i < 200; i++ {
		doc += fmt.Sprintf("Field number %d: value\n", i)
	}
	if got := len(ExtractEntities(doc, 10)); got != 10 {
		t.Fatalf("limit 10: got %d entities", got)
	}
	if got := len(ExtractEntities(doc, 0)); got != ExtractEntityLimit {
		t.Fatalf("default limit: got %d entities, want %d", got, ExtractEntityLimit)
	}
}

func TestExtractEntitiesRejectsJunk(t *testing.T) {
	cases := []string{
		": value only",
		"Label: ",
		"12345: 67890",
		"# comment: value",
	}
	for _, doc := range cases {
		if got := ExtractEntities(doc, 0); len(got) != 0 {
			t.Errorf("ExtractEntities(%q) = %+v, want none", doc, got)
		}
	}
}

func TestExtractEntitiesTruncationKeepsValidUTF8(t *testing.T) {
	label := ""
	for i := 0; i < 60; i++ {
		label += "é"
	}
	got := ExtractEntities(label+": value", 0)
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	if len(got[0].Label) > extractLabelBytes {
		t.Fatalf("label not truncated: %d bytes", len(got[0].Label))
	}
	if !hasLetter(got[0].Label) {
		t.Fatalf("truncated label lost its letters: %q", got[0].Label)
	}
}

func TestExtractEntitiesSkipsURLs(t *testing.T) {
	doc := "Read more at https://docs.example.com/a:b\nLink: //cdn.example.com/x\n" +
		"Notes: see https://example.com/x\nSite: https://example.com\n"
	got := ExtractEntities(doc, 0)
	if len(got) != 1 || got[0].Label != "Site" || got[0].Value != "https://example.com" {
		t.Fatalf("entities = %+v, want only the labeled Site entity", got)
	}
}

func TestNormalizeEntityRejectsURLNoise(t *testing.T) {
	cases := []Entity{
		{Label: "See https", Value: "//example.com/x"},
		{Label: "Notes", Value: "see https://example.com/x"},
		{Label: "Read more at https", Value: "//docs.example.com"},
	}
	for _, c := range cases {
		if entity, ok := NormalizeEntity(c); ok {
			t.Errorf("NormalizeEntity(%+v) = %+v, want rejection", c, entity)
		}
	}
	kept, ok := NormalizeEntity(Entity{Label: "Site", Value: "https://example.com"})
	if !ok || kept.Value != "https://example.com" {
		t.Errorf("clean URL value was rejected: %+v ok=%v", kept, ok)
	}
}

func TestNormalizeEntityCapsAndRejects(t *testing.T) {
	longLabel := strings.Repeat("é", 100)
	longValue := strings.Repeat("v", 400)
	entity, ok := NormalizeEntity(Entity{Label: longLabel, Value: longValue})
	if !ok {
		t.Fatal("long entity was rejected")
	}
	if len(entity.Label) > extractLabelBytes || len(entity.Value) > extractValueBytes {
		t.Fatalf("caps not applied: %d/%d bytes", len(entity.Label), len(entity.Value))
	}
	if _, ok := NormalizeEntity(Entity{Label: "12345", Value: "value"}); ok {
		t.Fatal("letterless label was accepted")
	}
	if _, ok := NormalizeEntity(Entity{Label: "Phone", Value: ""}); ok {
		t.Fatal("empty value was accepted")
	}
}
