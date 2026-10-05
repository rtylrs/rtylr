package infrastructure

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure/persistence/sqlc"
)

func mapNullableString(value sql.NullString) application.NullString {
	return application.NullString{String: value.String, Valid: value.Valid}
}

func nullableInt64(value sql.NullInt64) application.NullInt64 {
	return application.NullInt64{Int64: value.Int64, Valid: value.Valid}
}

func nullableTime(value sql.NullTime) application.NullTime {
	return application.NullTime{Time: value.Time, Valid: value.Valid}
}

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
		FullName:       mapNullableString(fullName),
		BusinessID:     nullableInt64(businessID),
		IsConsoleAdmin: isConsoleAdmin,
		IsBanned:       isBanned,
		BanReason:      mapNullableString(banReason),
		BannedAt:       nullableTime(bannedAt),
		CreatedAt:      nullableTime(createdAt),
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
	return &BusinessRow{ID: id, Name: name, OwnerID: ownerID, CreatedAt: nullableTime(createdAt)}
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
		OwnerName:            application.NullString{String: row.OwnerName, Valid: true},
		PlanID:               row.PlanID,
		PlanName:             row.PlanName,
		Status:               subscriptionStatusString(row.Status),
		SeatCount:            int(row.SeatCount),
		TrialEndsAt:          nullableTime(row.TrialEndsAt),
		NextBillingDate:      nullableTime(row.NextBillingDate),
		CurrentPeriodEndsAt:  nullableTime(row.CurrentPeriodEndsAt),
		PaddleSubscriptionID: mapNullableString(row.PaddleSubscriptionID),
		PaddleCustomerID:     mapNullableString(row.PaddleCustomerID),
		CreatedAt:            nullableTime(row.CreatedAt),
		UpdatedAt:            nullableTime(row.UpdatedAt),
		OwnerEmail:           mapNullableString(row.OwnerEmail),
		BusinessID:           nullableInt64(row.BusinessID),
		BusinessName:         mapNullableString(row.BusinessName),
	}
}

func subscriptionFromOwnerRow(row sqlc.GetSubscriptionByOwnerRow) *SubscriptionRow {
	return &SubscriptionRow{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		OwnerName:            application.NullString{String: row.OwnerName, Valid: true},
		PlanID:               row.PlanID,
		PlanName:             row.PlanName,
		Status:               subscriptionStatusString(row.Status),
		SeatCount:            int(row.SeatCount),
		TrialEndsAt:          nullableTime(row.TrialEndsAt),
		NextBillingDate:      nullableTime(row.NextBillingDate),
		CurrentPeriodEndsAt:  nullableTime(row.CurrentPeriodEndsAt),
		PaddleSubscriptionID: mapNullableString(row.PaddleSubscriptionID),
		PaddleCustomerID:     mapNullableString(row.PaddleCustomerID),
		CreatedAt:            nullableTime(row.CreatedAt),
		UpdatedAt:            nullableTime(row.UpdatedAt),
		OwnerEmail:           mapNullableString(row.OwnerEmail),
		BusinessID:           nullableInt64(row.BusinessID),
		BusinessName:         mapNullableString(row.BusinessName),
	}
}

func upsertSubscriptionParams(params SubscriptionWriteParams) sqlc.UpsertSubscriptionParams {
	return sqlc.UpsertSubscriptionParams{
		OwnerID:              params.OwnerID,
		PlanID:               params.PlanID,
		Status:               sqlc.NullSubscriptionsStatus{SubscriptionsStatus: sqlc.SubscriptionsStatus(params.Status), Valid: true},
		SeatCount:            int32(params.SeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		TrialEndsAt:          sql.NullTime{Time: params.TrialEndsAt.Time, Valid: params.TrialEndsAt.Valid},
		NextBillingDate:      sql.NullTime{Time: params.NextBillingDate.Time, Valid: params.NextBillingDate.Valid},
		CurrentPeriodEndsAt:  sql.NullTime{Time: params.CurrentPeriodEndsAt.Time, Valid: params.CurrentPeriodEndsAt.Valid},
		PaddleSubscriptionID: sql.NullString{String: params.PaddleSubscriptionID.String, Valid: params.PaddleSubscriptionID.Valid},
		PaddleCustomerID:     sql.NullString{String: params.PaddleCustomerID.String, Valid: params.PaddleCustomerID.Valid},
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
		CreatedAt:               nullableTime(row.CreatedAt),
	}
	if row.MaxSeats.Valid {
		p.MaxSeats = application.NullInt64{Int64: int64(row.MaxSeats.Int32), Valid: true}
	}
	if row.MaxLocations.Valid {
		p.MaxLocations = application.NullInt64{Int64: int64(row.MaxLocations.Int32), Valid: true}
	}
	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}
	return p
}
