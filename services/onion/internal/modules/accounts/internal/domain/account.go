package domain

import (
	"errors"
	"time"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrNegativeBalance = errors.New("account balance cannot be negative")
)

type Account struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	ParentID   int64     `json:"parent_id"`
	Type       string    `json:"type"`
	Currency   string    `json:"currency"`
	Balance    float64   `json:"balance"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// AccountTreeNode is a single node in the chart-of-accounts tree.
type AccountTreeNode struct {
	ID       int64              `json:"id"`
	Name     string             `json:"name"`
	Code     string             `json:"code"`
	Type     string             `json:"type"`
	Children []*AccountTreeNode `json:"children"`
}
type AccountInput struct {
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	IsActive bool    `json:"is_active"`
}
