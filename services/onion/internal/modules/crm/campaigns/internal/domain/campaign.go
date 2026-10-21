package domain

import (
	"errors"
	"time"
)

var ErrCampaignNotFound = errors.New("campaign not found")

type Campaign struct {
	ID                int64
	OwnerID           int64
	BusinessID        int64
	Name              string
	Description       string
	Type              string
	Status            string
	AudienceSegmentID *int64
	Subject           string
	Body              string
	AudienceCount     int
	SentCount         int
	DeliveredCount    int
	OpenCount         int
	ClickedCount      int
	ScheduledAt       *time.Time
	SentAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CampaignPage is a single page of campaigns plus the pagination metadata
// that was applied to produce it.
type CampaignPage struct {
	Campaigns []Campaign
	Total     int64
	Limit     int32
	Offset    int32
}

type CampaignInput struct {
	OwnerID           int64
	BusinessID        int64
	Name              string
	Description       string
	Type              string
	AudienceSegmentID *int64
	Subject           string
	Body              string
	ScheduledAt       *time.Time
}

type CampaignUpdateInput struct {
	ID                int64
	OwnerID           int64
	BusinessID        int64
	Name              string
	Description       string
	Type              string
	AudienceSegmentID *int64
	Subject           string
	Body              string
	ScheduledAt       *time.Time
}
