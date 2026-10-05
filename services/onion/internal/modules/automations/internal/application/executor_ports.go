package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
)

type runRepository interface {
	ListDueRuns(context.Context, int32) ([]domain.Run, error)
	ClaimRun(context.Context, uint64) error
	MarkRunSucceeded(context.Context, uint64) error
	MarkRunFailed(context.Context, uint64, int64, string) error
	MarkRunDead(context.Context, uint64, string) error
	RecoverAbandoned(context.Context, int64) (int64, error)
	GetAutomationForRun(context.Context, int64, int64) (domain.Automation, error)
	RecordExecution(context.Context, int64, int64) error
}

type eventReader interface {
	GetEvent(context.Context, uint64, int64) (eventbus.Event, error)
}

type EmailMessage struct {
	To        string
	Subject   string
	HTML      string
	Text      string
	DedupeKey string
}

type emailQueue interface {
	Enqueue(context.Context, EmailMessage) error
}
