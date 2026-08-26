package handlers

import (
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	crmcampaignv1 "github.com/voxire/voxie/gen/go/crm/campaign/v1"
)

func newCampaignPayload(c domain.Campaign) *crmcampaignv1.Campaign {
	p := &crmcampaignv1.Campaign{
		Id:             c.ID,
		OwnerId:        c.OwnerID,
		Name:           c.Name,
		Description:    c.Description,
		Channel:        c.Channel,
		Status:         c.Status,
		SegmentId:      c.SegmentID,
		Subject:        c.Subject,
		Content:        c.Content,
		AudienceCount:  int64(c.AudienceCount),
		SentCount:      int64(c.SentCount),
		DeliveredCount: int64(c.DeliveredCount),
		OpenedCount:    int64(c.OpenedCount),
		ClickedCount:   int64(c.ClickedCount),
		CreatedAt:      c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if c.ScheduledAt != nil {
		s := c.ScheduledAt.Format("2006-01-02T15:04:05Z")
		p.ScheduledAt = &s
	}

	if c.SentAt != nil {
		s := c.SentAt.Format("2006-01-02T15:04:05Z")
		p.SentAt = &s
	}

	return p
}
