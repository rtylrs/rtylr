package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/buddy/internal/infrastructure/persistence/sqlc"
)

func actionFromRow(r sqlc.CaretBuddyAction) domain.Action {
	return domain.Action{ID: r.ID, RunID: r.RunID, DecidedBy: r.DecidedBy.Int64, ToolName: r.ToolName, Status: string(r.Status), Input: r.Input, Result: r.Result, ExpiresAt: r.ExpiresAt, CreatedAt: r.CreatedAt, DecidedAt: timePointer(r.DecidedAt)}
}
func (r *Repository) Runs(ctx context.Context, businessID int64) ([]domain.RunView, error) {
	rows, err := r.q.ListRunsByBusiness(ctx, sqlc.ListRunsByBusinessParams{BusinessID: businessID, Limit: 100})
	if err != nil {
		return nil, err
	}
	out := make([]domain.RunView, 0, len(rows))
	for _, v := range rows {
		out = append(out, domain.RunView{ID: v.ID, TriggerKind: v.TriggerKind, Status: string(v.Status), Report: v.Report.String, Error: v.Error.String, ToolCalls: v.ToolCalls, Tokens: v.Tokens, CreatedAt: v.CreatedAt, StartedAt: timePointer(v.StartedAt), FinishedAt: timePointer(v.FinishedAt)})
	}
	return out, nil
}
func (r *Repository) Actions(ctx context.Context, businessID int64, status string) ([]domain.Action, error) {
	var rows []sqlc.CaretBuddyAction
	var err error
	if status == "" {
		rows, err = r.q.ListAllActionsByBusiness(ctx, sqlc.ListAllActionsByBusinessParams{BusinessID: businessID, Limit: 100})
	} else {
		rows, err = r.q.ListActionsByBusiness(ctx, sqlc.ListActionsByBusinessParams{BusinessID: businessID, Status: sqlc.CaretBuddyActionsStatus(status), Limit: 100})
	}
	if err != nil {
		return nil, err
	}
	out := make([]domain.Action, 0, len(rows))
	for _, v := range rows {
		out = append(out, actionFromRow(v))
	}
	return out, nil
}
func (r *Repository) Action(ctx context.Context, businessID, id int64) (domain.Action, error) {
	v, err := r.q.GetAction(ctx, sqlc.GetActionParams{ID: id, BusinessID: businessID})
	return actionFromRow(v), err
}
func (r *Repository) ClaimAction(ctx context.Context, businessID, id, userID int64, status string) (bool, error) {
	n, err := r.q.DecideAction(ctx, sqlc.DecideActionParams{ID: id, BusinessID: businessID, DecidedBy: sql.NullInt64{Int64: userID, Valid: true}, Status: sqlc.CaretBuddyActionsStatus(status)})
	return n == 1, err
}
func (r *Repository) FinishAction(ctx context.Context, businessID, id, userID int64, status string, result json.RawMessage) error {
	n, err := r.q.FinishAction(ctx, sqlc.FinishActionParams{ID: id, BusinessID: businessID, DecidedBy: sql.NullInt64{Int64: userID, Valid: true}, Status: sqlc.CaretBuddyActionsStatus(status), Result: result})
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("action state changed before completion")
	}
	return nil
}

func timePointer(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
