package gpu

import (
	"context"
	"errors"
	"fmt"
	"time"

	rpSdk "github.com/runpod/go-sdk/pkg/sdk"
	rpConfig "github.com/runpod/go-sdk/pkg/sdk/config"
	rpEndpoint "github.com/runpod/go-sdk/pkg/sdk/endpoint"
	appConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	rpCfg             *rpConfig.Config
	defaultEndpointID string
	ErrNotInitialized = errors.New("RunPod GPU not initialized")
	ErrJobFailed      = errors.New("RunPod job failed")
	ErrJobTimeout     = errors.New("RunPod job timed out")
	ErrMissingJobID   = errors.New("RunPod returned no job ID")
)

func Init(ctx context.Context) error {
	apiKey := appConfig.GetRunPodAPIKey()
	if apiKey == "" {
		logger.Ctx(ctx).Warn("RunPod API key not configured, GPU features disabled")

		return nil
	}

	rpCfg = &rpConfig.Config{
		ApiKey: rpSdk.String(apiKey),
	}
	defaultEndpointID = appConfig.GetRunPodEndpointID()

	logger.Ctx(ctx).Info("RunPod GPU initialized",
		logger.String("endpoint", defaultEndpointID),
	)

	return nil
}

func IsConfigured() bool {
	return rpCfg != nil && defaultEndpointID != ""
}

func newEndpoint(endpointID string) (*rpEndpoint.Endpoint, error) {
	if rpCfg == nil {
		return nil, ErrNotInitialized
	}

	return rpEndpoint.New(rpCfg, &rpEndpoint.Option{
		EndpointId: rpSdk.String(endpointID),
	})
}

func RunAsync(_ context.Context, input map[string]any) (string, error) {
	ep, err := newEndpoint(defaultEndpointID)
	if err != nil {
		return "", err
	}

	result, err := ep.Run(&rpEndpoint.RunInput{
		JobInput: &rpEndpoint.JobInput{Input: input},
	})
	if err != nil {
		return "", fmt.Errorf("RunPod async request failed: %w", err)
	}

	if result.Id == nil {
		return "", ErrMissingJobID
	}

	return *result.Id, nil
}

func RunSync(_ context.Context, input map[string]any, timeout int) (*rpEndpoint.RunSyncOutput, error) {
	ep, err := newEndpoint(defaultEndpointID)
	if err != nil {
		return nil, err
	}

	result, err := ep.RunSync(&rpEndpoint.RunSyncInput{
		JobInput: &rpEndpoint.JobInput{Input: input},
		Timeout:  rpSdk.Int(timeout),
	})
	if err != nil {
		return nil, fmt.Errorf("RunPod sync request failed: %w", err)
	}

	if result.Status != nil && *result.Status == "FAILED" {
		errMsg := ""
		if result.Error != nil {
			errMsg = *result.Error
		}

		return result, fmt.Errorf("%w: %s", ErrJobFailed, errMsg)
	}

	return result, nil
}

func GetStatus(_ context.Context, jobID string) (*rpEndpoint.StatusOutput, error) {
	ep, err := newEndpoint(defaultEndpointID)
	if err != nil {
		return nil, err
	}

	return ep.Status(&rpEndpoint.StatusInput{
		Id: rpSdk.String(jobID),
	})
}

func PollUntilDone(ctx context.Context, jobID string, pollInterval, timeout time.Duration) (*rpEndpoint.StatusOutput, error) {
	deadline := time.After(timeout)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, ErrJobTimeout
		case <-ticker.C:
			status, err := GetStatus(ctx, jobID)
			if err != nil {
				return nil, err
			}

			if status.Status == nil {
				continue
			}

			switch *status.Status {
			case "COMPLETED":
				return status, nil
			case "FAILED":
				errMsg := ""
				if status.Error != nil {
					errMsg = *status.Error
				}

				return status, fmt.Errorf("%w: %s", ErrJobFailed, errMsg)
			default:
				continue
			}
		}
	}
}

func RunOnEndpoint(_ context.Context, endpointID string, input map[string]any) (string, error) {
	ep, err := newEndpoint(endpointID)
	if err != nil {
		return "", err
	}

	result, err := ep.Run(&rpEndpoint.RunInput{
		JobInput: &rpEndpoint.JobInput{Input: input},
	})
	if err != nil {
		return "", fmt.Errorf("RunPod async request failed: %w", err)
	}

	if result.Id == nil {
		return "", ErrMissingJobID
	}

	return *result.Id, nil
}

func GetHealth(_ context.Context) (*rpEndpoint.HealthOutput, error) {
	ep, err := newEndpoint(defaultEndpointID)
	if err != nil {
		return nil, err
	}

	return ep.Health(&rpEndpoint.HealthInput{})
}
