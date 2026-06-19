package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrCategoryNotFound = errors.New("category not found")

type Category struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *Category) FromRow(row *sql.Rows) error {
	var description sql.NullString

	err := row.Scan(
		&c.ID,
		&c.OwnerID,
		&c.BusinessID,
		&c.Name,
		&description,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if description.Valid {
		c.Description = description.String
	}

	return nil
}

type CategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
