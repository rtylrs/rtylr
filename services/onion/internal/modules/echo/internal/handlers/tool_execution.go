package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxToolCallsPerTurn   = 6
	maxToolInputBytes     = 12 * 1024
	maxToolInputKeys      = 32
	toolExecutionTimeout  = 15 * time.Second
	maxToolStringValueLen = 4000
)

var (
	ErrToolCallLimit     = errors.New("tool call limit reached for this response")
	ErrDuplicateToolCall = errors.New("the same tool call was already executed")
)

type toolCallExecutionState struct {
	count int
	seen  map[string]struct{}
	limit int
}

func newToolCallExecutionState() *toolCallExecutionState {
	return &toolCallExecutionState{seen: make(map[string]struct{}), limit: maxToolCallsPerTurn}
}

func (s *toolCallExecutionState) admit(name string, input map[string]any) error {
	if s == nil {
		return nil
	}
	limit := s.limit
	if limit <= 0 {
		limit = maxToolCallsPerTurn
	}
	if s.count >= limit {
		return ErrToolCallLimit
	}
	key := toolCallFingerprint(name, input)
	if _, ok := s.seen[key]; ok {
		return ErrDuplicateToolCall
	}
	s.seen[key] = struct{}{}
	s.count++
	return nil
}

func toolCallFingerprint(name string, input map[string]any) string {
	payload, err := json.Marshal(struct {
		Name  string         `json:"name"`
		Input map[string]any `json:"input,omitempty"`
	}{Name: strings.TrimSpace(name), Input: input})
	if err != nil {
		payload = []byte(strings.TrimSpace(name))
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func validateToolInvocation(mode CaretMode, tool ToolDef, input map[string]any) error {
	input = normalizeToolInput(tool.Name, input)
	if !toolAllowedForMode(mode, tool) {
		return fmt.Errorf("tool %q is not available in %s mode", tool.Name, mode)
	}
	if len(input) > maxToolInputKeys {
		return fmt.Errorf("tool input contains too many fields")
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("tool input is not valid JSON: %w", err)
	}
	if len(encoded) > maxToolInputBytes {
		return fmt.Errorf("tool input is too large")
	}
	for key, value := range input {
		if strings.TrimSpace(key) == "" {
			return errors.New("tool input contains an empty field name")
		}
		if err := validateToolValue(value, 0); err != nil {
			return fmt.Errorf("tool input %q: %w", key, err)
		}
	}
	return nil
}

func validateToolValue(value any, depth int) error {
	if depth > 5 {
		return errors.New("nested input is too deep")
	}
	switch value := value.(type) {
	case string:
		if len(value) > maxToolStringValueLen {
			return errors.New("string value is too long")
		}
	case map[string]any:
		if len(value) > maxToolInputKeys {
			return errors.New("nested object contains too many fields")
		}
		for key, nested := range value {
			if strings.TrimSpace(key) == "" {
				return errors.New("nested object contains an empty field name")
			}
			if err := validateToolValue(nested, depth+1); err != nil {
				return err
			}
		}
	case []any:
		if len(value) > maxToolInputKeys {
			return errors.New("array contains too many values")
		}
		for _, nested := range value {
			if err := validateToolValue(nested, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func executeToolWithTimeout(ctx context.Context, executors []ToolExecutor, name string, input map[string]any) (output map[string]any, executed bool, err error, duration time.Duration) {
	started := time.Now()
	defer func() {
		duration = time.Since(started)
		if recovered := recover(); recovered != nil {
			output = nil
			executed = false
			err = fmt.Errorf("tool executor panic: %v", recovered)
		}
	}()

	callCtx, cancel := context.WithTimeout(ctx, toolExecutionTimeout)
	defer cancel()
	for _, executor := range executors {
		if executor == nil {
			continue
		}
		output, executed, err = executor.ExecuteCaretTool(callCtx, name, input)
		output = sanitizeToolOutput(output)
		if err != nil || executed {
			return output, executed, err, 0
		}
		select {
		case <-callCtx.Done():
			return nil, false, callCtx.Err(), 0
		default:
		}
	}
	return nil, false, nil, 0
}
