package domain

import "strings"

// FullName returns the customer's full name.
func (c *Customer) FullName() string {
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

// HasEmail returns true if the customer has an email address.
func (c *Customer) HasEmail() bool {
	return c.Email != ""
}

// HasPhone returns true if the customer has a phone number.
func (c *Customer) HasPhone() bool {
	return c.Phone != ""
}

// HasLoyaltyPoints returns true if the customer has any loyalty points.
func (c *Customer) HasLoyaltyPoints() bool {
	return c.LoyaltyPoints > 0
}

// IsHighValue returns true if the customer has spent above the given threshold.
func (c *Customer) IsHighValue(threshold float64) bool {
	return c.TotalSpent >= threshold
}

// ValidateCustomerInput checks required fields for creating a customer.
func ValidateCustomerInput(input CustomerInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.FirstName == "" && input.LastName == "" {
		return "at least one of first_name or last_name is required"
	}
	return ""
}

// ValidateCustomerUpdateInput checks required fields for updating a customer.
func ValidateCustomerUpdateInput(input CustomerUpdateInput) string {
	if input.ID <= 0 {
		return "customer id is required"
	}
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	return ""
}

// ValidateAwardPointsInput checks required fields for awarding loyalty points.
func ValidateAwardPointsInput(input AwardPointsInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.CustomerID <= 0 {
		return "customer_id is required"
	}
	if input.Points <= 0 {
		return "points must be positive"
	}
	return ""
}

// ValidateRedeemPointsInput checks required fields for redeeming loyalty points.
func ValidateRedeemPointsInput(input RedeemPointsInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.CustomerID <= 0 {
		return "customer_id is required"
	}
	if input.Points <= 0 {
		return "points must be positive"
	}
	return ""
}

// IsSentCampaign returns true when the campaign has been sent.
func (c *Campaign) IsSentCampaign() bool {
	return c.Status == "sent"
}

// IsDraftCampaign returns true when the campaign is in draft state.
func (c *Campaign) IsDraftCampaign() bool {
	return c.Status == "draft"
}

// CampaignOpenRate returns the fraction of sent messages that were opened.
func (c *Campaign) CampaignOpenRate() float64 {
	if c.SentCount == 0 {
		return 0
	}
	return float64(c.OpenedCount) / float64(c.SentCount)
}

// HasAudienceSegment returns true when the campaign targets a specific segment.
func (c *Campaign) HasAudienceSegment() bool {
	return c.SegmentID != nil
}

// LoyaltyConfigIsEnabled returns true if the loyalty program is enabled.
func (lc *LoyaltyConfig) LoyaltyConfigIsEnabled() bool {
	return lc.IsEnabled
}

// HasPointsExpiry returns true when points have a finite expiry window.
func (lc *LoyaltyConfig) HasPointsExpiry() bool {
	return lc.PointsExpiryDays > 0
}

// ValidateLoyaltyTierInput checks required fields for creating a loyalty tier.
func ValidateLoyaltyTierInput(input LoyaltyTierInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.Name == "" {
		return "tier name is required"
	}
	return ""
}

// HasPerks returns true when the loyalty tier has perks defined.
func (lt *LoyaltyTier) HasPerks() bool {
	return len(lt.Perks) > 0 && string(lt.Perks) != "null" && string(lt.Perks) != "[]"
}

// HasDiscount returns true when the loyalty tier provides a non-zero discount.
func (lt *LoyaltyTier) HasDiscount() bool {
	return lt.DiscountPct > 0
}

// IsEarned returns true when a customer has enough points to qualify for this tier.
func (lt *LoyaltyTier) IsEarned(points int) bool {
	return points >= lt.MinPoints
}
