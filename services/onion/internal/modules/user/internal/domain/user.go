package domain

import (
	"time"
)

type User struct {
	ID               int64
	BusinessID       int64
	Email            string
	PasswordHash     string
	FullName         string
	SignupIPAddress  string
	SignupDeviceInfo string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	IsConsoleAdmin   bool
	IsBanned         bool
	TokenEpoch       int64
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
