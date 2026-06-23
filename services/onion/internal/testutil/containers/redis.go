package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisConfig struct {
	Address string
}

func StartRedis(ctx context.Context) (RedisConfig, func(context.Context) error, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:8-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return RedisConfig{}, nil, err
	}

	cleanup := func(ctx context.Context) error {
		return container.Terminate(ctx)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = cleanup(ctx)
		return RedisConfig{}, nil, err
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		_ = cleanup(ctx)
		return RedisConfig{}, nil, err
	}

	return RedisConfig{Address: fmt.Sprintf("%s:%s", host, port.Port())}, cleanup, nil
}
