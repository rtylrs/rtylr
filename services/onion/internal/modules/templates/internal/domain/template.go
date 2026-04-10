package domain

import (
	"errors"
	"time"
)

var ErrTemplateNotFound = errors.New("template not found")

type Template struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Variables []byte    `json:"variables"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TemplateInput struct {
	OwnerID   int64
	Name      string
	Type      string
	Subject   string
	Body      string
	Variables []byte
	IsActive  bool
}

type TemplateUpdateInput struct {
	ID        int64
	OwnerID   int64
	Name      string
	Type      string
	Subject   string
	Body      string
	Variables []byte
	IsActive  bool
}
