package application

import (
	"context"
	"sort"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
)

// BusinessAvatar is one colleague's chosen avatar.
type BusinessAvatar struct {
	UserID int64
	Avatar domain.Avatar
}

// WithAvatarPhotoBaseURL sets the public CDN origin an uploaded avatar photo
// must come from. Without it no photo avatar validates.
func (s *Service) WithAvatarPhotoBaseURL(base string) *Service {
	s.avatarPhotoBase = base
	return s
}

// GetAvatar returns the user's chosen avatar, or nil for the default.
func (s *Service) GetAvatar(ctx context.Context, userID int64) (*domain.Avatar, error) {
	stored, err := s.repo.GetUserAvatar(ctx, userID)
	if err != nil {
		return nil, err
	}
	return domain.DecodeAvatar(stored, s.avatarPhotoBase), nil
}

// UpdateAvatar sets the user's avatar, or resets it to the default when nil.
func (s *Service) UpdateAvatar(ctx context.Context, userID int64, avatar *domain.Avatar) (*domain.Avatar, error) {
	defer invalidateUserScope(ctx, userID)
	if avatar == nil {
		return nil, s.repo.SetUserAvatar(ctx, userID, "")
	}
	if err := avatar.Validate(s.avatarPhotoBase); err != nil {
		return nil, err
	}
	stored, err := domain.EncodeAvatar(*avatar)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SetUserAvatar(ctx, userID, stored); err != nil {
		return nil, err
	}
	normalized := avatar.Normalized()
	return &normalized, nil
}

// ListBusinessAvatars returns the chosen avatars of the business's owner and
// active members. Members on the default are left out.
func (s *Service) ListBusinessAvatars(ctx context.Context, businessID int64) ([]BusinessAvatar, error) {
	rows, err := s.repo.ListBusinessUserAvatars(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]BusinessAvatar, 0, len(rows))
	for userID, stored := range rows {
		if a := domain.DecodeAvatar(stored, s.avatarPhotoBase); a != nil {
			out = append(out, BusinessAvatar{UserID: userID, Avatar: *a})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out, nil
}
