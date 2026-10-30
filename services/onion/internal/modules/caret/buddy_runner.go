package echo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/modules/caret/memory"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
)

const buddyThreadHistory = 30

// buddyThreads is the slice of chat memory the buddy needs: one chat per
// person that routines read from and post into.
type buddyThreads interface {
	BuddyThread(ctx context.Context, businessID int64, userID int) (string, error)
	BuddyLastPost(ctx context.Context, sessionID string, userID int) (time.Time, bool, error)
	RecentMessages(ctx context.Context, sessionID string, userID, limit int) ([]memory.ThreadMessage, error)
	StoreMessage(ctx context.Context, sessionID string, userID int, role, content string, tokens int) (any, error)
}

// StartBuddies runs queued buddy checks with the same tool executors the chat
// uses. It does nothing unless buddies are switched on and the AI is available.
func (m *Module) StartBuddies(ctx context.Context, buddies *buddy.Module, access authPkg.EffectiveAccessResolver) {
	m.buddies = buddies
	threads, ok := m.memoryService.(buddyThreads)
	if buddies == nil || !ok || !m.snipsService.IsAvailable() {
		return
	}
	agent := &application.BuddyAgent{
		AI:        m.snipsService,
		Executors: func() []application.ToolExecutor { return m.toolExecutors },
		Quota:     m.quotaService,
	}
	buddies.StartRunner(ctx, buddyAgent{agent: agent, threads: threads}, access)
}

type buddyAgent struct {
	agent   *application.BuddyAgent
	threads buddyThreads
}

// Run works the task with the person's recent chat as context, then posts the
// result into that chat.
func (a buddyAgent) Run(ctx context.Context, task buddy.Task) (buddy.Outcome, error) {
	sessionID, err := a.threads.BuddyThread(ctx, task.BusinessID, int(task.UserID))
	if err != nil {
		return buddy.Outcome{}, fmt.Errorf("open buddy chat: %w", err)
	}
	history, err := a.threads.RecentMessages(ctx, sessionID, int(task.UserID), buddyThreadHistory)
	if err != nil {
		logger.Ctx(ctx).Warn("caret buddy: reading chat failed", logger.Error(err))
	}
	in := application.BuddyTask{
		RunID: task.RunID, BusinessID: task.BusinessID, UserID: task.UserID,
		Instructions: task.Instructions, Thread: formatThread(history),
	}
	if task.Routine != nil {
		in.RoutineName, in.RoutineInstructions = task.Routine.Name, task.Routine.Instructions
	}
	out, err := a.agent.Run(ctx, in)
	result := buddy.Outcome{Report: out.Report, ToolCalls: out.ToolCalls, Proposals: out.Proposals, Tokens: out.Tokens}
	if err != nil || strings.TrimSpace(out.Report) == "" {
		return result, err
	}
	message := out.Report
	if task.Routine != nil {
		message = "**" + task.Routine.Name + "**\n\n" + message
	}
	if _, err := a.threads.StoreMessage(ctx, sessionID, int(task.UserID), "assistant", message, len(message)/4); err != nil {
		return result, fmt.Errorf("post to buddy chat: %w", err)
	}
	return result, nil
}

func formatThread(messages []memory.ThreadMessage) string {
	var b strings.Builder
	for _, message := range messages {
		who := "Them"
		if message.Role == "assistant" {
			who = "You"
		}
		fmt.Fprintf(&b, "[%s] %s: %s\n", message.CreatedAt.UTC().Format("Jan 2 15:04 UTC"), who, application.TruncateUTF8ToBytes(message.Content, 1500))
	}
	return b.String()
}
