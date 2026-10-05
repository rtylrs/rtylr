package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) listShortlistItems(ctx context.Context, shortlistID int64) ([]domain.ShortlistItem, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listShortlistItemsScopedSQL, ownerID, shortlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.ShortlistItem, 0)
	for rows.Next() {
		var row shortlistItemRecord
		if err := rows.Scan(&row.ID, &row.ShortlistID, &row.ApplicationID, &row.CandidateName,
			&row.CandidateEmail, &row.Status, &row.SortOrder, &row.Notes, &row.CvProfileJson,
			&row.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, shortlistItemFromRow(row))
	}
	return items, rows.Err()
}

func (r *Repository) getShortlistItemByID(ctx context.Context, id int64) (*domain.ShortlistItem, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var row shortlistItemRecord
	err = r.db.QueryRowContext(ctx, getShortlistItemScopedSQL, ownerID, id).Scan(
		&row.ID, &row.ShortlistID, &row.ApplicationID, &row.CandidateName, &row.CandidateEmail,
		&row.Status, &row.SortOrder, &row.Notes, &row.CvProfileJson, &row.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	item := domain.ShortlistItem{
		ID: row.ID, ShortlistID: row.ShortlistID, ApplicationID: row.ApplicationID,
		CandidateName: row.CandidateName, CandidateEmail: row.CandidateEmail, Status: row.Status,
		Rank: int(row.SortOrder), Notes: row.Notes, CreatedAt: adminCreatedAt(row.CreatedAt),
		CVProfile: parseCVProfileJSON(row.CvProfileJson),
	}
	return &item, nil
}

func (r *Repository) ListApplicationTimeline(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var meta struct {
		CreatedAt, CvProcessedAt, ReviewedAt  sql.NullTime
		CvProcessingStatus, CvProcessingError string
	}
	err = r.db.QueryRowContext(ctx, getApplicationTimelineMetaScopedSQL, ownerID, id).Scan(
		&meta.CreatedAt, &meta.CvProcessingStatus, &meta.CvProcessingError,
		&meta.CvProcessedAt, &meta.ReviewedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}

	events := []domain.TimelineEvent{{
		Kind:      "application",
		Title:     "Application submitted",
		CreatedAt: adminCreatedAt(meta.CreatedAt),
	}}
	if meta.CvProcessedAt.Valid {
		title := "CV processed"
		if meta.CvProcessingStatus == "error" {
			title = "CV extraction failed"
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "cv",
			Title:     title,
			Detail:    meta.CvProcessingError,
			CreatedAt: meta.CvProcessedAt.Time,
		})
	}
	if meta.ReviewedAt.Valid {
		events = append(events, domain.TimelineEvent{
			Kind:      "review",
			Title:     "Recruiter review saved",
			CreatedAt: meta.ReviewedAt.Time,
		})
	}
	extra, err := r.listRecordedApplicationEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, extra...)
	taskEvents, err := r.listTaskTimelineEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, taskEvents...)
	shortlistEvents, err := r.listShortlistTimelineEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, shortlistEvents...)
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].CreatedAt.After(events[j].CreatedAt)
	})
	return events, nil
}

func (r *Repository) listRecordedApplicationEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listApplicationEventsScopedSQL, ownerID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var row struct {
			Kind, Title, Detail string
			CreatedAt           sql.NullTime
		}
		if err := rows.Scan(&row.Kind, &row.Title, &row.Detail, &row.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{
			Kind: row.Kind, Title: row.Title, Detail: row.Detail, CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}
	return events, rows.Err()
}

func (r *Repository) listTaskTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTaskTimelineEvents(ctx, sqlc.ListTaskTimelineEventsParams{
		OwnerID: ownerID, ApplicationID: id,
	})
	if err != nil {
		return nil, err
	}
	events := make([]domain.TimelineEvent, 0)
	for _, row := range rows {
		events = append(events, domain.TimelineEvent{
			Kind: "task", Title: "Task opened", Detail: row.Title, CreatedAt: adminCreatedAt(row.CreatedAt),
		})
		if row.CompletedAt.Valid {
			events = append(events, domain.TimelineEvent{
				Kind: "task", Title: "Task completed", Detail: row.Title, CreatedAt: row.CompletedAt.Time,
			})
		} else if row.Status != "open" {
			events[len(events)-1].Title = "Task " + row.Status
		}
	}
	return events, nil
}

func (r *Repository) listShortlistTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listShortlistTimelineEventsScopedSQL, ownerID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var row struct {
			Name      string
			SortOrder int32
			CreatedAt sql.NullTime
		}
		if err := rows.Scan(&row.Name, &row.SortOrder, &row.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "shortlist",
			Title:     "Added to shortlist",
			Detail:    fmt.Sprintf("%s · rank %d", row.Name, row.SortOrder),
			CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}
	return events, rows.Err()
}

func (r *Repository) applicationExists(ctx context.Context, id int64) (bool, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.QueryRowContext(ctx, applicationExistsScopedSQL, ownerID, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
