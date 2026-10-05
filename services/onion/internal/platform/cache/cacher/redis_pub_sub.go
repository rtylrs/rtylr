package cacher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/voxire/onion/internal/platform/logger"
)

type redisConsumer interface {
	Execute(ctx context.Context) error
}

type RedisPubSub[C redisConsumer] struct {
	redisCacheCluster redis.UniversalClient
}

func NewPubSub[C redisConsumer]() RedisPubSub[C] {
	return RedisPubSub[C]{
		redisCacheCluster: client,
	}
}

func (r RedisPubSub[C]) Publish(channel string, message C) {
	if r.redisCacheCluster == nil {
		return
	}

	ctx := context.Background()

	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()

	payload, err := json.Marshal(message)
	if err != nil {
		logger.Ctx(ctx).Warn("unable to marshal pub sub message",
			logger.String("channel", channel),
			logger.String("message", string(payload)),
		)

		return
	}

	cmd := r.redisCacheCluster.Publish(ctx, channel, payload)

	err = cmd.Err()
	if err != nil {
		logger.Ctx(ctx).Warn("Error while publishing to redis",
			logger.Error(err),
			logger.String("channel", channel),
			logger.String("message", string(payload)),
		)
	}
}

func (r RedisPubSub[C]) SubscribeContext(ctx context.Context, channel string) {
	if r.redisCacheCluster == nil {
		return
	}

	subscriber := r.redisCacheCluster.Subscribe(ctx, channel)
	defer func() {
		if err := subscriber.Close(); err != nil {
			logger.Ctx(ctx).Warn("Unable to close redis pub sub",
				logger.Error(err),
				logger.String("channel", channel),
			)
		}
	}()

	const redisReadPeriod = 3 * time.Second

	for {
		message, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			logger.Ctx(ctx).Warn("Unable to receive message from redis pub sub",
				logger.Error(err),
				logger.String("channel", channel),
			)

			timer := time.NewTimer(redisReadPeriod)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			case <-timer.C:
				continue
			}
		}

		consumer := *new(C)

		err = json.Unmarshal([]byte(message.Payload), &consumer)
		if err != nil {
			logger.Ctx(ctx).Warn("Unable to unmarshal message from redis pub sub",
				logger.Error(err),
				logger.String("channel", channel),
				logger.String("msg", message.Payload),
			)

			continue
		}

		err = consumer.Execute(ctx)
		if err != nil {
			logger.Ctx(ctx).Warn("Redis Consumer Failed Executing",
				logger.String("channel", channel),
				logger.Error(err),
			)
		}
	}
}
