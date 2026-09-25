package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error) {
	s := domain.UserSettings{
		UserID:       userID,
		TaskAssigned: true,
		TaskComments: true,
		Mentions:     true,
		WeeklyDigest: true,
	}

	row, err := r.queries.GetUserSettings(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return domain.UserSettings{}, err
	}
	s.TaskAssigned = row.TaskAssigned.Bool
	s.TaskComments = row.TaskComments.Bool
	s.Mentions = row.Mentions.Bool
	s.StatusChanges = row.StatusChanges.Bool
	s.WeeklyDigest = row.WeeklyDigest.Bool
	s.EmailNotifications = row.EmailNotifications.Bool
	s.SidebarPreferencesJSON = row.SidebarPreferencesJson.String
	return s, nil
}

func (r *Repository) UpsertUserSettings(ctx context.Context, s domain.UserSettings) error {
	err := r.queries.UpsertUserSettings(ctx, sqlc.UpsertUserSettingsParams{
		UserID:                 s.UserID,
		TaskAssigned:           sql.NullBool{Bool: s.TaskAssigned, Valid: true},
		TaskComments:           sql.NullBool{Bool: s.TaskComments, Valid: true},
		Mentions:               sql.NullBool{Bool: s.Mentions, Valid: true},
		StatusChanges:          sql.NullBool{Bool: s.StatusChanges, Valid: true},
		WeeklyDigest:           sql.NullBool{Bool: s.WeeklyDigest, Valid: true},
		EmailNotifications:     sql.NullBool{Bool: s.EmailNotifications, Valid: true},
		SidebarPreferencesJson: sql.NullString{String: s.SidebarPreferencesJSON, Valid: s.SidebarPreferencesJSON != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpsertUserSettings failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Roadmap task/progress helpers
// ─────────────────────────────────────────────

func (r *Repository) GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int) {
	row, err := r.queries.GetProjectTaskCounts(ctx, projectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetProjectTaskCounts failed", logger.Error(err))
		return 0, 0
	}
	return int(row.Total), int(countIfaceToInt64(row.Done))
}

func (r *Repository) GetProjectTaskCountsByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64]domain.ProjectTaskCounts, error) {
	counts := make(map[int64]domain.ProjectTaskCounts, len(projectIDs))
	if len(projectIDs) == 0 {
		return counts, nil
	}

	rows, err := r.queries.GetProjectTaskCountsByProjectIDs(ctx, projectIDs)
	if err != nil {
		logger.Ctx(ctx).Warn("GetProjectTaskCountsByProjectIDs failed", logger.Error(err))
		return nil, err
	}
	for _, row := range rows {
		counts[row.ProjectID] = domain.ProjectTaskCounts{
			Total: int(row.Total),
			Done:  int(countIfaceToInt64(row.Done)),
		}
	}
	return counts, nil
}

func countIfaceToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		i, err := strconv.ParseInt(string(n), 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func countIfaceToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case []byte:
		f, err := strconv.ParseFloat(string(n), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func ifaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprint(v)
	}
}
