package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/channels/internal/infrastructure/persistence/sqlc"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrUserNotFound    = errors.New("zontriq user not found")
)

type Channel struct {
	ID                int64  `json:"id"`
	ZontriqUserID     int64  `json:"zontriq_user_id"`
	Platform          string `json:"platform"`
	PlatformChannelID string `json:"platform_channel_id"`
	ChannelName       string `json:"channel_name"`
	ChannelHandle     string `json:"channel_handle"`
	ChannelAvatarUrl  string `json:"channel_avatar_url"`
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	TokenExpiresAt    string `json:"token_expires_at"`
	ConnectionState   string `json:"connection_state"`
	ConnectedAt       string `json:"connected_at"`
	Status            int32  `json:"status"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func (c *Channel) FromRows(row sqlc.ZontriqChannel) error {
	c.ID = int64(row.ID)
	c.ZontriqUserID = int64(row.ZontriqUserID)
	c.Platform = row.Platform
	c.PlatformChannelID = row.PlatformChannelID
	c.ChannelName = row.ChannelName
	c.ChannelHandle = convert.ToString(row.ChannelHandle)
	c.ChannelAvatarUrl = convert.ToString(row.ChannelAvatarUrl)
	c.AccessToken = row.AccessToken
	c.RefreshToken = convert.ToString(row.RefreshToken)

	if row.TokenExpiresAt.Valid {
		c.TokenExpiresAt = row.TokenExpiresAt.Time.Format(time.RFC3339)
	}

	c.ConnectionState = row.ConnectionState

	if row.ConnectedAt.Valid {
		c.ConnectedAt = row.ConnectedAt.Time.Format(time.RFC3339)
	}

	c.Status = row.Status

	if row.CreatedAt.Valid {
		c.CreatedAt = row.CreatedAt.Time.Format(time.RFC3339)
	}

	if row.UpdatedAt.Valid {
		c.UpdatedAt = row.UpdatedAt.Time.Format(time.RFC3339)
	}

	return nil
}
