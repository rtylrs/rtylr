package domain

import (
	"time"
)

type ExpenseCategory struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	Icon       string    `json:"icon"`
	CreatedAt  time.Time `json:"created_at"`
}
type ExpenseCategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
