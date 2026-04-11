package domain

import "time"

type Article struct {
	ID          int64     `json:"id"`
	BusinessID  int64     `json:"businessId"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	IsPublished bool      `json:"isPublished"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ArticleInput struct {
	BusinessID  int64
	Title       string
	Slug        string
	Body        string
	Category    string
	IsPublished bool
}
