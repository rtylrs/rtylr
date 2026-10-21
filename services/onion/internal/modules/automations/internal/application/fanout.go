package application

import (
	"context"

	"github.com/voxire/onion/internal/platform/eventbus"
)

type eventQueueRepository interface {
	DispatchEvent(context.Context, eventbus.Event) error
}

type EventFanout struct{ repo eventQueueRepository }

func NewEventFanout(repo eventQueueRepository) *EventFanout {
	return &EventFanout{repo: repo}
}

func (f *EventFanout) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	return f.repo.DispatchEvent(ctx, event)
}
