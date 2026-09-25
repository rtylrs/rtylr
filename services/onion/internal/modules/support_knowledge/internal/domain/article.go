package domain

import (
	"errors"
	"time"
)

var ErrArticleNotFound = errors.New("support article not found")

type Article struct {
	ID          int64     `json:"id"`
	BusinessID  int64     `json:"business_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	IsPublished bool      `json:"is_published"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ArticleInput struct {
	BusinessID  int64
	Title       string
	Slug        string
	Body        string
	Category    string
	IsPublished bool
}
