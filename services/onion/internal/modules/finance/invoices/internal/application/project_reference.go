package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/platform/logger"
)

// Sentinels the HTTP layer maps to status codes, mirroring the customer link:
// a project outside the business is a caller-fixable 404, and an unwired port
// is a server misconfiguration callers cannot fix.
var (
	ErrProjectVerificationUnavailable = errors.New("project verification unavailable")
	ErrProjectNotFound                = errors.New("project not found")
)

// ProjectReferenceReader is the port invoices use to verify that a Work
// project belongs to the caller's business before tagging an invoice with it.
// Implemented by the flow references reader; neither module imports the other.
type ProjectReferenceReader interface {
	ProjectExists(ctx context.Context, businessID, projectID int64) (bool, error)
}

// SetProjectReferenceReader wires the flow implementation.
func (s *Service) SetProjectReferenceReader(r ProjectReferenceReader) {
	if r == nil {
		return
	}
	s.projectReferences = r
}

// verifyProject accepts no project (0) and otherwise fails loudly: an invoice
// must never carry a project_id this server could not verify belongs to the
// business, or another tenant's project would collect its totals.
func (s *Service) verifyProject(ctx context.Context, businessID, projectID int64) error {
	if projectID == 0 {
		return nil
	}
	if projectID < 0 {
		return fmt.Errorf("%w: %d", ErrProjectNotFound, projectID)
	}
	if s.projectReferences == nil {
		return ErrProjectVerificationUnavailable
	}
	exists, err := s.projectReferences.ProjectExists(ctx, businessID, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("project reference lookup failed",
			logger.Any("business_id", businessID), logger.Any("project_id", projectID), logger.Error(err))
		return err
	}
	if !exists {
		return fmt.Errorf("%w: %d", ErrProjectNotFound, projectID)
	}
	return nil
}

// projectLink converts a verified, non-zero project id into the stored link.
func projectLink(projectID int64) *int64 {
	if projectID <= 0 {
		return nil
	}
	return &projectID
}
