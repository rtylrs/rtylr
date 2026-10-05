package infrastructure

import (
	"database/sql"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}
