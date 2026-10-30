package buddy

import (
	"context"

	supportcontract "github.com/voxire/onion/internal/modules/support/contract"
	"github.com/voxire/onion/internal/platform/eventbus"
)

func (m *Module) DispatchEvent(ctx context.Context, event eventbus.Event) error {
	if !Enabled() {
		return nil
	}
	switch event.Type {
	case supportcontract.TicketCreated, supportcontract.CustomerMessageCreated:
		return m.repo.EnqueueEvent(ctx, event.ID, event.BusinessID)
	default:
		return nil
	}
}
