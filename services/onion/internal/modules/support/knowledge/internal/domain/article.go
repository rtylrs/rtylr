package domain

import (
	"errors"
	"time"
)

var ErrArticleNotFound = errors.New("support article not found")

type Article struct {
	ID          int64
	BusinessID  int64
	Title       string
	Slug        string
	Body        string
	Category    string
	IsPublished bool
	Views       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ArticleInput struct {
	BusinessID  int64
	Title       string
	Slug        string
	Body        string
	Category    string
	IsPublished bool
}
