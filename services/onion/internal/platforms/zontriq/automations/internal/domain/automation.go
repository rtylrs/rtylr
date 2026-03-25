package domain

import (
	"errors"
	"time"

	sqlc "github.com/voxire/onion/internal/platforms/zontriq/automations/internal/infrastructure/persistence/sqlc"
)

var ErrAutomationNotFound = errors.New("automation not found")

type Automation struct {
	ID              int64     `json:"id"`
	ZontriqUserID   int       `json:"zontriq_user_id"`
	SeriesID        int64     `json:"series_id"`
	SeriesName      string    `json:"series_name"`
	Name            string    `json:"name"`
	Frequency       string    `json:"frequency"`
	State           string    `json:"state"`
	NextRun         string    `json:"next_run"`
	LastRun         string    `json:"last_run"`
	VideosGenerated int32     `json:"videos_generated"`
	Channels        []Channel `json:"channels"`
	Status          int32     `json:"status"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

func (a *Automation) FromRows(row sqlc.ListAutomationsByUserIDRow) error {
	a.ID = int64(row.ID)
	a.ZontriqUserID = int(row.ZontriqUserID)
	a.SeriesID = int64(row.SeriesID)
	a.SeriesName = row.SeriesName
	a.Name = row.Name
	a.Frequency = row.Frequency
	a.State = row.State
	a.NextRun = formatTimestamptz(row.NextRun.Time, row.NextRun.Valid)
	a.LastRun = formatTimestamptz(row.LastRun.Time, row.LastRun.Valid)
	a.VideosGenerated = row.VideosGenerated
	a.Status = row.Status
	a.CreatedAt = formatTimestamptz(row.CreatedAt.Time, row.CreatedAt.Valid)
	a.UpdatedAt = formatTimestamptz(row.UpdatedAt.Time, row.UpdatedAt.Valid)

	return nil
}

type AutomationDetail struct {
	ID              int64     `json:"id"`
	ZontriqUserID   int       `json:"zontriq_user_id"`
	SeriesID        int64     `json:"series_id"`
	SeriesName      string    `json:"series_name"`
	Name            string    `json:"name"`
	Frequency       string    `json:"frequency"`
	State           string    `json:"state"`
	NextRun         string    `json:"next_run"`
	LastRun         string    `json:"last_run"`
	VideosGenerated int32     `json:"videos_generated"`
	Channels        []Channel `json:"channels"`
	Status          int32     `json:"status"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

func (a *AutomationDetail) FromRows(row sqlc.GetAutomationByIDRow) error {
	a.ID = int64(row.ID)
	a.ZontriqUserID = int(row.ZontriqUserID)
	a.SeriesID = int64(row.SeriesID)
	a.SeriesName = row.SeriesName
	a.Name = row.Name
	a.Frequency = row.Frequency
	a.State = row.State
	a.NextRun = formatTimestamptz(row.NextRun.Time, row.NextRun.Valid)
	a.LastRun = formatTimestamptz(row.LastRun.Time, row.LastRun.Valid)
	a.VideosGenerated = row.VideosGenerated
	a.Status = row.Status
	a.CreatedAt = formatTimestamptz(row.CreatedAt.Time, row.CreatedAt.Valid)
	a.UpdatedAt = formatTimestamptz(row.UpdatedAt.Time, row.UpdatedAt.Valid)

	return nil
}

type AutomationRow struct {
	ID              int64  `json:"id"`
	ZontriqUserID   int    `json:"zontriq_user_id"`
	SeriesID        int64  `json:"series_id"`
	Name            string `json:"name"`
	Frequency       string `json:"frequency"`
	State           string `json:"state"`
	NextRun         string `json:"next_run"`
	LastRun         string `json:"last_run"`
	VideosGenerated int32  `json:"videos_generated"`
	Status          int32  `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (a *AutomationRow) FromRows(row sqlc.ZontriqAutomation) error {
	a.ID = int64(row.ID)
	a.ZontriqUserID = int(row.ZontriqUserID)
	a.SeriesID = int64(row.SeriesID)
	a.Name = row.Name
	a.Frequency = row.Frequency
	a.State = row.State
	a.NextRun = formatTimestamptz(row.NextRun.Time, row.NextRun.Valid)
	a.LastRun = formatTimestamptz(row.LastRun.Time, row.LastRun.Valid)
	a.VideosGenerated = row.VideosGenerated
	a.Status = row.Status
	a.CreatedAt = formatTimestamptz(row.CreatedAt.Time, row.CreatedAt.Valid)
	a.UpdatedAt = formatTimestamptz(row.UpdatedAt.Time, row.UpdatedAt.Valid)

	return nil
}

type AutomationUpdatedRow struct {
	ID              int64  `json:"id"`
	ZontriqUserID   int    `json:"zontriq_user_id"`
	SeriesID        int64  `json:"series_id"`
	Name            string `json:"name"`
	Frequency       string `json:"frequency"`
	State           string `json:"state"`
	NextRun         string `json:"next_run"`
	LastRun         string `json:"last_run"`
	VideosGenerated int32  `json:"videos_generated"`
	Status          int32  `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (a *AutomationUpdatedRow) FromRows(row sqlc.ZontriqAutomation) error {
	a.ID = int64(row.ID)
	a.ZontriqUserID = int(row.ZontriqUserID)
	a.SeriesID = int64(row.SeriesID)
	a.Name = row.Name
	a.Frequency = row.Frequency
	a.State = row.State
	a.NextRun = formatTimestamptz(row.NextRun.Time, row.NextRun.Valid)
	a.LastRun = formatTimestamptz(row.LastRun.Time, row.LastRun.Valid)
	a.VideosGenerated = row.VideosGenerated
	a.Status = row.Status
	a.CreatedAt = formatTimestamptz(row.CreatedAt.Time, row.CreatedAt.Valid)
	a.UpdatedAt = formatTimestamptz(row.UpdatedAt.Time, row.UpdatedAt.Valid)

	return nil
}

type Channel struct {
	ID          int64  `json:"id"`
	ChannelID   int64  `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Platform    string `json:"platform"`
}

func (c *Channel) FromRows(row sqlc.ListAutomationChannelsRow) error {
	c.ID = int64(row.ID)
	c.ChannelID = int64(row.ChannelID)
	c.ChannelName = row.ChannelName
	c.Platform = row.Platform

	return nil
}

type DueAutomation struct {
	ID              int64  `json:"id"`
	ZontriqUserID   int    `json:"zontriq_user_id"`
	SeriesID        int64  `json:"series_id"`
	SeriesName      string `json:"series_name"`
	Name            string `json:"name"`
	Frequency       string `json:"frequency"`
	State           string `json:"state"`
	NextRun         string `json:"next_run"`
	LastRun         string `json:"last_run"`
	VideosGenerated int32  `json:"videos_generated"`
	Status          int32  `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (a *DueAutomation) FromRows(row sqlc.GetDueAutomationsRow) error {
	a.ID = int64(row.ID)
	a.ZontriqUserID = int(row.ZontriqUserID)
	a.SeriesID = int64(row.SeriesID)
	a.SeriesName = row.SeriesName
	a.Name = row.Name
	a.Frequency = row.Frequency
	a.State = row.State
	a.NextRun = formatTimestamptz(row.NextRun.Time, row.NextRun.Valid)
	a.LastRun = formatTimestamptz(row.LastRun.Time, row.LastRun.Valid)
	a.VideosGenerated = row.VideosGenerated
	a.Status = row.Status
	a.CreatedAt = formatTimestamptz(row.CreatedAt.Time, row.CreatedAt.Valid)
	a.UpdatedAt = formatTimestamptz(row.UpdatedAt.Time, row.UpdatedAt.Valid)

	return nil
}

func formatTimestamptz(t time.Time, valid bool) string {
	if !valid {
		return ""
	}

	return t.Format(time.RFC3339)
}
