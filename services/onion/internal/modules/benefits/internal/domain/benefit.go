package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrBenefitPlanNotFound = errors.New("benefit plan not found")

type BenefitPlan struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Provider     string    `json:"provider"`
	CostEmployer float64   `json:"costEmployer"`
	CostEmployee float64   `json:"costEmployee"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (b *BenefitPlan) FromRow(row *sql.Rows) error {
	var (
		benefitType sql.NullString
		provider    sql.NullString
	)

	err := row.Scan(
		&b.ID, &b.OwnerID, &b.Name, &benefitType, &provider,
		&b.CostEmployer, &b.CostEmployee, &b.IsActive, &b.CreatedAt,
	)
	if err != nil {
		return err
	}

	if benefitType.Valid {
		b.Type = benefitType.String
	}
	if provider.Valid {
		b.Provider = provider.String
	}

	return nil
}

type BenefitPlanInput struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Provider     string  `json:"provider"`
	CostEmployer float64 `json:"costEmployer"`
	CostEmployee float64 `json:"costEmployee"`
	IsActive     bool    `json:"isActive"`
}
