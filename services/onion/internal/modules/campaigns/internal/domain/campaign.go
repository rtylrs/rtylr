package domain

import (
	"errors"
	"time"
)

var ErrCampaignNotFound = errors.New("campaign not found")

type Campaign struct {
	ID                int64      `json:"id"`
	OwnerID           int64      `json:"owner_id"`
	BusinessID        int64      `json:"business_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Type              string     `json:"type"`
	Status            string     `json:"status"`
	AudienceSegmentID *int64     `json:"audience_segment_id"`
	Subject           string     `json:"subject"`
	Body              string     `json:"body"`
	AudienceCount     int        `json:"audience_count"`
	SentCount         int        `json:"sent_count"`
	DeliveredCount    int        `json:"delivered_count"`
	OpenCount         int        `json:"open_count"`
	ClickedCount      int        `json:"clicked_count"`
	ScheduledAt       *time.Time `json:"scheduled_at"`
	SentAt            *time.Time `json:"sent_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
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
