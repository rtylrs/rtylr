package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrBrandNotFound = errors.New("brand not found")

type Brand struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logoUrl"`
	Description  string    `json:"description"`
	ProductCount int64     `json:"productCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (b *Brand) FromRow(row *sql.Rows) error {
	var (
		logoURL      sql.NullString
		description  sql.NullString
		productCount sql.NullInt64
	)

	err := row.Scan(
		&b.ID,
		&b.OwnerID,
		&b.Name,
		&logoURL,
		&description,
		&productCount,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if logoURL.Valid {
		b.LogoURL = logoURL.String
	}

	if description.Valid {
		b.Description = description.String
	}

	if productCount.Valid {
		b.ProductCount = productCount.Int64
	}

	return nil
}

type BrandInput struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logoUrl"`
	Description string `json:"description"`
}
