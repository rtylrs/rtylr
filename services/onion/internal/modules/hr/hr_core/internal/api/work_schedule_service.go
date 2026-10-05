package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	workschedulev1 "github.com/voxire/voxie/gen/go/hr/work_schedule/v1"
)

// defaultWorkScheduleLimit bounds a request that omits limit. The contract makes
// limit optional, and an unbounded list is the thing CLAUDE.md forbids.
const defaultWorkScheduleLimit int32 = 100

// maxWorkScheduleLimit caps what a caller can ask for, so a large limit cannot
// turn a paginated endpoint back into an unbounded one.
const maxWorkScheduleLimit int32 = 500

// WorkScheduleAPI serves hr.work_schedule.v1.
//
// hr_core had no oneKit surface before this: work schedules were reachable only
// through hand-registered REST routes that no app called and no contract
// described. Those are removed; this is the contracted path.
type WorkScheduleAPI struct {
	Service *application.Service
}

func NewWorkScheduleAPI(service *application.Service) *WorkScheduleAPI {
	return &WorkScheduleAPI{Service: service}
}

// scheduleShape is the JSON stored in hr_work_schedules.schedule_data.
//
// The column is opaque JSON and nothing has ever defined its keys: the previous
// create handler took whatever string a caller sent. The contract, meanwhile,
// declares these seven fields individually. This type is where the two meet, so
// the shape is written down in exactly one place rather than assumed at each
// call site.
//
// Every field is omitempty and every reader tolerates absence, because rows
// written before this existed carry arbitrary or empty JSON.
type scheduleShape struct {
	WorkingDays              []string `json:"working_days,omitempty"`
	StartTime                string   `json:"start_time,omitempty"`
	EndTime                  string   `json:"end_time,omitempty"`
	BreakMinutes             int32    `json:"break_minutes,omitempty"`
	ContractedHoursPerWeek   float64  `json:"contracted_hours_per_week,omitempty"`
	DailyOTThresholdMinutes  int32    `json:"daily_ot_threshold_minutes,omitempty"`
	WeeklyOTThresholdMinutes int32    `json:"weekly_ot_threshold_minutes,omitempty"`
}

// decodeScheduleShape reads schedule_data, tolerating anything it cannot parse.
//
// A row whose JSON predates this shape yields zero values rather than an error.
// That is deliberate: a schedule that cannot report its contracted hours must
// read as "unknown" to the consumer, and a failed list is worse than a schedule
// with blank fields. The distinction between "zero hours" and "unknown hours"
// belongs to the consumer, which is why nothing is defaulted to 40 here.
func decodeScheduleShape(raw string) scheduleShape {
	var shape scheduleShape
	if strings.TrimSpace(raw) == "" {
		return shape
	}
	if err := json.Unmarshal([]byte(raw), &shape); err != nil {
		return scheduleShape{}
	}

	return shape
}

func encodeScheduleShape(shape scheduleShape) (string, error) {
	encoded, err := json.Marshal(shape)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func workScheduleToProto(ws domain.WorkScheduleRecord) *workschedulev1.WorkSchedulePayload {
	shape := decodeScheduleShape(ws.ScheduleData)

	return &workschedulev1.WorkSchedulePayload{
		Id:                       ws.ID,
		OwnerId:                  ws.OwnerID,
		Name:                     ws.Name,
		WorkingDays:              shape.WorkingDays,
		StartTime:                shape.StartTime,
		EndTime:                  shape.EndTime,
		BreakMinutes:             shape.BreakMinutes,
		ContractedHoursPerWeek:   shape.ContractedHoursPerWeek,
		DailyOtThresholdMinutes:  shape.DailyOTThresholdMinutes,
		WeeklyOtThresholdMinutes: shape.WeeklyOTThresholdMinutes,
		CreatedAt:                ws.CreatedAt.Unix(),
		UpdatedAt:                ws.UpdatedAt.Unix(),
	}
}

func resolveWorkScheduleBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}

	return businessID, nil
}

// clampLimit keeps the page bounded whatever the caller asks for.
func clampLimit(requested int32) int32 {
	if requested <= 0 {
		return defaultWorkScheduleLimit
	}
	if requested > maxWorkScheduleLimit {
		return maxWorkScheduleLimit
	}

	return requested
}

