package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	rememberToolName      = "memory.remember"
	defaultMemoryType     = "business_fact"
	maxRememberKeyRunes   = 80
	maxRememberValueRunes = 1000
)

type rememberToolInput struct {
	MemoryType    string
	Key           string
	Value         string
	SourceSession string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != rememberToolName {
		return nil, false, nil
	}
	if m == nil || m.memoryService == nil {
		return nil, true, errors.New("memory service is not available")
	}

	userID := authenticatedUserID(ctx)
	if userID <= 0 {
		return nil, true, errors.New("sign in is required before Caret can remember this")
	}

	parsed, err := parseRememberToolInput(input)
	if err != nil {
		return nil, true, err
	}

	stored, err := m.memoryService.StoreUserMemory(ctx, int(userID), parsed.MemoryType, parsed.Key, parsed.Value, parsed.SourceSession, 1)
	if err != nil {
		return nil, true, err
	}

	return map[string]any{
		"memory": map[string]any{
			"id":          stored.ID,
			"type":        stored.MemoryType,
			"key":         stored.Key,
			"value":       stored.Value,
			"updated_at":  stored.UpdatedAt,
			"confidence":  stored.Confidence,
			"source_chat": stored.SourceSession,
		},
		"message": "Saved for future Caret conversations.",
	}, true, nil
}

func authenticatedUserID(ctx context.Context) int64 {
	if user, ok := authPkg.GetAuthUser(ctx); ok && user != nil && user.UserID > 0 {
		return user.UserID
	}
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.UserID > 0 {
		return device.User.UserID
	}
	return 0
}

func parseRememberToolInput(input map[string]any) (rememberToolInput, error) {
	flat := flattenRememberInput(input)
	value := firstRememberString(flat, "value", "fact", "preference", "rule", "note", "content", "remember")
	value = sanitizeRememberText(value, maxRememberValueRunes)
	if value == "" {
		return rememberToolInput{}, errors.New("tell Caret exactly what to remember")
	}

	memoryType := normalizeRememberType(firstRememberString(flat, "memory_type", "memoryType", "type", "category", "kind"))
	key := normalizeRememberKey(firstRememberString(flat, "key", "name", "label", "slug"))
	if key == "" {
		key = deriveRememberKey(memoryType, value)
	}
	sourceSession := sanitizeRememberText(firstRememberString(flat, "source_session", "sourceSession", "session_id", "sessionId"), maxRememberKeyRunes)

	return rememberToolInput{
		MemoryType:    memoryType,
		Key:           key,
		Value:         value,
		SourceSession: sourceSession,
	}, nil
}

func flattenRememberInput(input map[string]any) map[string]any {
	flat := map[string]any{}
	for key, value := range input {
		flat[key] = value
	}
	for _, nestedKey := range []string{"memory", "item", "data"} {
		nested, ok := input[nestedKey].(map[string]any)
		if !ok {
			continue
		}
		for key, value := range nested {
			if _, exists := flat[key]; !exists {
				flat[key] = value
			}
		}
	}
	return flat
}

func firstRememberString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return v
			}
		case fmt.Stringer:
			if strings.TrimSpace(v.String()) != "" {
				return v.String()
			}
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				part := strings.TrimSpace(fmt.Sprint(item))
				if part != "" {
					parts = append(parts, part)
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, "; ")
			}
		default:
			text := strings.TrimSpace(fmt.Sprint(v))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func normalizeRememberType(value string) string {
	slug := rememberSlug(value, 40)
	switch {
	case slug == "":
		return defaultMemoryType
	case strings.Contains(slug, "report"):
		return "reporting_preference"
	case strings.Contains(slug, "pref") || strings.Contains(slug, "style") || strings.Contains(slug, "default"):
		return "preference"
	case strings.Contains(slug, "rule") || strings.Contains(slug, "policy") || strings.Contains(slug, "process"):
		return "operating_rule"
	case strings.Contains(slug, "business") || strings.Contains(slug, "fact"):
		return "business_fact"
	default:
		return slug
	}
}

func normalizeRememberKey(value string) string {
	return rememberSlug(value, maxRememberKeyRunes)
}

func deriveRememberKey(memoryType, value string) string {
	key := rememberSlug(value, maxRememberKeyRunes)
	if key == "" {
		key = "memory"
	}
	if strings.HasPrefix(key, memoryType+"_") {
		return key
	}
	return rememberSlug(memoryType+"_"+key, maxRememberKeyRunes)
}

func sanitizeRememberText(value string, limit int) string {
	value = strings.NewReplacer(
		"<tool>", "[tool]",
		"</tool>", "[/tool]",
		"<thinking>", "[thinking]",
		"</thinking>", "[/thinking]",
	).Replace(value)

	var b strings.Builder
	lastSpace := false
	for _, r := range value {
		if unicode.IsControl(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		b.WriteRune(r)
		lastSpace = false
	}
	return truncateRememberRunes(strings.TrimSpace(b.String()), limit)
}

func rememberSlug(value string, limit int) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var b strings.Builder
	lastSep := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastSep = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastSep = false
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			b.WriteRune(r)
			lastSep = false
		default:
			if b.Len() > 0 && !lastSep {
				b.WriteRune('_')
				lastSep = true
			}
		}
	}
	return strings.Trim(truncateRememberRunes(b.String(), limit), "_")
}

func truncateRememberRunes(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit]))
}
