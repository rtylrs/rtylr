package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/audio/internal/infrastructure/persistence/sqlc"
)

var (
	ErrAudioNotFound = errors.New("audio not found")
	ErrUserNotFound  = errors.New("zontriq user not found")
)

type AudioClip struct {
	ID              int32
	ZontriqUserID   int32
	Title           string
	Script          string
	VoiceID         string
	VoiceName       string
	Provider        string
	AudioUrl        string
	DurationSeconds int32
	State           string
	RetryCount      int32
	ExternalJobID   string
	ErrorMessage    string
	Status          int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a *AudioClip) FromRows(row sqlc.ZontriqAudio) error {
	a.ID = row.ID
	a.ZontriqUserID = row.ZontriqUserID
	a.Title = row.Title
	a.Script = row.Script
	a.VoiceID = row.VoiceID
	a.VoiceName = convert.ToString(row.VoiceName)
	a.Provider = row.Provider
	a.AudioUrl = convert.ToString(row.AudioUrl)
	a.DurationSeconds = convert.ToInt32(row.DurationSeconds)
	a.State = row.State
	a.RetryCount = row.RetryCount
	a.ExternalJobID = convert.ToString(row.ExternalJobID)
	a.ErrorMessage = convert.ToString(row.ErrorMessage)
	a.Status = row.Status
	a.CreatedAt = row.CreatedAt.Time
	a.UpdatedAt = row.UpdatedAt.Time

	return nil
}
