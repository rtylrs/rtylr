package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyCampaignName     = errors.New("campaign name is required")
	ErrEmptyCampaignType     = errors.New("campaign type is required")
	ErrInvalidCampaignID     = errors.New("campaign id is required")
	ErrInvalidCampaignOwner  = errors.New("owner_id is required")
	ErrInvalidCampaignStatus = errors.New("invalid campaign status")
)

// ValidateCampaignStatus rejects values that are not supported by the
// crm_campaigns status enum.
func ValidateCampaignStatus(status string) error {
	switch status {
	case StatusDraft, StatusScheduled, StatusSent, StatusActive, StatusPaused, StatusCompleted, StatusCancelled:
		return nil
	default:
		return ErrInvalidCampaignStatus
	}
}

// ValidateCampaignInput checks that required fields are present on campaign creation.
func ValidateCampaignInput(input CampaignInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrEmptyCampaignName
	}
	if strings.TrimSpace(input.Type) == "" {
		return ErrEmptyCampaignType
	}
	return nil
}

// ValidateCampaignUpdateInput checks that required fields are present on campaign update.
func ValidateCampaignUpdateInput(input CampaignUpdateInput) error {
	if input.ID <= 0 {
		return ErrInvalidCampaignID
	}
	if input.OwnerID <= 0 {
		return ErrInvalidCampaignOwner
	}
	if strings.TrimSpace(input.Name) == "" {
		return ErrEmptyCampaignName
	}
	return nil
}
