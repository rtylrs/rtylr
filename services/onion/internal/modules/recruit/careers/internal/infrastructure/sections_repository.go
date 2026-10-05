package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
)

var ErrSectionNotFound = domain.ErrSectionNotFound

func (r *Repository) ListSections(ctx context.Context, businessID int64, enabledOnly bool) ([]domain.Section, error) {
	if enabledOnly {
		rows, err := r.queries.ListEnabledSections(ctx, businessID)
		if err != nil {
			return nil, err
		}
		return sectionsFromEnabledRows(rows), nil
	}

	rows, err := r.queries.ListSections(ctx, businessID)
	if err != nil {
		return nil, err
	}
	return sectionsFromRows(rows), nil
}

func (r *Repository) CreateSection(ctx context.Context, in domain.SectionInput) (int64, error) {
	result, err := r.queries.CreateSection(ctx, sqlc.CreateSectionParams{
		OwnerID:    in.OwnerID,
		BusinessID: in.BusinessID,
		Kind:       in.Kind,
		Title:      nullStr(in.Title),
		Body:       nullStr(in.Body),
		MediaUrl:   nullStr(in.MediaURL),
		ItemsJson:  jsonBytesFromString(in.ItemsJSON),
		Position:   in.Position,
		Enabled:    in.Enabled,
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateSection(ctx context.Context, in domain.SectionInput) error {
	return r.queries.UpdateSection(ctx, sqlc.UpdateSectionParams{
		Kind:       in.Kind,
		Title:      nullStr(in.Title),
		Body:       nullStr(in.Body),
		MediaUrl:   nullStr(in.MediaURL),
		ItemsJson:  jsonBytesFromString(in.ItemsJSON),
		Position:   in.Position,
		Enabled:    in.Enabled,
		ID:         in.ID,
		BusinessID: in.BusinessID,
	})
}

func (r *Repository) ReorderSections(ctx context.Context, businessID int64, orderedIDs []int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)
	for idx, id := range orderedIDs {
		if execErr := qtx.UpdateSectionPosition(ctx, sqlc.UpdateSectionPositionParams{
			Position:   int32(idx),
			ID:         id,
			BusinessID: businessID,
		}); execErr != nil {
			_ = tx.Rollback()
			return execErr
		}
	}
	return tx.Commit()
}

func (r *Repository) DeleteSection(ctx context.Context, businessID, sectionID int64) error {
	result, err := r.queries.DeleteSection(ctx, sqlc.DeleteSectionParams{
		ID:         sectionID,
		BusinessID: businessID,
	})
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSectionNotFound
	}
	return nil
}

func (r *Repository) GetSection(ctx context.Context, businessID, sectionID int64) (domain.Section, error) {
	row, err := r.queries.GetSection(ctx, sqlc.GetSectionParams{
		ID:         sectionID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Section{}, ErrSectionNotFound
		}
		return domain.Section{}, err
	}
	return sectionFromRow(row), nil
}

func sectionsFromRows(rows []sqlc.ListSectionsRow) []domain.Section {
	out := make([]domain.Section, 0, len(rows))
	for _, row := range rows {
		out = append(out, sectionFromListRow(row))
	}
	return out
}

func sectionsFromEnabledRows(rows []sqlc.ListEnabledSectionsRow) []domain.Section {
	out := make([]domain.Section, 0, len(rows))
	for _, row := range rows {
		out = append(out, sectionFromEnabledRow(row))
	}
	return out
}

func sectionFromListRow(row sqlc.ListSectionsRow) domain.Section {
	return domain.Section{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Kind:       row.Kind,
		Title:      row.Title,
		Body:       row.Body,
		MediaURL:   row.MediaUrl,
		ItemsJSON:  interfaceToString(row.ItemsJson),
		Position:   row.Position,
		Enabled:    row.Enabled,
		CreatedAt:  nullTimeValue(row.CreatedAt),
		UpdatedAt:  nullTimeValue(row.UpdatedAt),
	}
}

func sectionFromEnabledRow(row sqlc.ListEnabledSectionsRow) domain.Section {
	return domain.Section{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Kind:       row.Kind,
		Title:      row.Title,
		Body:       row.Body,
		MediaURL:   row.MediaUrl,
		ItemsJSON:  interfaceToString(row.ItemsJson),
		Position:   row.Position,
		Enabled:    row.Enabled,
		CreatedAt:  nullTimeValue(row.CreatedAt),
		UpdatedAt:  nullTimeValue(row.UpdatedAt),
	}
}

func sectionFromRow(row sqlc.GetSectionRow) domain.Section {
	return domain.Section{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Kind:       row.Kind,
		Title:      row.Title,
		Body:       row.Body,
		MediaURL:   row.MediaUrl,
		ItemsJSON:  interfaceToString(row.ItemsJson),
		Position:   row.Position,
		Enabled:    row.Enabled,
		CreatedAt:  nullTimeValue(row.CreatedAt),
		UpdatedAt:  nullTimeValue(row.UpdatedAt),
	}
}
