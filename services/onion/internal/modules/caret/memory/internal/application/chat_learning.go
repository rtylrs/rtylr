package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	learnedThroughKey = "learned_through_id"
	// learnAfterMessages is how many new messages a chat needs before the
	// next pass reads it for durable facts: three exchanges.
	learnAfterMessages     = 6
	chatLearningInputBytes = 16000
	// LearnedMemoryConfidence keeps learned facts above the visibility bar
	// (0.7) but below facts the person stated or approved (1).
	LearnedMemoryConfidence = 0.8
	maxLearnedFactsPerPass  = 5
)

// ChatLearningWork is a stretch of a chat not yet read for durable facts.
type ChatLearningWork struct {
	Transcript string
	Known      []string
	ThroughID  int
}

// PendingChatLearning returns the messages of a chat added since its last
// learning pass, once there are enough of them, with what is already known.
func (s *Service) PendingChatLearning(ctx context.Context, session string, userID int) (ChatLearningWork, bool, error) {
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return ChatLearningWork{}, false, err
	}
	through, _ := sessionMetadata(chat)[learnedThroughKey].(float64)
	messages, err := s.repository.GetContextMessages(ctx, session)
	if err != nil {
		return ChatLearningWork{}, false, err
	}
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].ID < messages[j].ID })
	fresh := messages[:0:0]
	for _, message := range messages {
		if message.ID > int(through) {
			fresh = append(fresh, message)
		}
	}
	if len(fresh) < learnAfterMessages {
		return ChatLearningWork{}, false, nil
	}
	var transcript strings.Builder
	for _, message := range fresh {
		role := "Person"
		if strings.EqualFold(message.Role, "assistant") {
			role = "Caret"
		}
		fmt.Fprintf(&transcript, "%s: %s\n\n", role, clipMiddle(strings.TrimSpace(message.Content), 2500))
	}
	known := []string{}
	if memories, err := s.repository.AssistantMemories(ctx, session); err == nil {
		for _, memory := range memories {
			known = append(known, memory.Value)
		}
	}
	return ChatLearningWork{
		Transcript: validUTF8Suffix(transcript.String(), chatLearningInputBytes),
		Known:      known,
		ThroughID:  fresh[len(fresh)-1].ID,
	}, true, nil
}

// LearnedFact is one durable fact a learning pass found.
type LearnedFact struct {
	// Personal facts follow the person everywhere; business facts stay with
	// this business's chats.
	Personal bool
	Value    string
}

// SaveLearnedFacts stores what a learning pass found and records how far the
// chat has been read, so the same messages aren't read twice.
func (s *Service) SaveLearnedFacts(ctx context.Context, session string, userID int, facts []LearnedFact, throughID int) error {
	for i, fact := range facts {
		if i == maxLearnedFactsPerPass {
			break
		}
		value := strings.Join(strings.Fields(fact.Value), " ")
		if value == "" || len([]rune(value)) > MaxMemoryRunes {
			continue
		}
		kind := "business_fact"
		if fact.Personal {
			kind = "preference"
		}
		sum := sha256.Sum256([]byte(strings.ToLower(value)))
		if _, err := s.StoreUserMemory(ctx, userID, kind, "learned_"+hex.EncodeToString(sum[:8]), value, session, LearnedMemoryConfidence); err != nil {
			return err
		}
	}
	// One key, set atomically: the title and compaction passes update the
	// same metadata concurrently.
	return s.repository.SetChatMetadataNumber(ctx, session, userID, learnedThroughKey, throughID)
}
