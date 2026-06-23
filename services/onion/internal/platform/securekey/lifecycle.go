package securekey

import (
	"crypto/subtle"
	"errors"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid key")
	ErrExpired      = errors.New("key is expired")
	ErrRevoked      = errors.New("key is revoked")
	ErrGraceExpired = errors.New("rotated key grace period expired")
)

type Record struct {
	Hash         string
	PreviousHash string
	RotatedAt    *time.Time
	ExpiresAt    *time.Time
	RevokedAt    *time.Time
}

type MatchResult string

const (
	MatchNone     MatchResult = "none"
	MatchCurrent  MatchResult = "current"
	MatchPrevious MatchResult = "previous"
)

func Match(secret string, record Record, grace time.Duration, now time.Time) (MatchResult, error) {
	if record.RevokedAt != nil && !record.RevokedAt.IsZero() {
		return MatchNone, ErrRevoked
	}
	if record.ExpiresAt != nil && !record.ExpiresAt.IsZero() && now.After(*record.ExpiresAt) {
		return MatchNone, ErrExpired
	}

	incoming := Hash(secret)
	if constantTimeHashEqual(incoming, record.Hash) {
		return MatchCurrent, nil
	}
	if record.PreviousHash != "" && constantTimeHashEqual(incoming, record.PreviousHash) {
		if grace > 0 && record.RotatedAt != nil && now.After(record.RotatedAt.Add(grace)) {
			return MatchNone, ErrGraceExpired
		}
		return MatchPrevious, nil
	}
	return MatchNone, ErrInvalid
}

func constantTimeHashEqual(incoming, expected string) bool {
	if len(incoming) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(incoming), []byte(expected)) == 1
}
