package infrastructure

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure/persistence/sqlc"
)

func userFromRow(
	id int64,
	email string,
	fullName sql.NullString,
	businessID sql.NullInt64,
	isConsoleAdmin, isBanned bool,
	banReason sql.NullString,
	bannedAt, createdAt sql.NullTime,
	passwordHash string,
) *UserRow {
	return &UserRow{
		ID:             id,
		Email:          email,
		FullName:       fullName,
		BusinessID:     businessID,
		IsConsoleAdmin: isConsoleAdmin,
		IsBanned:       isBanned,
		BanReason:      banReason,
		BannedAt:       bannedAt,
		CreatedAt:      createdAt,
		PasswordHash:   passwordHash,
	}
}

func userFromGetUserByEmailRow(row sqlc.GetUserByEmailRow) *UserRow {
	return userFromRow(
		row.ID, row.Email, row.FullName, row.BusinessID,
		row.IsConsoleAdmin, row.IsBanned, row.BanReason, row.BannedAt, row.CreatedAt,
		row.PasswordHash,
	)
}

func userFromGetUserByIDRow(row sqlc.GetUserByIDRow) *UserRow {
	return userFromRow(
		row.ID, row.Email, row.FullName, row.BusinessID,
		row.IsConsoleAdmin, row.IsBanned, row.BanReason, row.BannedAt, row.CreatedAt,
		row.PasswordHash,
	)
}

func businessFromRow(id int64, name string, ownerID int64, createdAt sql.NullTime) *BusinessRow {
	return &BusinessRow{ID: id, Name: name, OwnerID: ownerID, CreatedAt: createdAt}
}

func subscriptionStatusString(status sqlc.NullSubscriptionsStatus) string {
	if status.Valid {
		return string(status.SubscriptionsStatus)
	}
	return ""
}

func subscriptionFromIDRow(row sqlc.GetSubscriptionByIDRow) *SubscriptionRow {
	return &SubscriptionRow{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		OwnerName:            sql.NullString{String: row.OwnerName, Valid: true},
		PlanID:               row.PlanID,
		PlanName:             row.PlanName,
		Status:               subscriptionStatusString(row.Status),
		SeatCount:            int(row.SeatCount),
		TrialEndsAt:          row.TrialEndsAt,
		NextBillingDate:      row.NextBillingDate,
		CurrentPeriodEndsAt:  row.CurrentPeriodEndsAt,
		PaddleSubscriptionID: row.PaddleSubscriptionID,
		PaddleCustomerID:     row.PaddleCustomerID,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		OwnerEmail:           row.OwnerEmail,
		BusinessID:           row.BusinessID,
		BusinessName:         row.BusinessName,
	}
}

func subscriptionFromOwnerRow(row sqlc.GetSubscriptionByOwnerRow) *SubscriptionRow {
	return &SubscriptionRow{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		OwnerName:            sql.NullString{String: row.OwnerName, Valid: true},
		PlanID:               row.PlanID,
		PlanName:             row.PlanName,
		Status:               subscriptionStatusString(row.Status),
		SeatCount:            int(row.SeatCount),
		TrialEndsAt:          row.TrialEndsAt,
		NextBillingDate:      row.NextBillingDate,
		CurrentPeriodEndsAt:  row.CurrentPeriodEndsAt,
		PaddleSubscriptionID: row.PaddleSubscriptionID,
		PaddleCustomerID:     row.PaddleCustomerID,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		OwnerEmail:           row.OwnerEmail,
		BusinessID:           row.BusinessID,
		BusinessName:         row.BusinessName,
	}
}

func upsertSubscriptionParams(params SubscriptionWriteParams) sqlc.UpsertSubscriptionParams {
	return sqlc.UpsertSubscriptionParams{
		OwnerID:              params.OwnerID,
		PlanID:               params.PlanID,
		Status:               sqlc.NullSubscriptionsStatus{SubscriptionsStatus: sqlc.SubscriptionsStatus(params.Status), Valid: true},
		SeatCount:            int32(params.SeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		TrialEndsAt:          params.TrialEndsAt,
		NextBillingDate:      params.NextBillingDate,
		CurrentPeriodEndsAt:  params.CurrentPeriodEndsAt,
		PaddleSubscriptionID: params.PaddleSubscriptionID,
		PaddleCustomerID:     params.PaddleCustomerID,
	}
}

func planFromListRow(row sqlc.ListPlansRow) PlanRow {
	p := PlanRow{
		ID:                      row.ID,
		Name:                    row.Name,
		Slug:                    row.Slug,
		PricePerSeatCents:       int(row.PricePerSeatCents),
		AnnualPricePerSeatCents: int(row.AnnualPricePerSeatCents),
		TrialDays:               int(row.TrialDays),
		IsSolo:                  row.IsSolo,
		IsFeatured:              row.IsFeatured,
		TierOrder:               int(row.TierOrder),
		Tagline:                 row.Tagline,
		CreatedAt:               row.CreatedAt,
	}
	if row.MaxSeats.Valid {
		p.MaxSeats = sql.NullInt64{Int64: int64(row.MaxSeats.Int32), Valid: true}
	}
	if row.MaxLocations.Valid {
		p.MaxLocations = sql.NullInt64{Int64: int64(row.MaxLocations.Int32), Valid: true}
	}
	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}
	return p
}

func selfHostLicenseFromModel(row sqlc.SelfHostLicense) *SelfHostLicenseRow {
	var features sql.NullString
	if len(row.Features) > 0 && string(row.Features) != "null" {
		features = sql.NullString{String: string(row.Features), Valid: true}
	}
	var seatLimit sql.NullInt64
	if row.SeatLimit.Valid {
		seatLimit = sql.NullInt64{Int64: int64(row.SeatLimit.Int32), Valid: true}
	}
	return &SelfHostLicenseRow{
		ID:            row.ID,
		OwnerEmail:    row.OwnerEmail,
		CustomerName:  row.CustomerName,
		Plan:          row.Plan,
		Features:      features,
		SeatLimit:     seatLimit,
		ExpiresAt:     row.ExpiresAt,
		GraceUntil:    row.GraceUntil,
		Status:        string(row.Status),
		LicenseKey:    row.LicenseKey,
		KeyHash:       row.KeyHash,
		CreatedBy:     row.CreatedBy,
		UpdatedBy:     row.UpdatedBy,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		RegeneratedAt: row.RegeneratedAt,
		RevokedAt:     row.RevokedAt,
	}
}

func featuresJSONFromString(raw string) json.RawMessage {
	if raw == "" {
		return nil
	}
	return json.RawMessage(raw)
}

func nullInt32FromInt(value int) sql.NullInt32 {
	if value <= 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(value), Valid: true} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func nullTimeFromPtr(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}
