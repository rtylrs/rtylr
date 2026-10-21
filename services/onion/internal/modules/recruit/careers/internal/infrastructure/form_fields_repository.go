package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
)

var ErrFormFieldNotFound = domain.ErrFormFieldNotFound

func (r *Repository) ListFormFields(ctx context.Context, businessID, careerID int64) ([]domain.FormField, error) {
	rows, err := r.queries.ListFormFields(ctx, sqlc.ListFormFieldsParams{
		BusinessID: businessID,
		CareerID:   careerID,
	})
	if err != nil {
		return nil, err
	}
	return formFieldsFromRows(rows), nil
}

func (r *Repository) ListFormFieldsByCareer(ctx context.Context, careerID int64) ([]domain.FormField, error) {
	rows, err := r.queries.ListFormFieldsByCareer(ctx, careerID)
	if err != nil {
		return nil, err
	}
	return formFieldsFromCareerRows(rows), nil
}

func (r *Repository) CreateFormField(ctx context.Context, in domain.FormFieldInput) (int64, error) {
	result, err := r.queries.CreateFormField(ctx, sqlc.CreateFormFieldParams{
		CareerID:    in.CareerID,
		OwnerID:     in.OwnerID,
		BusinessID:  in.BusinessID,
		Kind:        in.Kind,
		Label:       in.Label,
		Placeholder: nullStr(in.Placeholder),
		Required:    in.Required,
		OptionsJson: jsonBytesFromString(in.OptionsJSON),
		Position:    in.Position,
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateFormField(ctx context.Context, in domain.FormFieldInput) error {
	return r.queries.UpdateFormField(ctx, sqlc.UpdateFormFieldParams{
		Kind:        in.Kind,
		Label:       in.Label,
		Placeholder: nullStr(in.Placeholder),
		Required:    in.Required,
		OptionsJson: jsonBytesFromString(in.OptionsJSON),
		Position:    in.Position,
		ID:          in.ID,
		BusinessID:  in.BusinessID,
	})
}

func (r *Repository) DeleteFormField(ctx context.Context, businessID, fieldID int64) error {
	result, err := r.queries.DeleteFormField(ctx, sqlc.DeleteFormFieldParams{
		ID:         fieldID,
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
		return ErrFormFieldNotFound
	}
	return nil
}

func (r *Repository) GetFormField(ctx context.Context, businessID, fieldID int64) (domain.FormField, error) {
	row, err := r.queries.GetFormField(ctx, sqlc.GetFormFieldParams{
		ID:         fieldID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FormField{}, ErrFormFieldNotFound
		}
		return domain.FormField{}, err
	}
	return formFieldFromRow(row), nil
}

func formFieldsFromRows(rows []sqlc.ListFormFieldsRow) []domain.FormField {
	out := make([]domain.FormField, 0, len(rows))
	for _, row := range rows {
		out = append(out, formFieldFromListRow(row))
	}
	return out
}

func formFieldsFromCareerRows(rows []sqlc.ListFormFieldsByCareerRow) []domain.FormField {
	out := make([]domain.FormField, 0, len(rows))
	for _, row := range rows {
		out = append(out, formFieldFromCareerRow(row))
	}
	return out
}

func formFieldFromListRow(row sqlc.ListFormFieldsRow) domain.FormField {
	return domain.FormField{
		ID:          row.ID,
		CareerID:    row.CareerID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Kind:        row.Kind,
		Label:       row.Label,
		Placeholder: row.Placeholder,
		Required:    row.Required,
		OptionsJSON: interfaceToString(row.OptionsJson),
		Position:    row.Position,
		CreatedAt:   nullTimeValue(row.CreatedAt),
		UpdatedAt:   nullTimeValue(row.UpdatedAt),
	}
}

func formFieldFromCareerRow(row sqlc.ListFormFieldsByCareerRow) domain.FormField {
	return domain.FormField{
		ID:          row.ID,
		CareerID:    row.CareerID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Kind:        row.Kind,
		Label:       row.Label,
		Placeholder: row.Placeholder,
		Required:    row.Required,
		OptionsJSON: interfaceToString(row.OptionsJson),
		Position:    row.Position,
		CreatedAt:   nullTimeValue(row.CreatedAt),
		UpdatedAt:   nullTimeValue(row.UpdatedAt),
	}
}

func formFieldFromRow(row sqlc.GetFormFieldRow) domain.FormField {
	return domain.FormField{
		ID:          row.ID,
		CareerID:    row.CareerID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Kind:        row.Kind,
		Label:       row.Label,
		Placeholder: row.Placeholder,
		Required:    row.Required,
		OptionsJSON: interfaceToString(row.OptionsJson),
		Position:    row.Position,
		CreatedAt:   nullTimeValue(row.CreatedAt),
		UpdatedAt:   nullTimeValue(row.UpdatedAt),
	}
}
