package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/custom_domains/internal/domain"
	"github.com/voxire/onion/internal/modules/custom_domains/internal/infrastructure"
)

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) Get(ctx context.Context, businessID int64) (domain.CustomDomain, error) {
	return s.repo.Get(ctx, businessID)
}

func (s *Service) Register(ctx context.Context, ownerID, businessID int64, rawDomain string) (domain.CustomDomain, error) {
	domainName := strings.ToLower(strings.TrimSpace(rawDomain))
	if err := validateDomain(domainName); err != nil {
		return domain.CustomDomain{}, err
	}
	if current, err := s.repo.Get(ctx, businessID); err == nil && current.Domain != domainName {
		return domain.CustomDomain{}, domain.ErrConflict
	}
	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return domain.CustomDomain{}, err
	}
	return s.repo.Register(ctx, domain.CustomDomain{OwnerID: ownerID, BusinessID: businessID, Domain: domainName, VerificationToken: hex.EncodeToString(tokenBytes)})
}

func (s *Service) Verify(ctx context.Context, businessID int64, token string) (domain.CustomDomain, error) {
	current, err := s.repo.Get(ctx, businessID)
	if err != nil {
		return domain.CustomDomain{}, err
	}
	if strings.TrimSpace(token) == "" || token != current.VerificationToken {
		return domain.CustomDomain{}, domain.ErrInvalidToken
	}
	if err := s.repo.Verify(ctx, businessID, token); err != nil {
		return domain.CustomDomain{}, err
	}
	return s.repo.Get(ctx, businessID)
}

func (s *Service) Remove(ctx context.Context, businessID int64) error {
	return s.repo.Remove(ctx, businessID)
}

func validateDomain(value string) error {
	if value == "" || len(value) > 255 || strings.ContainsAny(value, "/: ") || !strings.Contains(value, ".") {
		return fmt.Errorf("invalid custom domain")
	}
	return nil
}
