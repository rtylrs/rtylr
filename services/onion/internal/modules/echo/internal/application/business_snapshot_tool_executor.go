package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

type BusinessSnapshotToolExecutor struct {
	AnalystService AnalystService
}

func NewBusinessSnapshotToolExecutor(analystService AnalystService) *BusinessSnapshotToolExecutor {
	return &BusinessSnapshotToolExecutor{AnalystService: analystService}
}

func (e *BusinessSnapshotToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != BusinessSnapshotToolName {
		return nil, false, nil
	}

	businessID, userID, err := businessSnapshotToolIdentity(ctx)
	if err != nil {
		return nil, true, err
	}

	question := strings.TrimSpace(businessSnapshotToolString(firstBusinessSnapshotToolValue(input, "question", "prompt", "query")))
	if question == "" {
		question = "business snapshot"
	}

	grounding, ok := BuildBusinessSnapshotGrounding(ctx, e.AnalystService, question, businessID, userID, time.Now().UTC())
	if !ok {
		return nil, true, errors.New("business snapshot is unavailable")
	}

	output := map[string]any{
		"section_count": grounding.RowCount,
		"row_count":     grounding.RowCount,
		"truncated":     grounding.Truncated,
		"sections":      grounding.Rows,
		"rows":          grounding.Rows,
		"duration_ms":   grounding.DurationMs,
	}
	if strings.TrimSpace(grounding.SpecJSON) != "" {
		output["spec"] = json.RawMessage(grounding.SpecJSON)
	}
	return output, true, nil
}

func businessSnapshotToolIdentity(ctx context.Context) (int64, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.BusinessID > 0 && device.User.UserID > 0 {
		return device.User.BusinessID, device.User.UserID, nil
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if ok && authUser != nil && authUser.BusinessID > 0 && authUser.UserID > 0 {
		return authUser.BusinessID, authUser.UserID, nil
	}
	return 0, 0, errors.New("authentication is required")
}

func firstBusinessSnapshotToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func businessSnapshotToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
