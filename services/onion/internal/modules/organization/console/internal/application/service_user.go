package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

type UserDTO struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	FullName       string  `json:"full_name"`
	BusinessID     *int64  `json:"business_id,omitempty"`
	IsConsoleAdmin bool    `json:"is_console_admin"`
	IsBanned       bool    `json:"is_banned"`
	BanReason      *string `json:"ban_reason,omitempty"`
	BannedAt       *string `json:"banned_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func toUserDTO(u UserRow) UserDTO {
	dto := UserDTO{
		ID:             u.ID,
		Email:          u.Email,
		IsConsoleAdmin: u.IsConsoleAdmin,
		IsBanned:       u.IsBanned,
	}
	if u.FullName.Valid {
		dto.FullName = u.FullName.String
	}
	if u.BusinessID.Valid {
		dto.BusinessID = &u.BusinessID.Int64
	}
	if u.BanReason.Valid {
		dto.BanReason = &u.BanReason.String
	}
	if u.BannedAt.Valid {
		t := u.BannedAt.Time.Format(time.RFC3339)
		dto.BannedAt = &t
	}
	if u.CreatedAt.Valid {
		dto.CreatedAt = u.CreatedAt.Time.Format(time.RFC3339)
	}
	return dto
}

type ListUsersResult struct {
	Users []UserDTO `json:"users"`
	Total int64     `json:"total"`
}

func (s *Service) ListUsers(
	ctx context.Context,
	limit, offset int,
	id int64,
	email, fullName string,
) (*ListUsersResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)
	key := consoleKey(fmt.Sprintf("users|l=%d|o=%d|id=%d|email=%s|name=%s", limit, offset, id, email, fullName))
	return cacher.Of[*ListUsersResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*ListUsersResult, error) {
			users, total, err := s.repo.ListUsers(ctx, limit, offset, id, email, fullName)
			if err != nil {
				return nil, err
			}
			dtos := make([]UserDTO, 0, len(users))
			for _, u := range users {
				dtos = append(dtos, toUserDTO(u))
			}
			return &ListUsersResult{Users: dtos, Total: total}, nil
		},
		cacher.WithExpiry(ctx, consoleTTL),
	)
}

func (s *Service) GetUser(ctx context.Context, id int64) (*UserDTO, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*u)
	return &dto, nil
}

func (s *Service) CreateUser(ctx context.Context, email, password, fullName string, isConsoleAdmin bool) (*UserDTO, error) {
	defer invalidateConsoleScope(ctx)

	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	id, err := s.repo.CreateUser(ctx, email, hash, fullName, isConsoleAdmin)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, id)
}

type CreateBusinessOwnerInput struct {
	Email              string
	Password           string
	FullName           string
	BusinessName       string
	Industry           string
	Country            string
	Timezone           string
	Currency           string
	GrantPlan          bool
	PlanID             int64
	SubscriptionStatus string
	GrantDays          int
}

type CreateBusinessOwnerResult struct {
	Owner        UserDTO          `json:"owner"`
	Business     BusinessDTO      `json:"business"`
	Subscription *SubscriptionDTO `json:"subscription,omitempty"`
	Message      string           `json:"message"`
}

func (s *Service) CreateBusinessOwner(ctx context.Context, input CreateBusinessOwnerInput) (*CreateBusinessOwnerResult, error) {
	defer invalidateConsoleScope(ctx)

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, errors.New("owner email is required")
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, errors.New("password is required")
	}
	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	businessName := strings.TrimSpace(input.BusinessName)
	if businessName == "" {
		return nil, errors.New("business name is required")
	}

	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}

	status := strings.ToLower(strings.TrimSpace(input.SubscriptionStatus))
	if status == "" {
		if input.GrantDays > 0 {
			status = "trial"
		} else {
			status = "active"
		}
	}
	if status != "active" && status != "trial" {
		return nil, errors.New("subscription status must be active or trial")
	}
	if input.GrantDays < 0 {
		return nil, errors.New("grant days cannot be negative")
	}
	if input.GrantPlan && input.PlanID <= 0 {
		return nil, errors.New("plan is required when granting access")
	}

	if _, err := s.repo.GetUserByEmail(ctx, email); err == nil {
		return nil, errors.New("owner email already exists")
	} else if !errors.Is(err, ErrRecordNotFound) {
		return nil, err
	}
	if _, total, err := s.repo.ListBusinesses(ctx, 1, 0, 0, businessName, 0); err != nil {
		return nil, err
	} else if total > 0 {
		return nil, errors.New("business name already exists")
	}

	hash, err := passwordutils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateBusinessOwner(ctx, CreateBusinessOwnerParams{
		Email:              email,
		PasswordHash:       hash,
		FullName:           strings.TrimSpace(input.FullName),
		BusinessName:       businessName,
		Industry:           strings.TrimSpace(input.Industry),
		Country:            strings.TrimSpace(input.Country),
		Timezone:           timezone,
		Currency:           currency,
		GrantPlan:          input.GrantPlan,
		PlanID:             input.PlanID,
		SubscriptionStatus: status,
		GrantDays:          input.GrantDays,
	})
	if err != nil {
		return nil, err
	}

	provisionErr := error(nil)
	if s.employeeProvisioner == nil {
		provisionErr = errors.New("HR employee provisioner is not configured")
	} else {
		name := strings.Fields(strings.TrimSpace(input.FullName))
		firstName, lastName := "", ""
		if len(name) > 0 {
			firstName = name[0]
		}
		if len(name) > 1 {
			lastName = strings.Join(name[1:], " ")
		}
		provisionErr = s.employeeProvisioner.EnsureEmployeeForUser(ctx, created.OwnerID, created.BusinessID, created.OwnerID, firstName, lastName, email)
	}
	if provisionErr != nil {
		var cleanupErrors []error
		if s.employeeProvisioner != nil {
			if cleanupErr := s.employeeProvisioner.DeleteEmployeeForUser(ctx, created.OwnerID, created.BusinessID, created.OwnerID); cleanupErr != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("delete partially provisioned HR employee: %w", cleanupErr))
			}
		}
		if cleanupErr := s.repo.DeleteCreatedBusinessOwner(ctx, created.OwnerID, created.BusinessID); cleanupErr != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("remove created owner and business: %w", cleanupErr))
		}
		if len(cleanupErrors) > 0 {
			return nil, fmt.Errorf("provision business owner employee: %w; compensation failed: %w", provisionErr, errors.Join(cleanupErrors...))
		}
		return nil, provisionErr
	}

	owner, err := s.GetUser(ctx, created.OwnerID)
	if err != nil {
		return nil, err
	}
	business, err := s.GetBusiness(ctx, created.BusinessID)
	if err != nil {
		return nil, err
	}

	var subscription *SubscriptionDTO
	if created.SubscriptionID.Valid {
		sub, subErr := s.GetSubscriptionByOwner(ctx, created.OwnerID)
		if subErr != nil {
			return nil, subErr
		}
		subscription = sub
	}

	return &CreateBusinessOwnerResult{
		Owner:        *owner,
		Business:     *business,
		Subscription: subscription,
		Message:      "Owner and business created",
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, fullName, email string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UpdateUser(ctx, id, fullName, email)
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteUser(ctx, id)
}
