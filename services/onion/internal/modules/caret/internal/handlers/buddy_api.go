package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	auth "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	pb "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

type buddyStore interface {
	Settings(context.Context, int64) (buddy.Settings, error)
	Save(context.Context, buddy.Settings) error
	Runs(context.Context, int64) ([]buddy.Run, error)
	Actions(context.Context, int64, string) ([]buddy.Action, error)
	Action(context.Context, int64, int64) (buddy.Action, error)
	ClaimAction(context.Context, int64, int64, int64, string) (bool, error)
	FinishAction(context.Context, int64, int64, int64, string, json.RawMessage) error
}

type BuddyAPI struct {
	Store     buddyStore
	Executors []application.ToolExecutor
}

func buddyIdentity(ctx context.Context) (int64, int64, error) {
	u, ok := resolveAuthUser(ctx)
	if !ok || u.BusinessID <= 0 || !auth.IsBusinessAdmin(u) {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return u.BusinessID, u.UserID, nil
}
func settingsProto(s buddy.Settings) *pb.BuddySettings {
	return &pb.BuddySettings{Enabled: s.Enabled, Instructions: s.Instructions, IntervalMinutes: int32(s.IntervalMinutes), DailyRunCap: int32(s.DailyRunCap), RunAsUserId: s.RunAsUserID}
}
func (a *BuddyAPI) GetBuddySettings(ctx context.Context, _ *pb.GetBuddySettingsRequest) (*pb.GetBuddySettingsResponse, error) {
	b, _, err := buddyIdentity(ctx)
	if err != nil {
		return nil, err
	}
	s, err := a.Store.Settings(ctx, b)
	if errors.Is(err, sql.ErrNoRows) {
		s = buddy.Settings{IntervalMinutes: 60, DailyRunCap: 24}
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return &pb.GetBuddySettingsResponse{Settings: settingsProto(s)}, nil
}
func (a *BuddyAPI) PutBuddySettings(ctx context.Context, r *pb.PutBuddySettingsRequest) (*pb.PutBuddySettingsResponse, error) {
	b, u, err := buddyIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if r.GetIntervalMinutes() < 15 || r.GetIntervalMinutes() > 1440 || r.GetDailyRunCap() < 1 || r.GetDailyRunCap() > 96 || len(r.GetInstructions()) > 65535 {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid buddy settings")
	}
	s := buddy.Settings{BusinessID: b, RunAsUserID: u, Enabled: r.GetEnabled(), Instructions: strings.TrimSpace(r.GetInstructions()), IntervalMinutes: int(r.GetIntervalMinutes()), DailyRunCap: int(r.GetDailyRunCap())}
	if err = a.Store.Save(ctx, s); err != nil {
		return nil, err
	}
	return &pb.PutBuddySettingsResponse{Settings: settingsProto(s)}, nil
}
func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }
func nullStamp(t *time.Time) string {
	if t == nil {
		return ""
	}
	return stamp(*t)
}
func actionProto(v buddy.Action) *pb.BuddyAction {
	return &pb.BuddyAction{Id: v.ID, RunId: v.RunID, ToolName: v.ToolName, Input: v.Input, Status: v.Status, DecidedBy: v.DecidedBy, Result: v.Result, ExpiresAt: stamp(v.ExpiresAt), CreatedAt: stamp(v.CreatedAt), DecidedAt: nullStamp(v.DecidedAt)}
}
func (a *BuddyAPI) ListBuddyRuns(ctx context.Context, _ *pb.ListBuddyRunsRequest) (*pb.ListBuddyRunsResponse, error) {
	b, _, err := buddyIdentity(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := a.Store.Runs(ctx, b)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.BuddyRun, 0, len(rows))
	for _, v := range rows {
		out = append(out, &pb.BuddyRun{Id: v.ID, TriggerKind: v.TriggerKind, Status: v.Status, Report: v.Report, Error: v.Error, ToolCalls: v.ToolCalls, Tokens: v.Tokens, CreatedAt: stamp(v.CreatedAt), StartedAt: nullStamp(v.StartedAt), FinishedAt: nullStamp(v.FinishedAt)})
	}
	return &pb.ListBuddyRunsResponse{Runs: out}, nil
}
func (a *BuddyAPI) ListBuddyActions(ctx context.Context, r *pb.ListBuddyActionsRequest) (*pb.ListBuddyActionsResponse, error) {
	b, _, err := buddyIdentity(ctx)
	if err != nil {
		return nil, err
	}
	status := r.GetStatus()
	switch status {
	case "", "pending", "approved", "denied", "executed", "failed", "expired":
	default:
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid action status")
	}
	rows, err := a.Store.Actions(ctx, b, status)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.BuddyAction, 0, len(rows))
	for _, v := range rows {
		out = append(out, actionProto(v))
	}
	return &pb.ListBuddyActionsResponse{Actions: out}, nil
}
func (a *BuddyAPI) DecideBuddyAction(ctx context.Context, r *pb.DecideBuddyActionRequest) (*pb.DecideBuddyActionResponse, error) {
	b, u, err := buddyIdentity(ctx)
	if err != nil {
		return nil, err
	}
	decision := strings.ToLower(strings.TrimSpace(r.GetDecision()))
	if r.GetActionId() <= 0 || (decision != "approve" && decision != "deny") {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid action decision")
	}
	v, err := a.Store.Action(ctx, b, r.GetActionId())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, onekiterrors.New(http.StatusNotFound, "action not found")
	}
	if err != nil {
		return nil, err
	}
	status := "denied"
	if decision == "approve" {
		status = "approved"
	}
	claimed, err := a.Store.ClaimAction(ctx, b, v.ID, u, status)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, onekiterrors.New(http.StatusConflict, "action is expired or no longer pending")
	}
	if decision == "approve" {
		result := map[string]any{}
		var input map[string]any
		tool, ok := application.ToolByName(v.ToolName)
		if !ok || tool.Scope != "write" {
			result["error"] = "tool is unavailable"
		} else if err = json.Unmarshal(v.Input, &input); err != nil || input == nil {
			result["error"] = "invalid stored input"
		} else {
			input = application.NormalizeToolInput(tool.Name, input)
			if err = application.ValidateToolInvocation(application.ModeAgent, tool, input); err != nil {
				result["error"] = "invalid stored input"
			} else {
				var executed bool
				var output map[string]any
				output, executed, _, err = application.ExecuteToolWithTimeout(ctx, a.Executors, tool.Name, input)
				if err != nil {
					result["error"] = userFacingToolError(err)
				} else if !executed {
					result["error"] = "tool is unavailable"
				} else {
					result = output
					status = "executed"
				}
			}
		}
		if status != "executed" {
			status = "failed"
		}
		raw, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			raw = []byte(`{"error":"invalid execution result"}`)
			status = "failed"
		}
		finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err = a.Store.FinishAction(finishCtx, b, v.ID, u, status, raw); err != nil {
			return nil, err
		}
	}
	v, err = a.Store.Action(context.WithoutCancel(ctx), b, v.ID)
	if err != nil {
		return nil, err
	}
	return &pb.DecideBuddyActionResponse{Action: actionProto(v)}, nil
}
