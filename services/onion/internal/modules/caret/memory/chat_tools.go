package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// listMemoriesTool shows Caret what it remembers, so it can check before
// proposing a duplicate and answer "what do you know about me".
func (m *Module) listMemoriesTool(ctx context.Context) (map[string]any, bool, error) {
	memories, err := m.memoryService.ListMemories(ctx)
	if err != nil {
		return nil, true, err
	}
	out := make([]map[string]any, 0, len(memories))
	for _, memory := range memories {
		out = append(out, map[string]any{
			"id": memory.ID, "type": memory.MemoryType, "value": memory.Value,
			"updated": memory.UpdatedAt.Format(time.DateOnly),
		})
	}
	return map[string]any{"memories": out, "total": len(out)}, true, nil
}

func (m *Module) forgetMemoryTool(ctx context.Context, input map[string]any) (map[string]any, bool, error) {
	id := toolInt(input, "id", "memory_id")
	if id <= 0 {
		return nil, true, errors.New("give the id of the memory to forget; memory.list shows ids")
	}
	if err := m.memoryService.DeleteMemory(ctx, id); err != nil {
		return nil, true, err
	}
	return map[string]any{"forgotten_id": id, "message": "Caret will no longer use that memory."}, true, nil
}

func (m *Module) searchChatsTool(ctx context.Context, userID int, input map[string]any) (map[string]any, bool, error) {
	query := toolString(input, "query", "q", "topic")
	hits, err := m.memoryService.SearchChats(ctx, userID, query, toolInt(input, "limit"))
	if err != nil {
		return nil, true, err
	}
	out := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		out = append(out, map[string]any{
			"chat_id": hit.SessionID, "chat_title": hit.Title, "from": hit.Role,
			"date": hit.CreatedAt.Format(time.DateOnly), "excerpt": hit.Snippet,
		})
	}
	note := "Use chats.read with a chat_id to see more of that conversation."
	if len(out) == 0 {
		note = "No past chat mentions that. Try other words."
	}
	return map[string]any{"query": query, "matches": out, "note": note}, true, nil
}

func (m *Module) readChatTool(ctx context.Context, userID int, input map[string]any) (map[string]any, bool, error) {
	sessionID := toolString(input, "chat_id", "session_id", "id")
	if sessionID == "" {
		return nil, true, errors.New("give the chat_id from chats.search")
	}
	session, messages, err := m.memoryService.ReadChat(ctx, sessionID, userID, toolInt(input, "limit"))
	if err != nil {
		return nil, true, errors.New("that chat was not found")
	}
	out := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		content := []rune(strings.TrimSpace(message.Content))
		if len(content) > 1500 {
			content = append(content[:1500], []rune(" [...]")...)
		}
		out = append(out, map[string]any{"from": message.Role, "date": message.CreatedAt.Format(time.DateOnly), "text": string(content)})
	}
	return map[string]any{"chat_id": session.SessionID, "title": session.Title, "messages": out}, true, nil
}

func toolString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := input[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func toolInt(input map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := input[key].(type) {
		case float64:
			return int(value)
		case int:
			return value
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				return parsed
			}
		case fmt.Stringer:
			if parsed, err := strconv.Atoi(value.String()); err == nil {
				return parsed
			}
		}
	}
	return 0
}
