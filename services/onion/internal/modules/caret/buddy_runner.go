package echo

import (
	"context"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

// StartBuddies runs queued buddy checks with the same tool executors the chat
// uses. It does nothing unless buddies are switched on and the AI is available.
func (m *Module) StartBuddies(ctx context.Context, buddies *buddy.Module, access authPkg.EffectiveAccessResolver) {
	m.buddies = buddies
	if buddies == nil || !m.snipsService.IsAvailable() {
		return
	}
	agent := &application.BuddyAgent{
		AI:        m.snipsService,
		Executors: func() []application.ToolExecutor { return m.toolExecutors },
		Store:     buddies,
		Quota:     m.quotaService,
	}
	buddies.StartRunner(ctx, buddyAgent{agent}, access)
}

type buddyAgent struct{ agent *application.BuddyAgent }

func (a buddyAgent) Run(ctx context.Context, task buddy.Task) (buddy.Outcome, error) {
	out, err := a.agent.Run(ctx, application.BuddyTask{
		RunID: task.RunID, BusinessID: task.BusinessID, UserID: task.UserID, Instructions: task.Instructions,
	})
	return buddy.Outcome{Report: out.Report, ToolCalls: out.ToolCalls, Proposals: out.Proposals, Tokens: out.Tokens}, err
}
