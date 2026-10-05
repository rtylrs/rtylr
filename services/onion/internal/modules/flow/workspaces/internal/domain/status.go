package domain

import (
	"errors"
	"strings"
)

// F-A: per-project custom statuses.
//
// A project with no statuses is the normal case, not an unconfigured one: it
// falls back to the six built-in string statuses. That is what makes the whole
// feature additive, so nothing here may treat an empty set as an error.

var (
	ErrInvalidStatusBucket = errors.New("bucket must be one of backlog, active, done, cancelled")
	ErrInvalidLegacyStatus = errors.New("legacy_status must be one of backlog, todo, in_progress, in_review, done, cancelled")
	ErrStatusNameRequired  = errors.New("status name is required")
	ErrStatusNameTaken     = errors.New("a status with that name already exists in this project")
	ErrStatusNotFound      = errors.New("status not found")
)

// The four semantic buckets. They exist so a rollup can count across projects
// whose pipelines have nothing in common.
const (
	StatusBucketBacklog   = "backlog"
	StatusBucketActive    = "active"
	StatusBucketDone      = "done"
	StatusBucketCancelled = "cancelled"
)

// The six built-in values flow_tasks.status can hold are declared in task.go
// as TaskStatus*, and reused here rather than redeclared: two lists of the same
// enum is how the frontend ended up with a hardcoded "Todo" beside an
// i18n-backed "To Do" (rtylr#756).

// Status is one entry in a project's status set.
type Status struct {
	ID           int64
	BusinessID   int64
	ProjectID    int64
	Name         string
	Bucket       string
	LegacyStatus string
	Color        string
	SortOrder    int32
	IsDefault    bool
	CreatedBy    int64
	CreatedAt    string
}

// StatusInput is what a create or update carries.
type StatusInput struct {
	ProjectID    int64
	Name         string
	Bucket       string
	LegacyStatus string
	Color        string
	SortOrder    int32
	SortOrderSet bool
	IsDefault    bool
}

// DefaultStatusColor matches the schema default, so a caller that omits the
// colour and a row inserted without one agree.
const DefaultStatusColor = "#6366f1"

// ValidStatusBucket reports whether b is one of the four buckets.
//
// The contract types bucket as a string because .onk has no enum, so this is the
// only thing standing between a typo and a task counted in the wrong column.
func ValidStatusBucket(b string) bool {
	switch b {
	case StatusBucketBacklog, StatusBucketActive, StatusBucketDone, StatusBucketCancelled:
		return true
	}
	return false
}

// BucketForTaskStatus maps a built-in string status onto its bucket.
//
// The mapping is not a fresh invention. apps/flow's constantsFlow.ts already
// derives OPEN, CLOSED and ACTIVE sets, and these four buckets are a refinement
// of those three: ACTIVE_STATUSES is in_progress and in_review only, so `todo`
// belongs to backlog rather than active. Every set the frontend computes today
// stays reconstructible from the buckets, which is why introducing them changes
// no behaviour.
//
// An unrecognised value returns "" rather than a default. A default is how a
// cancelled task ends up counted as open.
func BucketForTaskStatus(s string) string {
	switch s {
	case TaskStatusBacklog, TaskStatusTodo:
		return StatusBucketBacklog
	case TaskStatusInProgress, TaskStatusInReview:
		return StatusBucketActive
	case TaskStatusDone:
		return StatusBucketDone
	case TaskStatusCancelled:
		return StatusBucketCancelled
	}
	return ""
}

// LegacyStatusForBucket picks the built-in string status a custom status stands
// in for when the caller does not say.
//
// Deliberately lossy in one direction: `backlog` and `active` each cover two
// built-ins, and this returns the one that keeps existing filters honest. A
// custom backlog status reports `backlog` rather than `todo` because the bucket
// says work has not been committed, and a custom active status reports
// `in_progress` rather than `in_review` because review is the narrower claim.
func LegacyStatusForBucket(bucket string) string {
	switch bucket {
	case StatusBucketBacklog:
		return TaskStatusBacklog
	case StatusBucketActive:
		return TaskStatusInProgress
	case StatusBucketDone:
		return TaskStatusDone
	case StatusBucketCancelled:
		return TaskStatusCancelled
	}
	return ""
}

// NormalizeStatusInput validates and fills in an input.
//
// Returns the error rather than silently correcting, because a silently
// corrected bucket is a task filed in the wrong column with nothing to show the
// user why.
func NormalizeStatusInput(in StatusInput) (StatusInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, ErrStatusNameRequired
	}
	if !ValidStatusBucket(in.Bucket) {
		return in, ErrInvalidStatusBucket
	}
	if in.LegacyStatus == "" {
		in.LegacyStatus = LegacyStatusForBucket(in.Bucket)
	}
	if !IsValidTaskStatus(in.LegacyStatus) {
		return in, ErrInvalidLegacyStatus
	}
	if strings.TrimSpace(in.Color) == "" {
		in.Color = DefaultStatusColor
	}
	return in, nil
}
