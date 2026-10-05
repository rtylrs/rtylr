package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
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
	// When position isn't specified we append after the highest existing
	// position. Counting the rows (len) is not the same thing: positions
	// are sparse because deleted sections leave gaps, so len(existing)
	// can land the new row mid-list.
	if in.Position == 0 {
		existing, err := s.repo.ListSections(ctx, in.BusinessID, false)
		if err != nil {
			return 0, err
		}
		var next int32
		for _, section := range existing {
			if section.Position >= next {
				next = section.Position + 1
			}
		}
		in.Position = next
	}
	return s.repo.CreateSection(ctx, in)
}

// GetSection returns one section. Handlers use it to echo the row an insert
// produced instead of re-deriving it from the ordered list.
func (s *Service) GetSection(ctx context.Context, businessID, sectionID int64) (domain.Section, error) {
	section, err := s.repo.GetSection(ctx, businessID, sectionID)
	if errors.Is(err, domain.ErrSectionNotFound) {
		return domain.Section{}, ErrNotFound
	}
	return section, err
}

func (s *Service) UpdateSection(ctx context.Context, in domain.SectionInput) error {
	if _, err := s.repo.GetSection(ctx, in.BusinessID, in.ID); err != nil {
		if errors.Is(err, domain.ErrSectionNotFound) {
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
	if errors.Is(err, domain.ErrSectionNotFound) {
		return ErrNotFound
	}
	return err
}
