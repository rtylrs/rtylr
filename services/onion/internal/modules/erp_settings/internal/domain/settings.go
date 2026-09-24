package domain

import (
	"errors"
	"time"
)

var ErrSettingsNotFound = errors.New("erp settings not found")

const (
	DefaultCurrency                = "USD"
	DefaultTaxRate                 = 8.25
	DefaultLowStockThreshold       = 10.0
	DefaultWarehouseID       int64 = 1
)

type Settings struct {
	ID                 int64
	BusinessID         int64
	Currency           string
	TaxRate            float64
	LowStockThreshold  float64
	DefaultWarehouseID int64
	NotifyOnLowStock   bool
	NotifyOnExpiry     bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SettingsInput struct {
	BusinessID         int64
	Currency           string
	TaxRate            float64
	LowStockThreshold  float64
	DefaultWarehouseID int64
	NotifyOnLowStock   bool
	NotifyOnExpiry     bool
}

func DefaultSettings(businessID int64) Settings {
	now := time.Now().UTC()
	return Settings{
		BusinessID:         businessID,
		Currency:           DefaultCurrency,
		TaxRate:            DefaultTaxRate,
		LowStockThreshold:  DefaultLowStockThreshold,
		DefaultWarehouseID: DefaultWarehouseID,
		NotifyOnLowStock:   true,
		NotifyOnExpiry:     true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func ValidateInput(input SettingsInput) error {
	switch {
	case input.BusinessID <= 0:
		return errors.New("business_id must be greater than 0")
	case input.Currency == "":
		return errors.New("currency is required")
	case input.TaxRate < 0:
		return errors.New("tax_rate must be greater than or equal to 0")
	case input.TaxRate > 100:
		return errors.New("tax_rate must be less than or equal to 100")
	case input.LowStockThreshold < 0:
		return errors.New("low_stock_threshold must be greater than or equal to 0")
	case input.DefaultWarehouseID <= 0:
		return errors.New("default_warehouse_id must be greater than 0")
	default:
		return nil
	}
}

func (s Settings) Validate() error {
	return ValidateInput(SettingsInput{
		BusinessID:         s.BusinessID,
		Currency:           s.Currency,
		TaxRate:            s.TaxRate,
		LowStockThreshold:  s.LowStockThreshold,
		DefaultWarehouseID: s.DefaultWarehouseID,
		NotifyOnLowStock:   s.NotifyOnLowStock,
		NotifyOnExpiry:     s.NotifyOnExpiry,
	})
}
