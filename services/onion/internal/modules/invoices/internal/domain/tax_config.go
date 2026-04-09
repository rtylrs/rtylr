package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Rate      float64   `json:"rate"`
	Type      string    `json:"type"`
	IsDefault bool      `json:"isDefault"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

func (t *TaxConfig) FromRow(row *sql.Rows) error {
	var taxType sql.NullString

	err := row.Scan(
		&t.ID,
		&t.OwnerID,
		&t.Name,
		&t.Rate,
		&taxType,
		&t.IsDefault,
		&t.IsActive,
		&t.CreatedAt,
	)
	if err != nil {
		return err
	}

	if taxType.Valid {
		t.Type = taxType.String
	} else {
		t.Type = "vat"
	}

	return nil
}

type TaxConfigInput struct {
	Name      string  `json:"name"`
	Rate      float64 `json:"rate"`
	Type      string  `json:"type"`
	IsDefault bool    `json:"isDefault"`
}
