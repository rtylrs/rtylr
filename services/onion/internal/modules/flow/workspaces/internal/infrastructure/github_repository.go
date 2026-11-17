package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	return toGitHubInstallation(row), nil
}

// GetGitHubInstallationByInstallationID resolves the business that holds a
// GitHub installation id, for webhooks. domain.ErrGitHubNotConnected when no
// business holds it.
func (r *Repository) GetGitHubInstallationByInstallationID(ctx context.Context, installationID int64) (domain.GitHubInstallation, error) {
	row, err := r.queries.GetGitHubInstallationByInstallationID(ctx, installationID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GitHubInstallation{}, domain.ErrGitHubNotConnected
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetGitHubInstallationByInstallationID failed", logger.Error(err))
		return domain.GitHubInstallation{}, err
	}
	return toGitHubInstallation(row), nil
}

func toGitHubInstallation(row sqlc.FlowGithubInstallation) domain.GitHubInstallation {
	return domain.GitHubInstallation{
		ID:               row.ID,
		BusinessID:       row.BusinessID,
		InstallationID:   row.InstallationID,
		AccountLogin:     row.AccountLogin,
		AccountType:      row.AccountType,
		InstalledBy:      row.InstalledBy,
		SuspendedAt:      nullTimePtr(row.SuspendedAt),
		LastWebhookAt:    nullTimePtr(row.LastWebhookAt),
		LastWebhookEvent: row.LastWebhookEvent,
		CreatedAt:        row.CreatedAt,
	}
}

func nullTimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
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

// githubDeliveryRetention keeps processed delivery ids long enough to cover
// GitHub's manual redelivery window. Older ids are pruned as deliveries arrive.
const githubDeliveryRetention = 7 * 24 * time.Hour

// ApplyGitHubWebhook records the delivery id and applies its effect in one
// transaction. It returns true without changing anything when the delivery
// id was already recorded. A failed effect rolls back the id too, so the
// delivery can be redelivered.
func (r *Repository) ApplyGitHubWebhook(ctx context.Context, e domain.GitHubWebhookEffect) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)
	recorded, err := q.RecordGitHubWebhookDelivery(ctx, sqlc.RecordGitHubWebhookDeliveryParams{
		DeliveryID: e.DeliveryID,
		BusinessID: e.BusinessID,
		Event:      e.Event,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RecordGitHubWebhookDelivery failed", logger.Error(err))
		return false, err
	}
	if recorded == 0 {
		return true, nil
	}
	if err := applyGitHubWebhookEffect(ctx, q, e); err != nil {
		logger.Ctx(ctx).Error("applying github webhook failed", logger.String("event", e.Event), logger.Error(err))
		return false, err
	}
	if err := q.PruneGitHubWebhookDeliveries(ctx, e.ReceivedAt.Add(-githubDeliveryRetention)); err != nil {
		logger.Ctx(ctx).Error("PruneGitHubWebhookDeliveries failed", logger.Error(err))
		return false, err
	}
	return false, tx.Commit()
}

