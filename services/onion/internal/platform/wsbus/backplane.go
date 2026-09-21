package wsbus

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// wsBackplaneChannel is the Redis pub/sub channel every replica publishes to and
// subscribes on, so a realtime event produced on one replica reaches clients
// connected to any replica.
const wsBackplaneChannel = "wsbus-events"

// backplaneEnabled gates whether Publish routes through Redis. It is written by
// EnableRedisBackplane and read concurrently by publishers, so it is atomic.
var backplaneEnabled atomic.Bool

var (
	backplaneMu         sync.Mutex
	backplaneCancel     context.CancelFunc
	backplaneGeneration uint64
)

// startBackplaneSubscriber starts the Redis subscriber for the server run whose
// context is passed in. It is a variable so tests can observe the subscriber
// lifecycle without a Redis server.
var startBackplaneSubscriber = func(ctx context.Context) {
	go cacher.NewPubSub[backplaneEvent]().SubscribeContext(ctx, wsBackplaneChannel)
}

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
// of only the one that produced them. Call once per server run AFTER SetDefault
// and only when Redis is enabled, passing the server's context: the subscriber
// stops when the server stops, and calling this again (the in-process restart
// loop re-runs setup routes) cancels the previous subscriber before starting a
// new one. Stacking subscribers would fan every event into the live hub once
// per subscriber, duplicating invalidations after each restart, and leak a
// goroutine plus a Redis connection each time.
//
// Once the subscriber's context ends and no replacement has been started,
// Publish falls back to in-process fan-out rather than publishing into a
// channel nobody is reading on this replica.
func EnableRedisBackplane(ctx context.Context) {
	backplaneEnabled.Store(true)

	backplaneMu.Lock()
	if backplaneCancel != nil {
		backplaneCancel()
	}

	subCtx, cancel := context.WithCancel(ctx)
	backplaneCancel = cancel
	backplaneGeneration++
	generation := backplaneGeneration
	backplaneMu.Unlock()

	startBackplaneSubscriber(subCtx)

	go func() {
		<-subCtx.Done()

		backplaneMu.Lock()
		defer backplaneMu.Unlock()

		// A newer call has already replaced this subscription; leave its
		// state alone. Otherwise the backplane is no longer running, so stop
		// routing publishes into Redis.
		if backplaneGeneration == generation {
			backplaneEnabled.Store(false)
			backplaneCancel = nil
		}
	}()
}

// publishBackplane sends an event to the Redis channel. Each replica's subscriber
// (including this one) then delivers it to its local hub exactly once, which is
// why the local Publish path is skipped when the backplane is active.
func publishBackplane(e Event) {
	cacher.NewPubSub[backplaneEvent]().Publish(wsBackplaneChannel, backplaneEvent{Event: e})
}
