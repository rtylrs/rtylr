package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
)

func (s *Service) ListFormFields(ctx context.Context, businessID, careerID int64) ([]domain.FormField, error) {
	return s.repo.ListFormFields(ctx, businessID, careerID)
}

// ListFormFieldsByCareer is used by the public apply flow — no owner
// check because the caller isn't authenticated; the tenant has
// already been bound to the career via slug resolution in the
// handler.
func (s *Service) ListFormFieldsByCareer(ctx context.Context, careerID int64) ([]domain.FormField, error) {
	return s.repo.ListFormFieldsByCareer(ctx, careerID)
}

func (s *Service) CreateFormField(ctx context.Context, in domain.FormFieldInput) (int64, error) {
	// A form field is public input for a career, so validate the career inside
	// the caller's tenant before writing. Without this scoped read, an attacker
	// can attach a field to another tenant's career and render it on their
	// public application page.
	if _, err := s.repo.GetCareerByID(ctx, in.BusinessID, in.CareerID); err != nil {
		if errors.Is(err, domain.ErrCareerNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	// Positions are sparse after deletes, so len(existing) can place a new
	// field in the middle of the editor. Append after the highest position and
	// propagate a list failure rather than silently writing an unsafe position.
	if in.Position == 0 {
		existing, err := s.repo.ListFormFields(ctx, in.BusinessID, in.CareerID)
		if err != nil {
			return 0, err
		}
		var next int32
		for _, field := range existing {
			if field.Position >= next {
				next = field.Position + 1
			}
		}
		in.Position = next
	}
	return s.repo.CreateFormField(ctx, in)
}

// GetFormField returns one field. Handlers use it to echo the row an insert
// produced instead of re-deriving it from the (position-ordered, capped) list.
func (s *Service) GetFormField(ctx context.Context, businessID, fieldID int64) (domain.FormField, error) {
	field, err := s.repo.GetFormField(ctx, businessID, fieldID)
	if errors.Is(err, domain.ErrFormFieldNotFound) {
		return domain.FormField{}, ErrNotFound
	}
	return field, err
}

func (s *Service) UpdateFormField(ctx context.Context, in domain.FormFieldInput) error {
	if _, err := s.repo.GetFormField(ctx, in.BusinessID, in.ID); err != nil {
		if errors.Is(err, domain.ErrFormFieldNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.UpdateFormField(ctx, in)
}

func (s *Service) DeleteFormField(ctx context.Context, businessID, fieldID int64) error {
	err := s.repo.DeleteFormField(ctx, businessID, fieldID)
	if errors.Is(err, domain.ErrFormFieldNotFound) {
		return ErrNotFound
	}
	return err
}
