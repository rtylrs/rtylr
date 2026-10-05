package domain

import "time"

// FormField is a custom question a business asks candidates applying
// for a specific role. The baseline (name/email/phone/resume/cover
// letter) is fixed — these extras are per-career.
//
// OptionsJSON is a JSON-encoded string array for single_select and
// multi_select kinds; empty for other kinds. Kept as a string because
// the admin UI already serialises it in that shape.
type FormField struct {
	ID          int64
	CareerID    int64
	OwnerID     int64
	BusinessID  int64
	Kind        string
	Label       string
	Placeholder string
	Required    bool
	OptionsJSON string
	Position    int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FormFieldInput is the write shape. Reused by create and update; ID
// is zero on create, non-zero on update.
type FormFieldInput struct {
	ID          int64
	CareerID    int64
	OwnerID     int64
	BusinessID  int64
	Kind        string
	Label       string
	Placeholder string
	Required    bool
	OptionsJSON string
	Position    int32
}
