package domain

import (
	"errors"
	"time"
)

var (
	ErrDocumentNotFound      = errors.New("document not found")
	ErrDocumentStaffNotFound = errors.New("document staff not found")
)

type Document struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	StaffID    int64     `json:"staff_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	FileURL    string    `json:"file_url"`
	ExpiryDate time.Time `json:"expiry_date"`
	CreatedAt  time.Time `json:"created_at"`
}
type DocumentInput struct {
	StaffID    int64  `json:"staff_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	FileURL    string `json:"file_url"`
	ExpiryDate string `json:"expiry_date"`
}
