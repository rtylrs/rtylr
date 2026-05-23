package cacher

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const DefaultCacheTTL = time.Hour * 12

type ClientOption func(c *cacherConfiguration)

type cacherConfiguration struct {
	expiry              time.Duration
	memoryOnly          bool
	remoteOnly          bool
	evictionProbability float32
}

func getDefaultConfiguration() cacherConfiguration {
	return cacherConfiguration{
		expiry:              DefaultCacheTTL,
		memoryOnly:          false,
		remoteOnly:          false,
		evictionProbability: 0,
	}
}

func getConfiguration(options ...ClientOption) cacherConfiguration {
	configuration := getDefaultConfiguration()
	for _, option := range options {
		option(&configuration)
	}

	return configuration
}

func WithRandomInvalidation(probability float32) ClientOption {
	return func(c *cacherConfiguration) {
		c.evictionProbability = probability
	}
}

func WithExpiry(ctx context.Context, expiry time.Duration) ClientOption {
	return func(c *cacherConfiguration) {
		if expiry < 0 {
			logger.Ctx(ctx).Warn("negative ttl value",
				logger.Any("expiry", expiry),
			)

			c.expiry = DefaultCacheTTL
		} else {
			c.expiry = expiry
		}
	}
}

func MemoryOnly() ClientOption {
	return func(c *cacherConfiguration) {
		c.memoryOnly = true
		c.remoteOnly = false
	}
}

func RemoteOnly() ClientOption {
	return func(c *cacherConfiguration) {
		c.remoteOnly = true
		c.memoryOnly = false
	}
}
