package domain

import (
	"errors"
	"time"
)

var ErrCampaignNotFound = errors.New("campaign not found")

type Campaign struct {
	ID                int64      `json:"id"`
	OwnerID           int64      `json:"ownerId"`
	Name              string     `json:"name"`
	Type              string     `json:"type"`
	Status            string     `json:"status"`
	AudienceSegmentID *int64     `json:"audienceSegmentId"`
	Subject           string     `json:"subject"`
	Body              string     `json:"body"`
	SentCount         int        `json:"sentCount"`
	OpenCount         int        `json:"openCount"`
	ScheduledAt       *time.Time `json:"scheduledAt"`
	SentAt            *time.Time `json:"sentAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type CampaignInput struct {
	OwnerID           int64
	Name              string
	Type              string
	AudienceSegmentID *int64
	Subject           string
	Body              string
	ScheduledAt       *time.Time
}

type CampaignUpdateInput struct {
	ID                int64
	OwnerID           int64
	Name              string
	Type              string
	AudienceSegmentID *int64
	Subject           string
	Body              string
	ScheduledAt       *time.Time
}
