package hubprofile

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type Profile struct {
	ID             int64
	Email          string
	FullName       string
	Apps           []string
	Role           string
	BusinessID     int64
	BusinessName   string
	IsOwner        bool
	IsConsoleAdmin bool
	Onboarded      bool
}

type Resolver interface {
	GetProfile(ctx context.Context, userID, businessID int64) (Profile, error)
}
