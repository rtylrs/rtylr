package hubprofile

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/platform/authorization"
)

var ErrUserNotFound = errors.New("user not found")

type Profile struct {
	ID                 int64
	Email              string
	FullName           string
	BusinessID         int64
	BusinessName       string
	IsConsoleAdmin     bool
	Onboarded          bool
	EffectiveAccess    *authorization.EffectiveAccess
	ProvisioningStatus *authorization.ProvisioningStatus
}

type Resolver interface {
	GetProfile(ctx context.Context, userID, businessID int64) (Profile, error)
}
