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
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Code       string
	ParentID   int64
	Type       string
	Currency   string
	Balance    float64
	IsActive   bool
	CreatedAt  time.Time
}

// AccountTreeNode is a single node in the chart-of-accounts tree.
type AccountTreeNode struct {
	ID       int64
	Name     string
	Code     string
	Type     string
	Children []*AccountTreeNode
}
type AccountInput struct {
	Name     string
	Code     string
	Type     string
	Currency string
	Balance  float64
	IsActive bool
}
