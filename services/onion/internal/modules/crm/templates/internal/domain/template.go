package domain

import (
	"errors"
	"time"
)

var ErrTemplateNotFound = errors.New("template not found")

type Template struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Type       string
	Subject    string
	Body       string
	Variables  []byte
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