func applyGitHubWebhookEffect(ctx context.Context, q *sqlc.Queries, e domain.GitHubWebhookEffect) error {
	if e.RevokeGitHubUserID != 0 {
		if _, err := q.DeleteGitHubUserLinksByGitHubUser(ctx, e.RevokeGitHubUserID); err != nil {
			return err
		}
	}
	if e.InstallationID == 0 {
		return nil
	}
	if e.DeleteInstallation {
		if err := q.DeleteInstallationProjectRepositories(ctx, sqlc.DeleteInstallationProjectRepositoriesParams{
			BusinessID:     e.BusinessID,
			InstallationID: e.InstallationID,
		}); err != nil {
			return err
		}
		return q.DeleteGitHubInstallationByID(ctx, sqlc.DeleteGitHubInstallationByIDParams{
			BusinessID:     e.BusinessID,
			InstallationID: e.InstallationID,
		})
	}
	if e.SetSuspended {
		var suspendedAt sql.NullTime
		if e.SuspendedAt != nil {
			suspendedAt = sql.NullTime{Time: *e.SuspendedAt, Valid: true}
		}
		if err := q.SetGitHubInstallationSuspendedAt(ctx, sqlc.SetGitHubInstallationSuspendedAtParams{
			SuspendedAt:    suspendedAt,
			BusinessID:     e.BusinessID,
			InstallationID: e.InstallationID,
		}); err != nil {
			return err
		}
	}
	for _, repositoryID := range e.RemoveRepositoryIDs {
		if err := q.DeleteInstallationProjectRepository(ctx, sqlc.DeleteInstallationProjectRepositoryParams{
			BusinessID:     e.BusinessID,
			InstallationID: e.InstallationID,
			RepositoryID:   repositoryID,
		}); err != nil {
			return err
		}
	}
	return q.TouchGitHubInstallationWebhook(ctx, sqlc.TouchGitHubInstallationWebhookParams{
		LastWebhookAt:    sql.NullTime{Time: e.ReceivedAt, Valid: true},
		LastWebhookEvent: e.Event,
		BusinessID:       e.BusinessID,
		InstallationID:   e.InstallationID,
	})
}

// ListInstallationProjectRepositoryIDs returns every repository id the
// business's projects picked from one installation.
func (r *Repository) ListInstallationProjectRepositoryIDs(ctx context.Context, businessID, installationID int64) ([]int64, error) {
	ids, err := r.queries.ListInstallationProjectRepositoryIDs(ctx, sqlc.ListInstallationProjectRepositoryIDsParams{
		BusinessID:     businessID,
		InstallationID: installationID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListInstallationProjectRepositoryIDs failed", logger.Error(err))
	}
	return ids, err
}

// GetGitHubUserLink returns the member's linked account, or
// domain.ErrGitHubNotLinked.
func (r *Repository) GetGitHubUserLink(ctx context.Context, businessID, userID int64) (domain.GitHubUserLink, error) {
	row, err := r.queries.GetGitHubUserLink(ctx, sqlc.GetGitHubUserLinkParams{BusinessID: businessID, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GitHubUserLink{}, domain.ErrGitHubNotLinked
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetGitHubUserLink failed", logger.Error(err))
		return domain.GitHubUserLink{}, err
	}
	return domain.GitHubUserLink{
		BusinessID:   row.BusinessID,
		UserID:       row.UserID,
		GitHubUserID: row.GithubUserID,
		Login:        row.GithubLogin,
		AvatarURL:    row.GithubAvatarUrl,
		CreatedAt:    row.CreatedAt,
	}, nil
}

// ReplaceGitHubUserLink stores the member's account in place of any earlier
// one. A duplicate key after the delete can only be the GitHub account being
// linked to another member of the business.
func (r *Repository) ReplaceGitHubUserLink(ctx context.Context, link domain.GitHubUserLink) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)
	if err := q.DeleteGitHubUserLink(ctx, sqlc.DeleteGitHubUserLinkParams{BusinessID: link.BusinessID, UserID: link.UserID}); err != nil {
		logger.Ctx(ctx).Error("DeleteGitHubUserLink failed", logger.Error(err))
		return err
	}
	err = q.CreateGitHubUserLink(ctx, sqlc.CreateGitHubUserLinkParams{
		BusinessID:      link.BusinessID,
		UserID:          link.UserID,
		GithubUserID:    link.GitHubUserID,
		GithubLogin:     link.Login,
		GithubAvatarUrl: link.AvatarURL,
	})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return domain.ErrGitHubAccountLinkedElsewhere
	}
	if err != nil {
		logger.Ctx(ctx).Error("CreateGitHubUserLink failed", logger.Error(err))
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteGitHubUserLink(ctx context.Context, businessID, userID int64) error {
	err := r.queries.DeleteGitHubUserLink(ctx, sqlc.DeleteGitHubUserLinkParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteGitHubUserLink failed", logger.Error(err))
	}
	return err
}
