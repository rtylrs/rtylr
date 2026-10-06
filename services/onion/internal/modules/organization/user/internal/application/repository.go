package application

import (
	"context"
	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"time"
)

type repository interface {
	AssignSubscription(ctx context.Context, ownerID int64, planSlug string) error
	BusinessNameExists(ctx context.Context, name string) (bool, error)
	CheckBusinessAccess(ctx context.Context, userID, businessID int64) (bool, error)
	ConsumeEmailVerificationCode(ctx context.Context, token string) error
	CreateBusiness(ctx context.Context, name string, ownerID int64) (int64, error)
	CreateEmailVerificationCode(ctx context.Context, email string) (*domain.EmailVerification, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	DeleteUser(ctx context.Context, id int64) error
	EmailExists(ctx context.Context, email string) (bool, error)
	GetUserAccessState(ctx context.Context, userID int64) (domain.UserAccessState, error)
	GetUserBusinessID(ctx context.Context, userID int64) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
	RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error
	CreateSession(ctx context.Context, userID int64, tokenHash, ip, userAgent, method string, expiresAt time.Time) error
	GetSessionUserID(ctx context.Context, tokenHash string) (int64, error)
	RevokeSession(ctx context.Context, tokenHash string) error
	SetUserBusinessID(ctx context.Context, userID, businessID int64) error
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
}

type emailNotifications interface {
	SendVerification(ctx context.Context, to, subject, htmlBody, textBody string) error
	QueueWelcome(ctx context.Context, to, name string) error
	QueuePasswordReset(ctx context.Context, to, resetLink string) error
}
