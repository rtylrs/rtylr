package domain

import (
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

type User struct {
	ID               int64
	Email            string
	PasswordHash     string
	FullName         string
	SignupIPAddress  string
	SignupDeviceInfo string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewUser(email, passwordHash, fullName, signupIP, deviceInfo string) *User {
	return &User{
		Email:            email,
		PasswordHash:     passwordHash,
		FullName:         fullName,
		SignupIPAddress:  signupIP,
		SignupDeviceInfo: deviceInfo,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (u *User) FromRows(row sqlc.User) {
	u.ID = row.ID
	u.Email = row.Email
	u.PasswordHash = row.PasswordHash
	u.FullName = convert.ToString(row.FullName)
	u.SignupDeviceInfo = convert.ToString(row.SignupDeviceInfo)

	if row.SignupIpAddress.Valid {
		u.SignupIPAddress = row.SignupIpAddress.String
	}

	if row.CreatedAt.Valid {
		u.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		u.UpdatedAt = row.UpdatedAt.Time
	}
}
