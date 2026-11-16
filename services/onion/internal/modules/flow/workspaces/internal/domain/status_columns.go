package domain

import (
	"errors"
	"strings"
)

// A project's statuses are its board columns. These errors guard the two ways
// a delete could leave a project unusable.
var (
	ErrStatusInUse        = errors.New("status still has tasks; choose a status to move them to")
	ErrLastDoneStatus     = errors.New("a project needs at least one done status")
	ErrStatusMoveTarget   = errors.New("move target must be another status in the same project")
	ErrStatusTemplateDone = errors.New("the template needs at least one done status")
	ErrStatusTemplateSize = errors.New("the template needs between 1 and 30 statuses")
)

// StatusTemplateItem is one entry of a business's status template, the set a
// new project copies.
type StatusTemplateItem struct {
	Name         string
	Bucket       string
	LegacyStatus string
	Color        string
	SortOrder    int32
}

// DefaultStatusSet is the template a business has until it saves its own.
func DefaultStatusSet() []StatusTemplateItem {
	return []StatusTemplateItem{
		{Name: "Backlog", Bucket: StatusBucketBacklog, LegacyStatus: TaskStatusBacklog, Color: "#94a3b8", SortOrder: 0},
		{Name: "To Do", Bucket: StatusBucketBacklog, LegacyStatus: TaskStatusTodo, Color: "#9ca3af", SortOrder: 10},
		{Name: "In Progress", Bucket: StatusBucketActive, LegacyStatus: TaskStatusInProgress, Color: "#3b82f6", SortOrder: 20},
		{Name: "In Review", Bucket: StatusBucketActive, LegacyStatus: TaskStatusInReview, Color: "#f59e0b", SortOrder: 30},
		{Name: "Done", Bucket: StatusBucketDone, LegacyStatus: TaskStatusDone, Color: "#22c55e", SortOrder: 900},
		{Name: "Cancelled", Bucket: StatusBucketCancelled, LegacyStatus: TaskStatusCancelled, Color: "#ef4444", SortOrder: 1000},
	}
}

// NormalizeStatusTemplate validates a template before it replaces the stored
// one. Positions are rewritten to the submitted order so the template never
// carries gaps or ties.
func NormalizeStatusTemplate(items []StatusTemplateItem) ([]StatusTemplateItem, error) {
	if len(items) == 0 || len(items) > 30 {
		return nil, ErrStatusTemplateSize
	}
	seen := make(map[string]bool, len(items))
	out := make([]StatusTemplateItem, 0, len(items))
	hasDone := false
	for i, item := range items {
		in, err := NormalizeStatusInput(StatusInput{
			Name: item.Name, Bucket: item.Bucket, LegacyStatus: item.LegacyStatus, Color: item.Color,
		})
		if err != nil {
			return nil, err
		}
		key := strings.ToLower(in.Name)
		if seen[key] {
			return nil, ErrStatusNameTaken
		}
		seen[key] = true
		if in.Bucket == StatusBucketDone {
			hasDone = true
		}
		out = append(out, StatusTemplateItem{
			Name: in.Name, Bucket: in.Bucket, LegacyStatus: in.LegacyStatus, Color: in.Color,
			SortOrder: int32(i * 10), // #nosec G115 -- bounded by the 30-entry limit above.
		})
	}
	if !hasDone {
		return nil, ErrStatusTemplateDone
	}
	return out, nil
}
