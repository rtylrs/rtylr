package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/templates/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Template, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, type, subject, body, variables, is_active, created_at, updated_at
		 FROM crm_templates WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list templates", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Template, 0)

	for rows.Next() {
		t, scanErr := scanTemplate(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		items = append(items, t)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.TemplateInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_templates
		   (owner_id, name, type, subject, body, variables, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.Name, input.Type, input.Subject,
		input.Body, string(input.Variables), input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create template", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TemplateUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_templates
		 SET name = ?, type = ?, subject = ?, body = ?, variables = ?, is_active = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.Type, input.Subject, input.Body,
		string(input.Variables), input.IsActive,
		input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update template", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, templateID int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM crm_templates WHERE id = ? AND owner_id = ?",
		templateID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete template", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

type templateScanner interface {
	Scan(dest ...interface{}) error
}

func scanTemplate(s templateScanner) (domain.Template, error) {
	var t domain.Template
	var subject, variables sql.NullString

	if err := s.Scan(
		&t.ID, &t.OwnerID, &t.Name, &t.Type, &subject, &t.Body,
		&variables, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Template{}, domain.ErrTemplateNotFound
		}

		return domain.Template{}, err
	}

	t.Subject = subject.String

	if variables.Valid {
		t.Variables = []byte(variables.String)
	} else {
		t.Variables = []byte("[]")
	}

	return t, nil
}
