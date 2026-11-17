package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// GetGitHubInstallation returns the business's installation, or
// domain.ErrGitHubNotConnected when it has none.
func (r *Repository) GetGitHubInstallation(ctx context.Context, businessID int64) (domain.GitHubInstallation, error) {
	row, err := r.queries.GetGitHubInstallation(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GitHubInstallation{}, domain.ErrGitHubNotConnected
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetGitHubInstallation failed", logger.Error(err))
		return domain.GitHubInstallation{}, err
	}
	return domain.GitHubInstallation{
		ID:             row.ID,
		BusinessID:     row.BusinessID,
		InstallationID: row.InstallationID,
		AccountLogin:   row.AccountLogin,
		AccountType:    row.AccountType,
		InstalledBy:    row.InstalledBy,
		CreatedAt:      row.CreatedAt,
	}, nil
}

// CreateGitHubInstallation stores a new installation. A duplicate key means
// either this business already has one or another business holds the same
// installation id.
func (r *Repository) CreateGitHubInstallation(ctx context.Context, in domain.GitHubInstallation) error {
	err := r.queries.CreateGitHubInstallation(ctx, sqlc.CreateGitHubInstallationParams{
		BusinessID:     in.BusinessID,
		InstallationID: in.InstallationID,
		AccountLogin:   in.AccountLogin,
		AccountType:    in.AccountType,
		InstalledBy:    in.InstalledBy,
	})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return domain.ErrGitHubInstallationClaimed
	}
	if err != nil {
		logger.Ctx(ctx).Error("CreateGitHubInstallation failed", logger.Error(err))
	}
	return err
}

func (r *Repository) UpdateGitHubInstallationAccount(ctx context.Context, businessID, installationID int64, login, accountType string) error {
	err := r.queries.UpdateGitHubInstallationAccount(ctx, sqlc.UpdateGitHubInstallationAccountParams{
		AccountLogin:   login,
		AccountType:    accountType,
		BusinessID:     businessID,
		InstallationID: installationID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateGitHubInstallationAccount failed", logger.Error(err))
	}
	return err
}

// DeleteGitHubInstallation removes the business's installation and every
// project's repository selection in one transaction.
func (r *Repository) DeleteGitHubInstallation(ctx context.Context, businessID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)
	if err := q.DeleteBusinessProjectRepositories(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("DeleteBusinessProjectRepositories failed", logger.Error(err))
		return err
	}
	if err := q.DeleteGitHubInstallation(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("DeleteGitHubInstallation failed", logger.Error(err))
		return err
	}
	return tx.Commit()
}

func (r *Repository) ListProjectRepositories(ctx context.Context, businessID, projectID int64) ([]domain.ProjectRepository, error) {
	rows, err := r.queries.ListProjectRepositories(ctx, sqlc.ListProjectRepositoriesParams{
		BusinessID: businessID,
		ProjectID:  projectID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectRepositories failed", logger.Error(err))
		return nil, err
	}
	out := make([]domain.ProjectRepository, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ProjectRepository{
			RepositoryID: row.RepositoryID,
			FullName:     row.FullName,
			Private:      row.IsPrivate,
			CreatedAt:    row.CreatedAt,
		})
	}
	return out, nil
}

// ReplaceProjectRepositories swaps a project's whole selection in one
// transaction.
func (r *Repository) ReplaceProjectRepositories(ctx context.Context, businessID, projectID, installationID, callerID int64, repos []domain.GitHubRepository) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)
	if err := q.DeleteProjectRepositories(ctx, sqlc.DeleteProjectRepositoriesParams{
		BusinessID: businessID,
		ProjectID:  projectID,
	}); err != nil {
		logger.Ctx(ctx).Error("DeleteProjectRepositories failed", logger.Error(err))
		return err
	}
	for _, repo := range repos {
		if err := q.CreateProjectRepository(ctx, sqlc.CreateProjectRepositoryParams{
			BusinessID:     businessID,
			ProjectID:      projectID,
			InstallationID: installationID,
			RepositoryID:   repo.ID,
			FullName:       repo.FullName,
			IsPrivate:      repo.Private,
			AddedBy:        callerID,
		}); err != nil {
			logger.Ctx(ctx).Error("CreateProjectRepository failed", logger.Error(err))
			return err
		}
	}
	return tx.Commit()
}
