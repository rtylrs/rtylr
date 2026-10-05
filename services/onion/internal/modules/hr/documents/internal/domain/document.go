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
	ID         int64
	OwnerID    int64
	BusinessID int64
	StaffID    int64
	Name       string
	Type       string
	FileURL    string
	ExpiryDate time.Time
	CreatedAt  time.Time
}
type DocumentInput struct {
	StaffID    int64
	Name       string
	Type       string
	FileURL    string
	ExpiryDate string
}
