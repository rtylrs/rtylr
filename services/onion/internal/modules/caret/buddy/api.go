package buddy

import (
	"context"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
)

type Action = domain.Action
type Run = domain.RunView

func (m *Module) Settings(ctx context.Context, businessID int64) (Settings, error) {
	b, err := m.repo.BuddyByBusiness(ctx, businessID)
	return Settings{BusinessID: b.BusinessID, RunAsUserID: b.RunAsUserID, Enabled: b.Enabled, Instructions: b.Instructions, IntervalMinutes: b.IntervalMinutes, DailyRunCap: b.DailyRunCap}, err
}
func (m *Module) Runs(ctx context.Context, businessID int64) ([]Run, error) {
	return m.repo.Runs(ctx, businessID)
}
func (m *Module) Actions(ctx context.Context, businessID int64, status string) ([]Action, error) {
	return m.repo.Actions(ctx, businessID, status)
}
func (m *Module) Action(ctx context.Context, businessID, id int64) (Action, error) {
	return m.repo.Action(ctx, businessID, id)
}
func (m *Module) ClaimAction(ctx context.Context, businessID, id, userID int64, status string) (bool, error) {
	return m.repo.ClaimAction(ctx, businessID, id, userID, status)
}
func (m *Module) FinishAction(ctx context.Context, businessID, id, userID int64, status string, result json.RawMessage) error {
	return m.repo.FinishAction(ctx, businessID, id, userID, status, result)
}
