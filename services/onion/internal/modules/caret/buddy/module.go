package buddy

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/buddy/internal/infrastructure"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/logger"
)

// EnvEnabled switches the whole buddy feature on. It is off by default so the
// schedule job and the runner cost nothing until it is turned on.
const EnvEnabled = "CARET_BUDDY_ENABLED"

const (
	scheduleJobName    = "caret.buddy_schedule"
	defaultSchedule    = "* * * * *"
	scheduleBatchLimit = 100
)

// Enabled reports whether buddies are switched on for this process.
func Enabled() bool {
	enabled, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(EnvEnabled)))
	return enabled
}

type Module struct {
	repo *infrastructure.Repository
}

func NewModule(_ context.Context, db *sql.DB) *Module {
	return &Module{repo: infrastructure.NewRepository(db)}
}

// RegisterJobs adds the job that queues a run for every buddy that is due. It
// only touches the database, so the worker can run it without the tool
// modules the runner needs.
func (m *Module) RegisterJobs(_ context.Context, registry *platformjobs.Registry) {
	if !Enabled() {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        scheduleJobName,
		Description: "Queue a Caret buddy run for every business whose check is due.",
		Schedule:    platformjobs.ScheduleFromEnv("CARET_BUDDY_SCHEDULE", defaultSchedule),
		Handler: func(ctx context.Context) error {
			queued, err := m.repo.EnqueueDue(ctx, scheduleBatchLimit)
			if err != nil {
				return err
			}
			if queued > 0 {
				logger.Ctx(ctx).Info("caret buddy: queued runs", logger.Int("runs", queued))
			}
			return nil
		},
	})
}

// ProposeAction stores a write a buddy wants to make until a person decides.
func (m *Module) ProposeAction(ctx context.Context, runID, businessID int64, tool string, input map[string]any, fingerprint string, ttl time.Duration) error {
	return m.repo.ProposeAction(ctx, runID, businessID, tool, input, fingerprint, ttl)
}

// Settings is what a business configures about its buddy.
type Settings struct {
	BusinessID      int64
	RunAsUserID     int64
	Enabled         bool
	Instructions    string
	IntervalMinutes int
	DailyRunCap     int
}

// Save creates or updates a business's buddy.
func (m *Module) Save(ctx context.Context, s Settings) error {
	return m.repo.UpsertBuddy(ctx, domain.Buddy{
		BusinessID: s.BusinessID, RunAsUserID: s.RunAsUserID, Enabled: s.Enabled,
		Instructions: strings.TrimSpace(s.Instructions), IntervalMinutes: s.IntervalMinutes, DailyRunCap: s.DailyRunCap,
	})
}
