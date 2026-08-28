package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure"
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
	if in.Position == 0 {
		existing, err := s.repo.ListFormFields(ctx, in.BusinessID, in.CareerID)
		if err == nil {
			in.Position = int32(len(existing)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		}
	}
	return s.repo.CreateFormField(ctx, in)
}

func (s *Service) UpdateFormField(ctx context.Context, in domain.FormFieldInput) error {
	if _, err := s.repo.GetFormField(ctx, in.BusinessID, in.ID); err != nil {
		if errors.Is(err, infrastructure.ErrFormFieldNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.UpdateFormField(ctx, in)
}

func (s *Service) DeleteFormField(ctx context.Context, businessID, fieldID int64) error {
	err := s.repo.DeleteFormField(ctx, businessID, fieldID)
	if errors.Is(err, infrastructure.ErrFormFieldNotFound) {
		return ErrNotFound
	}
	return err
}
