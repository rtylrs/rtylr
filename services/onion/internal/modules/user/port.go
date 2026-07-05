package user

import (
	"context"

	userapplication "github.com/voxire/onion/internal/modules/user/internal/application"
)

// Onboarding errors re-exported for cross-module handlers.
var (
	ErrBusinessNameRequired = userapplication.ErrBusinessNameRequired
	ErrBusinessNameTaken    = userapplication.ErrBusinessNameTaken
)

// OnboardingPort covers owner onboarding flows implemented by the user service.
type OnboardingPort interface {
	BusinessNameExists(ctx context.Context, name string) (bool, error)
	OnboardCreateBusiness(ctx context.Context, userID int64, name string) (string, error)
}
