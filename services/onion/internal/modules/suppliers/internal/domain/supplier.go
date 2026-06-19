package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrSupplierNotFound = errors.New("supplier not found")

type Supplier struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	TaxID        string    `json:"tax_id"`
	PaymentTerms string    `json:"payment_terms"`
	Notes        string    `json:"notes"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Supplier) FromRow(row *sql.Rows) error {
	var (
		email        sql.NullString
		phone        sql.NullString
		address      sql.NullString
		city         sql.NullString
		country      sql.NullString
		taxID        sql.NullString
		paymentTerms sql.NullString
		notes        sql.NullString
	)

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
		&s.Name,
		&email,
		&phone,
		&address,
		&city,
		&country,
		&taxID,
		&paymentTerms,
		&notes,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if email.Valid {
		s.Email = email.String
	}
	if phone.Valid {
		s.Phone = phone.String
	}
	if address.Valid {
		s.Address = address.String
	}
	if city.Valid {
		s.City = city.String
	}
	if country.Valid {
		s.Country = country.String
	}
	if taxID.Valid {
		s.TaxID = taxID.String
	}
	if paymentTerms.Valid {
		s.PaymentTerms = paymentTerms.String
	}
	if notes.Valid {
		s.Notes = notes.String
	}

	return nil
}

type SupplierInput struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Country      string `json:"country"`
	TaxID        string `json:"tax_id"`
	PaymentTerms string `json:"payment_terms"`
	Notes        string `json:"notes"`
	IsActive     bool   `json:"is_active"`
}
