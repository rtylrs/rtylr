package domain

import "strings"

const (
	TransactionTypeAward  = "award"
	TransactionTypeRedeem = "redeem"
	TransactionTypeExpiry = "expiry"
	TransactionTypeManual = "manual"

	TierBronze   = "bronze"
	TierSilver   = "silver"
	TierGold     = "gold"
	TierPlatinum = "platinum"
)

// IsAward returns true if the transaction credits points to the customer.
func (t *LoyaltyTransaction) IsAward() bool {
	return t.Type == TransactionTypeAward
}

// IsRedeem returns true if the transaction debits points from the customer.
func (t *LoyaltyTransaction) IsRedeem() bool {
	return t.Type == TransactionTypeRedeem
}

// Initials returns the first two initials of a member's name as uppercase.
func (m *LoyaltyMember) Initials() string {
	if m.Name == "" {
		return ""
	}
	parts := strings.Fields(m.Name)
	result := ""
	for i, part := range parts {
		if i >= 2 {
			break
		}
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0]))
		}
	}
	return result
}

// IsActive returns true if the reward is available for redemption.
func (r *LoyaltyReward) IsActive() bool {
	return r.Active
}

// IsInStock returns true if the reward has stock remaining (or unlimited stock).
func (r *LoyaltyReward) IsInStock() bool {
	if r.Stock == nil {
		return true
	}
	return *r.Stock > 0
}

// ValidateAwardInput checks required fields for awarding points.
func ValidateAwardInput(input AwardInput) string {
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

// ValidateRedeemInput checks required fields for redeeming points.
func ValidateRedeemInput(input RedeemInput) string {
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

// ValidateLoyaltyConfigInput checks required fields for upserting a loyalty config.
func ValidateLoyaltyConfigInput(input LoyaltyConfigInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.MinRedemptionPoints < 0 {
		return "min_redemption_points must not be negative"
	}
	return ""
}

// TierFor returns the tier name for the given points count using this config.
func (cfg *LoyaltyConfig) TierFor(points int) string {
	if points < 0 {
		return TierBronze
	}
	if cfg.MinRedemptionPoints > 0 && points >= cfg.MinRedemptionPoints*10 {
		return TierPlatinum
	}
	return TierBronze
}

// CanRedeem returns true when the member has enough points for at least the minimum redemption.
func (cfg *LoyaltyConfig) CanRedeem(points int) bool {
	if cfg.MinRedemptionPoints <= 0 {
		return points > 0
	}
	return points >= cfg.MinRedemptionPoints
}

// HasReferenceID returns true when the transaction references an external entity.
func (t *LoyaltyTransaction) HasReferenceID() bool {
	return t.ReferenceID != nil
}
