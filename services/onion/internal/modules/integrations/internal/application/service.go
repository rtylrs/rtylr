package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	integrationsScope = "integrations"
	integrationsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Integration, error)
	Connect(ctx context.Context, businessID int64, provider string, config json.RawMessage) (int64, error)
	Disconnect(ctx context.Context, businessID, integrationID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateIntegrationsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(integrationsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Integration, error) {
	return cacher.Of[[]domain.Integration]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(integrationsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Integration, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, integrationsTTL),
	)
}

const maxConfigBytes = 8 << 10

func (s *Service) Connect(ctx context.Context, businessID int64, provider string, config json.RawMessage) (int64, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if _, ok := domain.Providers[provider]; !ok {
		return 0, domain.ErrInvalidProvider
	}
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	var object map[string]json.RawMessage
	if len(config) > maxConfigBytes || json.Unmarshal(config, &object) != nil || object == nil {
		return 0, domain.ErrInvalidConfig
	}
	id, err := s.repo.Connect(ctx, businessID, provider, config)
	if err != nil {
		return 0, err
	}
	invalidateIntegrationsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Disconnect(ctx context.Context, businessID, integrationID int64) error {
	if integrationID == 0 {
		return errors.New("integration id is required")
	}
	if err := s.repo.Disconnect(ctx, businessID, integrationID); err != nil {
		return err
	}
	invalidateIntegrationsScope(ctx, businessID)
	return nil
}

var secretKeyFragments = []string{"secret", "token", "key", "password", "passwd", "credential", "authorization", "private"}

// RedactConfig masks secret-looking values anywhere in a stored config, so a
// listing never hands credentials back to the browser.
func RedactConfig(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return json.RawMessage("{}")
	}
	redacted, err := json.Marshal(redactValue(value))
	if err != nil {
		return json.RawMessage("{}")
	}
	return redacted
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if isSecretKey(key) {
				typed[key] = "[redacted]"
				continue
			}
			typed[key] = redactValue(child)
		}
		return typed
	case []any:
		for i, child := range typed {
			typed[i] = redactValue(child)
		}
		return typed
	default:
		return value
	}
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, fragment := range secretKeyFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}
