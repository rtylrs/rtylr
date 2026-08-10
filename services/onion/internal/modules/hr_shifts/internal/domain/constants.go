package domain

// Shift assignment status values.
const (
	AssignmentStatusScheduled = "scheduled"
	AssignmentStatusConfirmed = "confirmed"
	AssignmentStatusSwapped   = "swapped"
	AssignmentStatusCanceled  = "canceled"
)

// DefaultShiftColor and DefaultShiftStatus mirror the column defaults in
// schemas/hr_shifts/schema.sql. These are product-level defaults for a row that
// omits them, not tenant-configurable policy: the colour is a roster rendering
// detail and the status is the first state of a fixed lifecycle, so neither
// differs between one business and another.
const (
	DefaultShiftColor  = "#4CAF50"
	DefaultShiftStatus = "scheduled"
)
