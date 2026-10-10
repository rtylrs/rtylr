package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// chatRepository implements only what the context window and compaction use.
type chatRepository struct {
	repository
	chat     domain.ChatSession
	messages []domain.ConversationMessage
	updates  int
}

func (r *chatRepository) GetChatSession(context.Context, string, int) (*domain.ChatSession, error) {
	chat := r.chat
	return &chat, nil
}

func (r *chatRepository) GetContextMessages(context.Context, string) ([]domain.ConversationMessage, error) {
	return r.messages, nil
}

func (r *chatRepository) AssistantProfile(context.Context) (string, error) {
	return "Signed-in user: Sam.", nil
}

func (r *chatRepository) AssistantMemories(context.Context, string) ([]domain.UserMemory, error) {
	return []domain.UserMemory{{Value: "Prefers short answers"}}, nil
}

func (r *chatRepository) UpdateChatSession(_ context.Context, _ string, _ int, title, summary, metadata string) (*domain.ChatSession, error) {
	r.updates++
	r.chat.Title, r.chat.Summary, r.chat.Metadata = title, summary, metadata
	chat := r.chat
	return &chat, nil
}

func chatWithTurns(turns int, size int) *chatRepository {
	repo := &chatRepository{chat: domain.ChatSession{SessionID: "s", Title: "Chat", Metadata: "{}"}}
	start := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
	for i := 1; i <= turns*2; i++ {
		role := "user"
		if i%2 == 0 {
			role = "assistant"
		}
		repo.messages = append(repo.messages, domain.ConversationMessage{
			ID:        i,
			Role:      role,
			Content:   fmt.Sprintf("message-%03d %s", i, strings.Repeat("x", size)),
			CreatedAt: start.Add(time.Duration(i) * time.Minute),
		})
	}
	// The request being answered is stored before the context is built.
	repo.messages = append(repo.messages, domain.ConversationMessage{ID: turns*2 + 1, Role: "user", Content: "current request"})
	return repo
}

func TestAssistantContextKeepsNewestTurnsWithinBudget(t *testing.T) {
	repo := chatWithTurns(40, 2000)
	got, err := NewService(repo).AssistantContext(context.Background(), "s", 1, "")
	if err != nil {
		t.Fatalf("AssistantContext: %v", err)
	}

	if !strings.Contains(got, "Signed-in user: Sam.") || !strings.Contains(got, "- Prefers short answers") {
		t.Fatalf("profile or preferences missing:\n%.300s", got)
	}
	if strings.Contains(got, "current request") {
		t.Fatal("current request should be supplied separately, not in history")
	}
	if !strings.Contains(got, "message-080") || !strings.Contains(got, "message-079") {
		t.Fatal("newest turns must always be present")
	}
	if strings.Contains(got, "message-001 ") {
		t.Fatal("oldest turn should fall outside the window")
	}
	if !strings.Contains(got, "older message(s) are not shown yet") {
		t.Fatal("omitted history should be called out")
	}
	if len(got) > transcriptBudgetBytes+profileBudgetBytes+memoriesBudgetBytes+1000 {
		t.Fatalf("context is %d bytes, over budget", len(got))
	}
	if strings.Index(got, "message-079") > strings.Index(got, "message-080") {
		t.Fatal("transcript must read oldest first")
	}
}

func TestAssistantContextUsesSummaryForCompactedTurns(t *testing.T) {
	repo := chatWithTurns(3, 10)
	repo.chat.Summary = "User is planning a launch for Friday."
	repo.chat.Metadata = `{"compacted_through_id":2,"project_id":""}`

	got, err := NewService(repo).AssistantContext(context.Background(), "s", 1, "")
	if err != nil {
		t.Fatalf("AssistantContext: %v", err)
	}
	if !strings.Contains(got, "User is planning a launch for Friday.") {
		t.Fatal("summary missing")
	}
	if strings.Contains(got, "message-001") || strings.Contains(got, "message-002") {
		t.Fatal("compacted turns should be represented by the summary only")
	}
	if !strings.Contains(got, "message-003") {
		t.Fatal("turns after the summary must stay verbatim")
	}
}

