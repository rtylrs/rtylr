package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrExpenseCategoryNotFound = errors.New("expense category not found")

type ExpenseCategory struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	Icon       string    `json:"icon"`
	CreatedAt  time.Time `json:"created_at"`
}

func (c *ExpenseCategory) FromRow(row *sql.Rows) error {
	var (
		color sql.NullString
		icon  sql.NullString
	)

	err := row.Scan(
		&c.ID,
		&c.OwnerID,
		&c.BusinessID,
		&c.Name,
		&color,
		&icon,
		&c.CreatedAt,
	)
	if err != nil {
		return err
	}

	if color.Valid {
		c.Color = color.String
	}

	if icon.Valid {
		c.Icon = icon.String
	}

	return nil
}

type ExpenseCategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
