package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/videos/internal/infrastructure/persistence/sqlc"
)

var (
	ErrVideoNotFound = errors.New("video not found")
	ErrJobNotFound   = errors.New("active job not found")
	ErrUserNotFound  = errors.New("zontriq user not found")
)

type Video struct {
	ID            int32
	ZontriqUserID int32
	SeriesID      int32
	Title         string
	Script        string
	VideoUrl      string
	ThumbnailUrl  string
	State         string
	Duration      int32
	EpisodeNumber int32
	ErrorMessage  string
	Status        int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (v *Video) FromRows(row sqlc.ZontriqVideo) error {
	v.ID = row.ID
	v.ZontriqUserID = row.ZontriqUserID
	v.SeriesID = convert.ToInt32(row.SeriesID)
	v.Title = row.Title
	v.Script = convert.ToString(row.Script)
	v.VideoUrl = convert.ToString(row.VideoUrl)
	v.ThumbnailUrl = convert.ToString(row.ThumbnailUrl)
	v.State = row.State
	v.Duration = convert.ToInt32(row.Duration)
	v.EpisodeNumber = convert.ToInt32(row.EpisodeNumber)
	v.ErrorMessage = convert.ToString(row.ErrorMessage)
	v.Status = row.Status
	v.CreatedAt = row.CreatedAt.Time
	v.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type VideoDetail struct {
	ID            int32
	ZontriqUserID int32
	SeriesID      int32
	SeriesName    string
	Title         string
	Script        string
	VideoUrl      string
	ThumbnailUrl  string
	State         string
	Duration      int32
	EpisodeNumber int32
	ErrorMessage  string
	Status        int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (v *VideoDetail) FromRows(row sqlc.GetVideoByIDRow) error {
	v.ID = row.ID
	v.ZontriqUserID = row.ZontriqUserID
	v.SeriesID = convert.ToInt32(row.SeriesID)
	v.SeriesName = row.SeriesName
	v.Title = row.Title
	v.Script = convert.ToString(row.Script)
	v.VideoUrl = convert.ToString(row.VideoUrl)
	v.ThumbnailUrl = convert.ToString(row.ThumbnailUrl)
	v.State = row.State
	v.Duration = convert.ToInt32(row.Duration)
	v.EpisodeNumber = convert.ToInt32(row.EpisodeNumber)
	v.ErrorMessage = convert.ToString(row.ErrorMessage)
	v.Status = row.Status
	v.CreatedAt = row.CreatedAt.Time
	v.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type VideoTarget struct {
	ID             int32
	VideoID        int32
	ChannelID      int32
	State          string
	PublishedAt    time.Time
	PlatformPostID string
	ErrorMessage   string
	Status         int32
	CreatedAt      time.Time
	ChannelName    string
	Platform       string
}

func (vt *VideoTarget) FromRows(row sqlc.ListVideoTargetsByVideoIDRow) error {
	vt.ID = row.ID
	vt.VideoID = row.VideoID
	vt.ChannelID = row.ChannelID

	vt.State = row.State
	if row.PublishedAt.Valid {
		vt.PublishedAt = row.PublishedAt.Time
	}

	vt.PlatformPostID = convert.ToString(row.PlatformPostID)
	vt.ErrorMessage = convert.ToString(row.ErrorMessage)
	vt.Status = row.Status
	vt.CreatedAt = row.CreatedAt.Time
	vt.ChannelName = row.ChannelName
	vt.Platform = row.Platform

	return nil
}

type Job struct {
	ID           int32
	VideoID      int32
	Type         string
	State        string
	Progress     int32
	CurrentStep  string
	ErrorMessage string
	GpuID        string
	StartedAt    time.Time
	CompletedAt  time.Time
	Status       int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (j *Job) FromRows(row sqlc.ZontriqJob) error {
	j.ID = row.ID
	j.VideoID = row.VideoID
	j.Type = row.Type
	j.State = row.State
	j.Progress = row.Progress
	j.CurrentStep = convert.ToString(row.CurrentStep)
	j.ErrorMessage = convert.ToString(row.ErrorMessage)

	j.GpuID = convert.ToString(row.GpuID)
	if row.StartedAt.Valid {
		j.StartedAt = row.StartedAt.Time
	}

	if row.CompletedAt.Valid {
		j.CompletedAt = row.CompletedAt.Time
	}

	j.Status = row.Status
	j.CreatedAt = row.CreatedAt.Time
	j.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type PublishLogEntry struct {
	ID              int32
	VideoID         int32
	ChannelID       int32
	Platform        string
	PlatformPostID  string
	PlatformPostUrl string
	Result          string
	ErrorMessage    string
	PublishedAt     time.Time
	Status          int32
	CreatedAt       time.Time
}

func (p *PublishLogEntry) FromRows(row sqlc.ZontriqPublishLog) error {
	p.ID = row.ID
	p.VideoID = row.VideoID
	p.ChannelID = row.ChannelID
	p.Platform = row.Platform
	p.PlatformPostID = convert.ToString(row.PlatformPostID)
	p.PlatformPostUrl = convert.ToString(row.PlatformPostUrl)
	p.Result = row.Result

	p.ErrorMessage = convert.ToString(row.ErrorMessage)
	if row.PublishedAt.Valid {
		p.PublishedAt = row.PublishedAt.Time
	}

	p.Status = row.Status
	p.CreatedAt = row.CreatedAt.Time

	return nil
}
