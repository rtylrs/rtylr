package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GenerateReportRoutes() []string { return []string{"/hr/reports/generate"} }

type GenerateReportHandler struct {
	Service     *application.Service
	Permissions PermissionChecker
}

type PermissionChecker interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
}

type GenerateReportRequest struct {
	ReportType   string `json:"report_type"`
	DateFrom     string `json:"date_from"`
	DateTo       string `json:"date_to"`
	LocationID   *int64 `json:"location_id"`
	DepartmentID *int64 `json:"department_id"`
	EmployeeID   *int64 `json:"employee_id"`
	Format       string `json:"format"`
}

// This handler sets Content-Type manually for CSV output — an intentional
// exception to the standard JSON-only response pattern. It is registered
// directly on the router (bypassing wrapHandler) to support dual-format output.
func (h *GenerateReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		writeError(w, onekiterrors.PermissionDenied())
		return
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		writeError(w, onekiterrors.PermissionDenied())
		return
	}
	if h.Permissions == nil {
		writeError(w, onekiterrors.PermissionDenied())
		return
	}
	employeeID, employeeErr := h.Permissions.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if employeeErr != nil {
		writeError(w, onekiterrors.PermissionDenied())
		return
	}
	allowed, permissionErr := h.Permissions.HasPermission(ctx, businessID, employeeID, "view_reports")
	if permissionErr != nil || !allowed {
		writeError(w, onekiterrors.PermissionDenied())
		return
	}

	var req GenerateReportRequest
	if err := serverHttp.DecodeJSONRequestBody(w, r, &req); err != nil {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "Invalid request body"))
		return
	}

	req.ReportType = strings.TrimSpace(req.ReportType)
	if req.ReportType == "" {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "report_type is required"))
		return
	}

	req.DateFrom = strings.TrimSpace(req.DateFrom)
	if req.DateFrom == "" {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "date_from is required"))
		return
	}

	req.DateTo = strings.TrimSpace(req.DateTo)
	if req.DateTo == "" {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "date_to is required"))
		return
	}

	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "Invalid date_from format, expected ISO-8601 date (YYYY-MM-DD)"))
		return
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "Invalid date_to format, expected ISO-8601 date (YYYY-MM-DD)"))
		return
	}

	if dateTo.Before(dateFrom) {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "date_to must not be before date_from"))
		return
	}
	if dateTo.Sub(dateFrom) > domain.MaxInteractiveReportWindow {
		writeError(w, onekiterrors.New(http.StatusBadRequest, "report date range must not exceed 366 days"))
		return
	}

	params := domain.ReportParams{
		ActorID:      device.User.UserID,
		OwnerID:      device.User.UserID,
		BusinessID:   businessID,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		LocationID:   req.LocationID,
		DepartmentID: req.DepartmentID,
		EmployeeID:   req.EmployeeID,
		Format:       req.Format,
	}

	rows, serviceErr := h.Service.GenerateReport(ctx, businessID, req.ReportType, params)
	if serviceErr != nil {
		if errors.Is(serviceErr, domain.ErrUnsupportedReportType) {
			writeError(w, onekiterrors.New(http.StatusUnprocessableEntity, "Unsupported report type"))
			return
		}
		if errors.Is(serviceErr, domain.ErrReportGenerationFailed) {
			logger.Ctx(ctx).Error("report generation failed",
				logger.Any("owner_id", device.User.UserID),
				logger.Any("report_type", req.ReportType),
				logger.Error(serviceErr))
			writeError(w, onekiterrors.New(http.StatusInternalServerError, "Report generation failed"))
			return
		}
		if errors.Is(serviceErr, domain.ErrReportTooLarge) {
			writeError(w, onekiterrors.New(http.StatusUnprocessableEntity, "Report is too large; narrow the date range or filters"))
			return
		}
		logger.Ctx(ctx).Error("failed to generate report",
			logger.Any("owner_id", device.User.UserID),
			logger.Any("report_type", req.ReportType),
			logger.Error(serviceErr))
		writeError(w, onekiterrors.New(http.StatusInternalServerError, "Failed to generate report"))
		return
	}

	if req.Format == "csv" {
		writeCSV(w, rows)
		return
	}

	writeJSON(w, rows)
}

func writeCSV(w http.ResponseWriter, rows []map[string]interface{}) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="report.csv"`)
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if len(rows) == 0 {
		return
	}

	headers := make([]string, 0, len(rows[0]))
	for key := range rows[0] {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	if err := writer.Write(headers); err != nil {
		return
	}

	for _, row := range rows {
		record := make([]string, 0, len(headers))
		for _, h := range headers {
			record = append(record, sanitizeCSVCell(fmt.Sprintf("%v", row[h])))
		}
		if err := writer.Write(record); err != nil {
			return
		}
	}
}

func sanitizeCSVCell(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

func writeJSON(w http.ResponseWriter, rows []map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if rows == nil {
		rows = []map[string]interface{}{}
	}

	if err := json.NewEncoder(w).Encode(rows); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, serverError error) {
	serverHttp.WriteOneKitError(context.Background(), w, serverError)
}