func TestPendingCompactionWaitsUntilHistoryOverflows(t *testing.T) {
	service := NewService(chatWithTurns(3, 100))
	if _, ok, err := service.PendingCompaction(context.Background(), "s", 1); err != nil || ok {
		t.Fatalf("short chat should not compact: ok=%v err=%v", ok, err)
	}
}

func TestCompactionFoldsOldTurnsAndKeepsRecentOnes(t *testing.T) {
	repo := chatWithTurns(40, 2000)
	service := NewService(repo)

	work, ok, err := service.PendingCompaction(context.Background(), "s", 1)
	if err != nil || !ok {
		t.Fatalf("long chat should compact: ok=%v err=%v", ok, err)
	}
	if work.PreviousThrough != 0 || !strings.Contains(work.Transcript, "message-001") {
		t.Fatalf("first compaction should start at the oldest turn: %+v", work.PreviousThrough)
	}
	if strings.Contains(work.Transcript, "message-080") {
		t.Fatal("newest turns must stay verbatim, not be compacted")
	}

	if err := service.SaveCompaction(context.Background(), "s", 1, work, "Summary one."); err != nil {
		t.Fatalf("SaveCompaction: %v", err)
	}
	if repo.chat.Summary != "Summary one." || !strings.Contains(repo.chat.Metadata, fmt.Sprintf(`"compacted_through_id":%d`, work.NewThrough)) {
		t.Fatalf("compaction not stored: %q %s", repo.chat.Summary, repo.chat.Metadata)
	}

	// A stale compaction (based on an older summary) must not overwrite it.
	if err := service.SaveCompaction(context.Background(), "s", 1, work, "Stale summary."); err != nil {
		t.Fatalf("stale SaveCompaction: %v", err)
	}
	if repo.chat.Summary != "Summary one." {
		t.Fatal("stale compaction overwrote a newer summary")
	}

	if _, ok, _ := service.PendingCompaction(context.Background(), "s", 1); ok {
		t.Fatal("history that fits after compaction should not compact again right away")
	}
}

func TestRenderTurnNotesToolStepsAndClipsLongMessages(t *testing.T) {
	turn := renderTurn(domain.ConversationMessage{
		Role:     "assistant",
		Content:  "start " + strings.Repeat("é", messageClipBytes) + " end",
		Metadata: `{"blocks":[{"type":"tool_call","name":"web.search","status":"success","input":{"query":"plague news"}}]}`,
	})
	if !strings.Contains(turn, `[searched the web for "plague news"]`) {
		t.Fatalf("missing tool note:\n%.200s", turn)
	}
	if !strings.HasPrefix(turn, "Caret:") || !strings.Contains(turn, "start ") || !strings.HasSuffix(turn, " end") {
		t.Fatal("long message should keep its start and end")
	}
	if !strings.Contains(turn, "middle of message omitted") || len(turn) > messageClipBytes+200 {
		t.Fatalf("long message not clipped: %d bytes", len(turn))
	}
	if !strings.Contains(turn, "é") || strings.ContainsRune(turn, '�') {
		t.Fatal("clipping must not split a multi-byte character")
	}
}

type lessonRepository struct{ *chatRepository }

func (r lessonRepository) AssistantMemories(context.Context, string) ([]domain.UserMemory, error) {
	return []domain.UserMemory{
		{MemoryType: "preference", Value: "Currency is USD"},
		{MemoryType: LessonMemoryType, Value: "Lead with the number, then one line of context."},
	}, nil
}

func TestAssistantContextGivesLessonsTheirOwnSection(t *testing.T) {
	got, err := NewService(lessonRepository{chatWithTurns(1, 10)}).AssistantContext(context.Background(), "s", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	lessons := strings.Index(got, "Lessons from this person's feedback")
	saved := strings.Index(got, "Saved user preferences:\n- Currency is USD")
	if lessons < 0 || saved < 0 || !strings.Contains(got[lessons:saved], "Lead with the number") {
		t.Fatalf("context = %s", got)
	}
}
