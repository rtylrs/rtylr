package domain

import "time"

// Section is one content block on a careers page (perks, culture,
// testimonial, video, FAQ, values, CTA). Rendered server-side into
// the public payload so the frontend doesn't need separate requests
// per kind.
//
// ItemsJSON is a pre-serialised blob for kinds whose content is a
// list of objects (perks with icons, FAQ entries, etc.). Plain-text
// kinds leave it empty and use Body.
type Section struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Kind       string
	Title      string
	Body       string
	MediaURL   string
	ItemsJSON  string
	Position   int32
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SectionInput is the write shape. ID is zero for create, non-zero
// for update. Items arrive as a pre-serialised string because the
// admin UI already has it in that shape; keeping the repository's
// SQL surface simple is worth more than an extra marshal round-trip.
type SectionInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Kind       string
	Title      string
	Body       string
	MediaURL   string
	ItemsJSON  string
	Position   int32
	Enabled    bool
}
