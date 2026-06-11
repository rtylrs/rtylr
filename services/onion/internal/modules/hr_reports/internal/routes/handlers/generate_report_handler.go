package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_reports/internal/application"
	"github.com/voxire/onion/internal/modules/hr_reports/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func GenerateReportRoutes() []string { return []string{"/hr/reports/generate"} }

type GenerateReportHandler struct {
	Service *application.Service
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
		writeError(w, errorModels.PermissionDeniedServerError())
		return
	}

	var req GenerateReportRequest
	if err := serverHttp.DecodeJSONRequestBody(w, r, &req); err != nil {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid request body",
		})
		return
	}

	req.ReportType = strings.TrimSpace(req.ReportType)
	if req.ReportType == "" {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "report_type is required",
		})
		return
	}

	req.DateFrom = strings.TrimSpace(req.DateFrom)
	if req.DateFrom == "" {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "date_from is required",
		})
		return
	}

	req.DateTo = strings.TrimSpace(req.DateTo)
	if req.DateTo == "" {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "date_to is required",
		})
		return
	}

	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date_from format, expected ISO-8601 date (YYYY-MM-DD)",
		})
		return
	}

	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date_to format, expected ISO-8601 date (YYYY-MM-DD)",
		})
		return
	}

	if dateTo.Before(dateFrom) {
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "date_to must not be before date_from",
		})
		return
	}

	params := domain.ReportParams{
		OwnerID:      device.User.UserID,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		LocationID:   req.LocationID,
		DepartmentID: req.DepartmentID,
		EmployeeID:   req.EmployeeID,
		Format:       req.Format,
	}

	rows, serviceErr := h.Service.GenerateReport(ctx, device.User.UserID, req.ReportType, params)
	if serviceErr != nil {
		if errors.Is(serviceErr, domain.ErrUnsupportedReportType) {
			writeError(w, &errorModels.ServerError{
				HTTPCode: http.StatusUnprocessableEntity,
				Code:     http.StatusUnprocessableEntity,
				Message:  "Unsupported report type",
			})
			return
		}
		if errors.Is(serviceErr, domain.ErrReportGenerationFailed) {
			logger.Ctx(ctx).Error("report generation failed",
				logger.Any("owner_id", device.User.UserID),
				logger.Any("report_type", req.ReportType),
				logger.Error(serviceErr))
			writeError(w, &errorModels.ServerError{
				HTTPCode: http.StatusInternalServerError,
				Code:     http.StatusInternalServerError,
				Message:  "Report generation failed",
			})
			return
		}
		logger.Ctx(ctx).Error("failed to generate report",
			logger.Any("owner_id", device.User.UserID),
			logger.Any("report_type", req.ReportType),
			logger.Error(serviceErr))
		writeError(w, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to generate report",
		})
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
			record = append(record, fmt.Sprintf("%v", row[h]))
		}
		if err := writer.Write(record); err != nil {
			return
		}
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

func writeError(w http.ResponseWriter, serverError *errorModels.ServerError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(serverError.HTTPCode)

	response := errorModels.ErrorResponse{
		Error:  *serverError,
		Status: "error",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
