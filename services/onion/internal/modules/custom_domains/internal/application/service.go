package application

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/voxire/onion/internal/modules/custom_domains/internal/domain"
	"github.com/voxire/onion/internal/modules/custom_domains/internal/infrastructure"
)

type txtResolver interface {
	LookupTXT(context.Context, string) ([]string, error)
}

type Service struct {
	repo     *infrastructure.Repository
	resolver txtResolver
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db), resolver: net.DefaultResolver}
}

func (s *Service) Get(ctx context.Context, businessID int64) (domain.CustomDomain, error) {
	return s.repo.Get(ctx, businessID)
}

func (s *Service) Register(ctx context.Context, ownerID, businessID int64, rawDomain string) (domain.CustomDomain, error) {
	domainName := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(rawDomain)), ".")
	if err := validateDomain(domainName); err != nil {
		return domain.CustomDomain{}, err
	}
	if current, err := s.repo.GetByDomain(ctx, domainName); err == nil && current.BusinessID != businessID {
		return domain.CustomDomain{}, domain.ErrConflict
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.CustomDomain{}, err
	}
	if current, err := s.repo.Get(ctx, businessID); err == nil && current.Domain != domainName {
		return domain.CustomDomain{}, domain.ErrConflict
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.CustomDomain{}, err
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
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" || subtle.ConstantTimeCompare([]byte(trimmedToken), []byte(current.VerificationToken)) != 1 {
		return domain.CustomDomain{}, domain.ErrInvalidToken
	}
	if !s.dnsChallengeMatches(ctx, current.Domain, current.VerificationToken) {
		return domain.CustomDomain{}, domain.ErrDNSVerificationFailed
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
	if value == "" || len(value) > 253 || net.ParseIP(value) != nil {
		return fmt.Errorf("invalid custom domain")
	}
	blockedSuffixes := []string{"rtylr.com", "tryrtylr.com", "voxire.com", "voxire.dev"}
	for _, suffix := range blockedSuffixes {
		if value == suffix || strings.HasSuffix(value, "."+suffix) {
			return fmt.Errorf("invalid custom domain")
		}
	}
	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return fmt.Errorf("invalid custom domain")
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("invalid custom domain")
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
				return fmt.Errorf("invalid custom domain")
			}
		}
	}
	return nil
}

func (s *Service) dnsChallengeMatches(ctx context.Context, domainName, token string) bool {
	if s.resolver == nil {
		return false
	}
	records, err := s.resolver.LookupTXT(ctx, "_rtylr-challenge."+domainName)
	if err != nil {
		return false
	}
	for _, record := range records {
		if subtle.ConstantTimeCompare([]byte(record), []byte(token)) == 1 {
			return true
		}
	}
	return false
}
