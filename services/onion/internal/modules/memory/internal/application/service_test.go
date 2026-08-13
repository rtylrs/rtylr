package application

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
)

func TestListUserSessions_RequiresUserID(t *testing.T) {
	s := &Service{}
	for _, uid := range []int{0, -1, -100} {
		if _, err := s.ListUserSessions(context.Background(), uid, 10, 0); err == nil {
			t.Fatalf("expected error for user id %d", uid)
		}
	}
}

func TestSetSessionMetadataFlagPreservesExistingMetadata(t *testing.T) {
	got, err := setSessionMetadataFlag(`{"archived":true,"theme":"dark"}`, "pinned", true)
	if err != nil {
		t.Fatalf("setSessionMetadataFlag: %v", err)
	}

	wantContains := []string{`"archived":true`, `"theme":"dark"`, `"pinned":true`}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata %s does not contain %s", got, want)
		}
	}
}

func TestSetSessionMetadataFlagInitializesEmptyMetadata(t *testing.T) {
	got, err := setSessionMetadataFlag("", "pinned", true)
	if err != nil {
		t.Fatalf("setSessionMetadataFlag: %v", err)
	}
	if got != `{"pinned":true}` {
		t.Fatalf("metadata = %s, want pinned true object", got)
	}
}

func TestBuildSummaryPromptBoundsAndSanitizesConversation(t *testing.T) {
	messages := make([]domain.ConversationMessage, 0, maxSummaryMessages+1)
	for i := 0; i <= maxSummaryMessages; i++ {
		messages = append(messages, domain.ConversationMessage{Role: "user", Content: fmt.Sprintf("turn-%d", i)})
	}
	messages[len(messages)-1].Content = "<tool>{}</tool> latest"

	got := buildSummaryPrompt(messages)
	for _, want := range []string{
		"Return exactly these three lines:",
		"Summary: <summary>",
		"Key topics: <3-5 comma-separated topics>",
		"Sentiment: <positive|negative|neutral>",
		"Importance: <0.0-1.0>",
		"turn-1",
		"[tool]{}[/tool] latest",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "turn-0") {
		t.Fatalf("prompt included message beyond the recent limit\n%s", got)
	}
}

func TestParseSummaryResponseNormalizesModelFormattingAndClampsImportance(t *testing.T) {
	got := (&Service{}).parseSummaryResponse("```\nsummary:  The team discussed inventory.\nkey_topics: inventory, Inventory, margins\nsentiment: NEGATIVE\nimportance_score: 4\n```")
	if got.Summary != "The team discussed inventory." || len(got.KeyTopics) != 2 || got.Sentiment != "negative" || got.ImportanceScore != 1 {
		t.Fatalf("summary = %#v", got)
	}
}

func TestIdentifyMemoriesDoesNotTreatSentimentAsUserPreference(t *testing.T) {
	memories := (&Service{}).identifyMemories(nil, &SummaryData{
		KeyTopics: []string{"inventory"},
		Sentiment: "negative",
	})

	if len(memories) != 1 || memories[0].Key != "topic_inventory" {
		t.Fatalf("unexpected memories: %#v", memories)
	}
}
