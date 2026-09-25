package domain

import "time"

var (
	ErrNotFound              = errorString("custom domain not found")
	ErrInvalidToken          = errorString("invalid verification token")
	ErrDNSVerificationFailed = errorString("custom domain DNS verification failed")
	ErrConflict              = errorString("custom domain already belongs to another business")
)

type errorString string

func (e errorString) Error() string { return string(e) }

type CustomDomain struct {
	ID                int64
	OwnerID           int64
	BusinessID        int64
	Domain            string
	Status            string
	VerificationToken string
	SSLStatus         string
	VerifiedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
