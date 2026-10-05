package domain

import (
	"errors"
	"time"
)

var ErrBrandNotFound = errors.New("brand not found")

type Brand struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	LogoURL      string
	Description  string
	ProductCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type BrandInput struct {
	Name        string
	LogoURL     string
	Description string
}
