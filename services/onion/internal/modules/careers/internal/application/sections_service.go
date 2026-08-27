package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure"
)

// Sections service — wraps the repository with permission-agnostic
// business methods. Handlers translate sentinel errors into HTTP
// codes; callers don't touch sql or infrastructure types.

// ListSections returns all of a tenant's sections. `enabledOnly=true`
// is used by the public page load so draft sections stay hidden.
func (s *Service) ListSections(ctx context.Context, businessID int64, enabledOnly bool) ([]domain.Section, error) {
	return s.repo.ListSections(ctx, businessID, enabledOnly)
}

func (s *Service) CreateSection(ctx context.Context, in domain.SectionInput) (int64, error) {
	// When position isn't specified we append — count the existing
	// sections and use that as the new row's position. Callers that
	// want a specific slot pass their own position.
	if in.Position == 0 {
		existing, err := s.repo.ListSections(ctx, in.BusinessID, false)
		if err == nil {
			in.Position = int32(len(existing)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		}
	}
	return s.repo.CreateSection(ctx, in)
}

func (s *Service) UpdateSection(ctx context.Context, in domain.SectionInput) error {
	if _, err := s.repo.GetSection(ctx, in.BusinessID, in.ID); err != nil {
		if errors.Is(err, infrastructure.ErrSectionNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.UpdateSection(ctx, in)
}

func (s *Service) ReorderSections(ctx context.Context, businessID int64, orderedIDs []int64) error {
	return s.repo.ReorderSections(ctx, businessID, orderedIDs)
}

func (s *Service) DeleteSection(ctx context.Context, businessID, sectionID int64) error {
	err := s.repo.DeleteSection(ctx, businessID, sectionID)
	if errors.Is(err, infrastructure.ErrSectionNotFound) {
		return ErrNotFound
	}
	return err
}
