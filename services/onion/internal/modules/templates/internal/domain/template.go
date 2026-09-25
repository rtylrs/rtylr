package domain

import (
	"errors"
	"time"
)

var ErrTemplateNotFound = errors.New("template not found")

type Template struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	Variables  []byte    `json:"variables"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TemplateInput struct {
	OwnerID    int64
	BusinessID int64
	Name       string
	Type       string
	Subject    string
	Body       string
	Variables  []byte
	IsActive   bool
}

type TemplateUpdateInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Type       string
	Subject    string
	Body       string
	Variables  []byte
	IsActive   bool
}
