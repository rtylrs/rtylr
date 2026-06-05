package wsbus

import (
	"context"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// wsBackplaneChannel is the Redis pub/sub channel every replica publishes to and
// subscribes on, so a realtime event produced on one replica reaches clients
// connected to any replica.
const wsBackplaneChannel = "wsbus-events"

// backplaneEnabled gates whether Publish routes through Redis. It is set once at
// startup by EnableRedisBackplane and only read thereafter, so no lock is needed.
var backplaneEnabled bool

// backplaneEvent is the Redis pub/sub envelope. Its Execute runs on every replica
// that receives the message and performs a LOCAL fan-out only (Hub.Publish) — it
// never re-publishes to Redis, so there is no loop.
type backplaneEvent struct {
	Event
}

// Execute is invoked by the generic cacher pub/sub consumer for each received
// message. It delivers the event to this replica's local hub.
func (b backplaneEvent) Execute(_ context.Context) error {
	if h := Default(); h != nil {
		h.Publish(b.Event)
	}
	return nil
}

// EnableRedisBackplane makes Publish fan events out across all replicas instead
// of only the one that produced them. Call once at startup AFTER SetDefault and
// only when Redis is enabled. It starts a subscriber that relays incoming events
// to this replica's local hub; without it, Publish stays in-process (correct for
// a single replica).
func EnableRedisBackplane() {
	backplaneEnabled = true
	go cacher.NewPubSub[backplaneEvent]().Subscribe(wsBackplaneChannel)
}

// publishBackplane sends an event to the Redis channel. Each replica's subscriber
// (including this one) then delivers it to its local hub exactly once, which is
// why the local Publish path is skipped when the backplane is active.
func publishBackplane(e Event) {
	cacher.NewPubSub[backplaneEvent]().Publish(wsBackplaneChannel, backplaneEvent{Event: e})
}
