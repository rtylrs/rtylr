package domain

import (
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/series/internal/infrastructure/persistence/sqlc"
)

var (
	ErrSeriesNotFound  = errors.New("series not found")
	ErrContextNotFound = errors.New("series context not found")
)

type Series struct {
	Id          int64
	Name        string
	Description string
	Niche       string
	Voice       string
	Style       string
	Frequency   string
	Config      string
	PublishTime string
	State       string
	VideoCount  int32
	Status      int32
	CreatedAt   string
	UpdatedAt   string
}

func (s *Series) FromRows(row sqlc.ZontriqSeries) error {
	s.Id = int64(row.ID)
	s.Name = row.Name
	s.Description = convert.ToString(row.Description)
	s.Niche = row.Niche
	s.Voice = row.Voice
	s.Style = convert.ToString(row.Style)
	s.Frequency = row.Frequency
	s.Config = string(row.Config)
	s.PublishTime = convert.ToString(row.PublishTime)
	s.State = row.State
	s.VideoCount = row.VideoCount
	s.Status = row.Status
	s.CreatedAt = convert.ToTimestamp(row.CreatedAt).Time.String()
	s.UpdatedAt = convert.ToTimestamp(row.UpdatedAt).Time.String()

	return nil
}

type SeriesDetail struct {
	Id          int64
	Name        string
	Description string
	Niche       string
	Voice       string
	Style       string
	Frequency   string
	Config      string
	PublishTime string
	State       string
	VideoCount  int32
	Status      int32
	CreatedAt   string
	UpdatedAt   string
}

func (s *SeriesDetail) FromRows(row sqlc.ZontriqSeries) error {
	s.Id = int64(row.ID)
	s.Name = row.Name
	s.Description = convert.ToString(row.Description)
	s.Niche = row.Niche
	s.Voice = row.Voice
	s.Style = convert.ToString(row.Style)
	s.Frequency = row.Frequency
	s.Config = string(row.Config)
	s.PublishTime = convert.ToString(row.PublishTime)
	s.State = row.State
	s.VideoCount = row.VideoCount
	s.Status = row.Status
	s.CreatedAt = convert.ToTimestamp(row.CreatedAt).Time.String()
	s.UpdatedAt = convert.ToTimestamp(row.UpdatedAt).Time.String()

	return nil
}

type SeriesCreated struct {
	Id          int64
	Name        string
	Description string
	Niche       string
	Voice       string
	Style       string
	Frequency   string
	Config      string
	PublishTime string
	State       string
	VideoCount  int32
	Status      int32
	CreatedAt   string
	UpdatedAt   string
}

func (s *SeriesCreated) FromRows(row sqlc.ZontriqSeries) error {
	s.Id = int64(row.ID)
	s.Name = row.Name
	s.Description = convert.ToString(row.Description)
	s.Niche = row.Niche
	s.Voice = row.Voice
	s.Style = convert.ToString(row.Style)
	s.Frequency = row.Frequency
	s.Config = string(row.Config)
	s.PublishTime = convert.ToString(row.PublishTime)
	s.State = row.State
	s.VideoCount = row.VideoCount
	s.Status = row.Status
	s.CreatedAt = convert.ToTimestamp(row.CreatedAt).Time.String()
	s.UpdatedAt = convert.ToTimestamp(row.UpdatedAt).Time.String()

	return nil
}

type SeriesUpdated struct {
	Id          int64
	Name        string
	Description string
	Niche       string
	Voice       string
	Style       string
	Frequency   string
	Config      string
	PublishTime string
	State       string
	VideoCount  int32
	Status      int32
	CreatedAt   string
	UpdatedAt   string
}

func (s *SeriesUpdated) FromRows(row sqlc.ZontriqSeries) error {
	s.Id = int64(row.ID)
	s.Name = row.Name
	s.Description = convert.ToString(row.Description)
	s.Niche = row.Niche
	s.Voice = row.Voice
	s.Style = convert.ToString(row.Style)
	s.Frequency = row.Frequency
	s.Config = string(row.Config)
	s.PublishTime = convert.ToString(row.PublishTime)
	s.State = row.State
	s.VideoCount = row.VideoCount
	s.Status = row.Status
	s.CreatedAt = convert.ToTimestamp(row.CreatedAt).Time.String()
	s.UpdatedAt = convert.ToTimestamp(row.UpdatedAt).Time.String()

	return nil
}

type SeriesContext struct {
	Id                int64
	SeriesId          int64
	Context           string
	TopicsCovered     []string
	LastEpisodeNumber int32
}

func (sc *SeriesContext) FromRows(row sqlc.ZontriqSeriesContext) error {
	sc.Id = int64(row.ID)
	sc.SeriesId = int64(row.SeriesID)
	sc.Context = row.Context
	sc.TopicsCovered = parseTopicsCovered(row.TopicsCovered)
	sc.LastEpisodeNumber = row.LastEpisodeNumber

	return nil
}

type SeriesContextCreated struct {
	Id                int64
	SeriesId          int64
	Context           string
	TopicsCovered     []string
	LastEpisodeNumber int32
}

func (sc *SeriesContextCreated) FromRows(row sqlc.ZontriqSeriesContext) error {
	sc.Id = int64(row.ID)
	sc.SeriesId = int64(row.SeriesID)
	sc.Context = row.Context
	sc.TopicsCovered = parseTopicsCovered(row.TopicsCovered)
	sc.LastEpisodeNumber = row.LastEpisodeNumber

	return nil
}

type SeriesContextUpdated struct {
	Id                int64
	SeriesId          int64
	Context           string
	TopicsCovered     []string
	LastEpisodeNumber int32
}

func (sc *SeriesContextUpdated) FromRows(row sqlc.ZontriqSeriesContext) error {
	sc.Id = int64(row.ID)
	sc.SeriesId = int64(row.SeriesID)
	sc.Context = row.Context
	sc.TopicsCovered = parseTopicsCovered(row.TopicsCovered)
	sc.LastEpisodeNumber = row.LastEpisodeNumber

	return nil
}

func parseTopicsCovered(data []byte) []string {
	var topics []string
	if len(data) == 0 {
		return []string{}
	}

	if err := json.Unmarshal(data, &topics); err != nil {
		return []string{}
	}

	return topics
}