func (a *WorkScheduleAPI) ListWorkSchedules(
	ctx context.Context, req *workschedulev1.ListWorkSchedulesRequest,
) (*workschedulev1.ListWorkSchedulesResponse, error) {
	businessID, err := resolveWorkScheduleBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	limit := clampLimit(req.GetLimit())
	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	schedules, err := a.Service.ListWorkSchedules(ctx, businessID, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list work schedules",
			logger.Any("business_id", businessID), logger.Error(err))

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list work schedules")
	}

	total, err := a.Service.CountWorkSchedules(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count work schedules",
			logger.Any("business_id", businessID), logger.Error(err))

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list work schedules")
	}

	payloads := make([]*workschedulev1.WorkSchedulePayload, 0, len(schedules))
	for _, ws := range schedules {
		payloads = append(payloads, workScheduleToProto(ws))
	}

	return &workschedulev1.ListWorkSchedulesResponse{
		Schedules: payloads,
		Total:     int32(total), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}, nil
}

func (a *WorkScheduleAPI) CreateWorkSchedule(
	ctx context.Context, req *workschedulev1.CreateWorkScheduleRequest,
) (*workschedulev1.CreateWorkScheduleResponse, error) {
	businessID, err := resolveWorkScheduleBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	scheduleData, err := encodeScheduleShape(scheduleShape{
		WorkingDays:              req.GetWorkingDays(),
		StartTime:                req.GetStartTime(),
		EndTime:                  req.GetEndTime(),
		BreakMinutes:             req.GetBreakMinutes(),
		ContractedHoursPerWeek:   req.GetContractedHoursPerWeek(),
		DailyOTThresholdMinutes:  req.GetDailyOtThresholdMinutes(),
		WeeklyOTThresholdMinutes: req.GetWeeklyOtThresholdMinutes(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create work schedule")
	}

	created, err := a.Service.CreateWorkSchedule(ctx, businessID, domain.WorkScheduleRecord{
		Name:         req.GetName(),
		ScheduleData: scheduleData,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create work schedule",
			logger.Any("business_id", businessID), logger.Error(err))

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create work schedule")
	}

	return &workschedulev1.CreateWorkScheduleResponse{Schedule: workScheduleToProto(*created)}, nil
}

func (a *WorkScheduleAPI) UpdateWorkSchedule(
	ctx context.Context, req *workschedulev1.UpdateWorkScheduleRequest,
) (*workschedulev1.UpdateWorkScheduleResponse, error) {
	businessID, err := resolveWorkScheduleBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "body is required")
	}

	// The update body is a full replacement of the shape: every field on it is
	// optional in the contract, but schedule_data is one column, so a partial
	// write would silently drop the fields the caller omitted. Read-modify-write
	// would need a Get RPC that does not exist; until it does, the contract's
	// "absent to keep current" cannot be honoured field by field, and pretending
	// otherwise would lose data. Callers send the whole shape.
	scheduleData, err := encodeScheduleShape(scheduleShape{
		WorkingDays:              body.GetWorkingDays(),
		StartTime:                body.GetStartTime(),
		EndTime:                  body.GetEndTime(),
		BreakMinutes:             body.GetBreakMinutes(),
		ContractedHoursPerWeek:   body.GetContractedHoursPerWeek(),
		DailyOTThresholdMinutes:  body.GetDailyOtThresholdMinutes(),
		WeeklyOTThresholdMinutes: body.GetWeeklyOtThresholdMinutes(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update work schedule")
	}

	updated, err := a.Service.UpdateWorkSchedule(ctx, businessID, domain.WorkScheduleRecord{
		ID:           req.GetId(),
		Name:         body.GetName(),
		ScheduleData: scheduleData,
	})
	if err != nil {
		if errors.Is(err, domain.ErrWorkScheduleNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Work schedule not found")
		}
		logger.Ctx(ctx).Error("failed to update work schedule",
			logger.Any("business_id", businessID), logger.Any("schedule_id", req.GetId()), logger.Error(err))

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update work schedule")
	}

	return &workschedulev1.UpdateWorkScheduleResponse{Schedule: workScheduleToProto(*updated)}, nil
}
