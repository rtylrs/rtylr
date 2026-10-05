package domain

import (
	"time"
)

type ExpenseCategory struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Color      string
	Icon       string
	CreatedAt  time.Time
}
type ExpenseCategoryInput struct {
	Name  string
	Color string
	Icon  string
}
