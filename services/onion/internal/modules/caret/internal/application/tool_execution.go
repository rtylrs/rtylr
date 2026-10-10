package application

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
	MaxToolCallsPerTurn   = 10
	maxToolInputBytes     = 12 * 1024
	maxToolInputKeys      = 32
	ToolExecutionTimeout  = 15 * time.Second
	MaxToolStringValueLen = 4000
	MaxToolOutputBytes    = 24000
)

// slowToolTimeouts lets the tools that do several steps of their own run
// longer than a single read.
var slowToolTimeouts = map[string]time.Duration{
	"web.fetch":     25 * time.Second,
	"places.search": 26 * time.Second,
	"agents.run":    150 * time.Second,
}

// ToolTimeout is how long one call of the named tool may run.
func ToolTimeout(name string) time.Duration {
	if timeout, ok := slowToolTimeouts[name]; ok {
		return timeout
	}
	return ToolExecutionTimeout
}

var (
	ErrToolCallLimit     = errors.New("tool call limit reached for this response")
	ErrDuplicateToolCall = errors.New("the same tool call was already executed")
)

type ToolExecutor interface {
	ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (output map[string]any, executed bool, err error)
}

type ToolCallExecutionState = toolCallExecutionState

func NewToolCallExecutionState() *ToolCallExecutionState { return newToolCallExecutionState() }
func (s *toolCallExecutionState) Admit(name string, input map[string]any) error {
	return s.admit(name, input)
}

func (s *toolCallExecutionState) SetLimit(limit int) {
	if s != nil && limit > 0 {
		s.limit = limit
	}
}

type toolCallExecutionState struct {
	count int
	seen  map[string]struct{}
	limit int
}

func newToolCallExecutionState() *toolCallExecutionState {
	return &toolCallExecutionState{seen: make(map[string]struct{}), limit: MaxToolCallsPerTurn}
}

func (s *toolCallExecutionState) admit(name string, input map[string]any) error {
	if s == nil {
		return nil
	}
	limit := s.limit
	if limit <= 0 {
		limit = MaxToolCallsPerTurn
	}
	if s.count >= limit {
		return ErrToolCallLimit
	}
	key := ToolCallFingerprint(name, input)
	if _, ok := s.seen[key]; ok {
		return ErrDuplicateToolCall
	}
	s.seen[key] = struct{}{}
	s.count++
	return nil
}

func ToolCallFingerprint(name string, input map[string]any) string {
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

func ValidateToolInvocation(mode CaretMode, tool ToolDef, input map[string]any) error {
	input = NormalizeToolInput(tool.Name, input)
	if !ToolAllowedForMode(mode, tool) {
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
		if len(value) > MaxToolStringValueLen {
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

func ExecuteToolWithTimeout(ctx context.Context, executors []ToolExecutor, name string, input map[string]any) (output map[string]any, executed bool, duration time.Duration, err error) {
	started := time.Now()
	defer func() {
		duration = time.Since(started)
		if recovered := recover(); recovered != nil {
			output = nil
			executed = false
			err = fmt.Errorf("tool executor panic: %v", recovered)
		}
	}()

	callCtx, cancel := context.WithTimeout(ctx, ToolTimeout(name))
	defer cancel()
	for _, executor := range executors {
		if executor == nil {
			continue
		}
		output, executed, err = executor.ExecuteCaretTool(callCtx, name, input)
		output = SanitizeToolOutput(output)
		if err != nil || executed {
			return output, executed, 0, err
		}
		select {
		case <-callCtx.Done():
			return nil, false, 0, callCtx.Err()
		default:
		}
	}
	return nil, false, 0, nil
}
