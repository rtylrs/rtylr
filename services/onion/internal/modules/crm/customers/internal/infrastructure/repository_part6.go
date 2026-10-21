package infrastructure

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/crm/campaigns"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/crm/segments"
	"github.com/voxire/onion/internal/platform/logger"
)

func customerFromRow(row sqlc.CrmCustomer) domain.Customer {
	c := domain.Customer{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		FirstName:     row.FirstName,
		LastName:      row.LastName.String,
		Email:         row.Email.String,
		Phone:         row.Phone.String,
		Gender:        string(row.Gender.CrmCustomersGender),
		Address:       row.Address.String,
		City:          row.City.String,
		Country:       row.Country.String,
		LoyaltyPoints: int(row.LoyaltyPoints.Int32),
		VisitCount:    int(row.VisitCount.Int32),
		LoyaltyTier:   string(row.LoyaltyTier.CrmCustomersLoyaltyTier),
		Notes:         row.Notes.String,
		IsActive:      row.IsActive.Bool,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	if row.Consent != nil {
		c.Consent = append([]byte(nil), row.Consent...)
	}

	spent := parseDecimal(row.TotalSpent.String)
	c.TotalSpent = spent

	if row.DateOfBirth.Valid {
		t := row.DateOfBirth.Time
		c.DateOfBirth = &t
	}

	if row.LastVisitAt.Valid {
		t := row.LastVisitAt.Time
		c.LastVisitAt = &t
	}

	if row.Tags != nil {
		if err := json.Unmarshal(row.Tags, &c.Tags); err != nil {
			logger.Global().Warn("failed to parse customer tags JSON from database",
				logger.Any("customer_id", row.ID), logger.Error(err))
		}
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c
}

func segmentFromCanonical(row segments.Segment) domain.Segment {
	return domain.Segment{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
		Rules: row.Rules, CustomerCount: row.CustomerCount, LastCalculatedAt: row.LastCalculatedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func campaignFromCanonical(row campaigns.Campaign) domain.Campaign {
	return domain.Campaign{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
		Channel: row.Channel, Status: row.Status, SegmentID: row.SegmentID, Subject: row.Subject, Content: row.Content,
		ScheduledAt: row.ScheduledAt, SentAt: row.SentAt, AudienceCount: row.AudienceCount, SentCount: row.SentCount,
		DeliveredCount: row.DeliveredCount, OpenedCount: row.OpenedCount, ClickedCount: row.ClickedCount,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func scanCustomerRaw(rows *sql.Rows) (domain.Customer, error) {
	var c domain.Customer
	var (
		lastName    sql.NullString
		email       sql.NullString
		phone       sql.NullString
		dob         sql.NullTime
		gender      sql.NullString
		address     sql.NullString
		city        sql.NullString
		country     sql.NullString
		lastVisitAt sql.NullTime
		tagsJSON    sql.NullString
		notes       sql.NullString
		consent     []byte
	)

	if err := rows.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.FirstName, &lastName, &email, &phone, &dob, &gender,
		&address, &city, &country, &c.LoyaltyPoints, &c.TotalSpent, &c.VisitCount, &lastVisitAt,
		&c.LoyaltyTier, &tagsJSON, &notes, &consent, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return domain.Customer{}, err
	}

	c.LastName = lastName.String
	c.Email = email.String
	c.Phone = phone.String

	if dob.Valid {
		t := dob.Time
		c.DateOfBirth = &t
	}

	c.Gender = gender.String
	c.Address = address.String
	c.City = city.String
	c.Country = country.String

	if lastVisitAt.Valid {
		t := lastVisitAt.Time
		c.LastVisitAt = &t
	}

	c.Notes = notes.String
	if consent != nil {
		c.Consent = append([]byte(nil), consent...)
	}

	if tagsJSON.Valid && tagsJSON.String != "" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &c.Tags); err != nil {
			logger.Global().Warn("failed to parse customer tags JSON from database",
				logger.Any("customer_id", c.ID), logger.Error(err))
		}
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c, nil
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *t, Valid: true}
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if v.Valid {
		return &v.Int64
	}

	return nil
}

func nullInt64FromPtr(p *int64) sql.NullInt64 {
	if p != nil {
		return sql.NullInt64{Int64: *p, Valid: true}
	}

	return sql.NullInt64{}
}

func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int64:
		return int(val), true
	case int32:
		return int(val), true
	case int:
		return val, true
	case []byte:
		n, err := strconv.Atoi(string(val))
		if err == nil {
			return n, true
		}
	}

	return 0, false
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse customer decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
