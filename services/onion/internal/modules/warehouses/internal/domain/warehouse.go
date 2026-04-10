package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrWarehouseNotFound = errors.New("warehouse not found")

type Warehouse struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	ManagerName string    `json:"managerName"`
	Capacity    int64     `json:"capacity"`
	IsDefault   bool      `json:"isDefault"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (w *Warehouse) FromRow(row *sql.Rows) error {
	var (
		address     sql.NullString
		city        sql.NullString
		country     sql.NullString
		managerName sql.NullString
		capacity    sql.NullInt64
	)

	err := row.Scan(
		&w.ID,
		&w.OwnerID,
		&w.Name,
		&address,
		&city,
		&country,
		&managerName,
		&capacity,
		&w.IsDefault,
		&w.IsActive,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if address.Valid {
		w.Address = address.String
	}

	if city.Valid {
		w.City = city.String
	}

	if country.Valid {
		w.Country = country.String
	}

	if managerName.Valid {
		w.ManagerName = managerName.String
	}

	if capacity.Valid {
		w.Capacity = capacity.Int64
	}

	return nil
}

type WarehouseInput struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	ManagerName string `json:"managerName"`
	Capacity    int64  `json:"capacity"`
	IsDefault   bool   `json:"isDefault"`
	IsActive    bool   `json:"isActive"`
}
