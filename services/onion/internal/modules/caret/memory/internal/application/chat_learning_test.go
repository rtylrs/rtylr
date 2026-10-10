package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type learningRepository struct {
	*chatRepository
	stored  []domain.UserMemory
	through int
}

func (r *learningRepository) StoreUserMemory(_ context.Context, memory *domain.UserMemory) (*domain.UserMemory, error) {
	r.stored = append(r.stored, *memory)
	return memory, nil
}

func (r *learningRepository) SetChatMetadataNumber(_ context.Context, _ string, _ int, key string, value int) error {
	if key == learnedThroughKey {
		r.through = value
	}
	return nil
}

func TestChatLearningWaitsForTwoExchangesAndRecordsProgress(t *testing.T) {
	repo := &learningRepository{chatRepository: chatWithTurns(1, 5)}
	service := NewService(repo)
	if _, ok, err := service.PendingChatLearning(context.Background(), "s", 1); err != nil || ok {
		t.Fatalf("one exchange and a request must not trigger learning: ok=%v err=%v", ok, err)
	}
	repo.chatRepository = chatWithTurns(3, 5)
	work, ok, err := service.PendingChatLearning(context.Background(), "s", 1)
	if err != nil || !ok || work.ThroughID != 7 || !strings.Contains(work.Transcript, "Person: message-001") || !strings.Contains(work.Transcript, "Caret: message-002") {
		t.Fatalf("work = %#v ok=%v err=%v", work, ok, err)
	}
	if len(work.Known) != 1 || work.Known[0] != "Prefers short answers" {
		t.Fatalf("known = %q", work.Known)
	}
	facts := []LearnedFact{{Personal: true, Value: "Prefers Arabic for customer messages"}, {Value: "The Hamra branch opens at 7am"}}
	if err := service.SaveLearnedFacts(context.Background(), "s", 1, facts, work.ThroughID); err != nil {
		t.Fatal(err)
	}
	if len(repo.stored) != 2 || repo.stored[0].MemoryType != "preference" || repo.stored[1].MemoryType != "business_fact" || repo.stored[1].SourceSession != "s" || repo.stored[1].Confidence != LearnedMemoryConfidence || repo.through != 7 {
		t.Fatalf("stored = %#v through = %d", repo.stored, repo.through)
	}
}
