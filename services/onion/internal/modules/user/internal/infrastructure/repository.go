package infrastructure

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure/persistence/sqlc"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	var signupIP *netip.Addr

	if user.SignupIPAddress != "" {
		if ip, err := netip.ParseAddr(user.SignupIPAddress); err == nil {
			signupIP = &ip
		}
	}

	arg := sqlc.CreateUserParams{
		Email:            user.Email,
		PasswordHash:     user.PasswordHash,
		FullName:         pgtype.Text{String: user.FullName, Valid: user.FullName != ""},
		SignupIpAddress:  signupIP,
		SignupDeviceInfo: pgtype.Text{String: user.SignupDeviceInfo, Valid: user.SignupDeviceInfo != ""},
	}

	createdUser, err := db.NewWriteExecutor[domain.User, sqlc.User](db.MainDatabase).Insert(ctx, r.queries.CreateUser, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user", logger.Error(err))

		return 0, err
	}

	return createdUser.ID, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := db.NewReadExecutor[domain.User, sqlc.User](db.MainDatabase).Fetch(ctx, r.queries.GetUserByEmail, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := db.NewReadExecutor[domain.User, sqlc.User](db.MainDatabase).Fetch(ctx, r.queries.GetUserByID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get user by ID", logger.Error(err))

		return nil, err
	}

	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *domain.User) error {
	arg := sqlc.UpdateUserProfileParams{
		ID:       user.ID,
		FullName: pgtype.Text{String: user.FullName, Valid: user.FullName != ""},
	}

	_, err := db.NewWriteExecutor[domain.User, sqlc.User](db.MainDatabase).Update(ctx, r.queries.UpdateUserProfile, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update user", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	arg := sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	}

	err := db.NewWriteExecutor[domain.User, sqlc.User](db.MainDatabase).Exec(ctx, r.queries.UpdateUserPassword, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update user password", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	err := db.NewWriteExecutor[domain.User, sqlc.User](db.MainDatabase).Delete(ctx, r.queries.DeleteUser, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete user", logger.Error(err))

		return err
	}

	return nil
}
