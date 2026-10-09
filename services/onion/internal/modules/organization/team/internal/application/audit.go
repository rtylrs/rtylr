package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/platform/logger"
)

type auditRecorder interface {
	RecordAudit(ctx context.Context, event audit.Event) error
}

func (s *Service) WithAuditRecorder(recorder auditRecorder) *Service {
	s.audit = recorder
	return s
}

// recordAudit adds a team change to the activity log once it has committed. A
// failure is logged rather than returned: the change itself already happened.
func (s *Service) recordAudit(ctx context.Context, businessID, actorID int64, entityType string, entityID int64, action, description string) {
	if s.audit == nil || businessID <= 0 || entityID <= 0 {
		return
	}
	if err := s.audit.RecordAudit(ctx, audit.Event{
		BusinessID: businessID, ActorID: actorID, EntityType: entityType, EntityID: entityID,
		Action: action, Description: description,
	}); err != nil {
		logger.Ctx(ctx).Error("failed to record team audit entry",
			logger.String("action", action), logger.Int("business_id", int(businessID)), logger.Error(err))
	}
}

func memberLabel(name, email string) string {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)
	switch {
	case name != "" && email != "":
		return fmt.Sprintf("%s (%s)", name, email)
	case name != "":
		return name
	case email != "":
		return email
	default:
		return "a member"
	}
}
