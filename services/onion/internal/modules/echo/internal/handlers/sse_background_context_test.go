package handlers

import (
	"context"
	"testing"
)

type detachedMemoryFake struct {
	storeContextErr     error
	summarizeContextErr error
}

func (m *detachedMemoryFake) CreateSession(context.Context, int, string) (string, error) {
	return "", nil
}

func (m *detachedMemoryFake) StoreMessage(ctx context.Context, _ string, _ int, _, _ string, _ int) (any, error) {
	m.storeContextErr = ctx.Err()
	return nil, nil
}

func (m *detachedMemoryFake) SummarizeSession(ctx context.Context, _ string, _ int) error {
	m.summarizeContextErr = ctx.Err()
	return nil
}

func (m *detachedMemoryFake) BuildEnhancedPrompt(context.Context, string, int, string) string {
	return ""
}

func TestStoreAIResponseDetachedSurvivesRequestCancellation(t *testing.T) {
	requestCtx, cancel := context.WithCancel(context.Background())
	cancel()
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{MemoryService: memory}

	handler.storeAIResponseDetached(context.WithoutCancel(requestCtx), "session-1", 7, "answer")

	if memory.storeContextErr != nil {
		t.Fatalf("store context was canceled: %v", memory.storeContextErr)
	}
	if memory.summarizeContextErr != nil {
		t.Fatalf("summarize context was canceled: %v", memory.summarizeContextErr)
	}
}
