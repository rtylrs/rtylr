package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/reports/export/revenue?from=&to=
//
// Streams a text/csv response of daily paid revenue. This handler bypasses
// the generic JSON `RegisterGET` helper because the wire format is CSV,
// not JSON — see hr_reports/GenerateReportHandler for the same pattern.

type ExportRevenueHandler struct {
	Service *application.Service
}

func ExportRevenueRoutes() []string { return []string{"/insights/reports/export/revenue"} }

func (h *ExportRevenueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		http.Error(w, errorModels.PermissionDeniedServerError().Message, http.StatusForbidden)
		return
	}

	from, to, err := application.ResolveDateRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, `{"error":"invalid date range"}`, http.StatusBadRequest)
		return
	}

	series, err := h.Service.GetRevenueSeries(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: export revenue failed", logger.Error(err))
		http.Error(w, `{"error":"failed to load revenue"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(
		`attachment; filename="revenue_%s_%s.csv"`,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
	))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if writeErr := writer.Write([]string{"date", "amount_minor_units"}); writeErr != nil {
		logger.Ctx(ctx).Error("insights: csv header write failed", logger.Error(writeErr))
		return
	}
	for _, point := range series {
		if writeErr := writer.Write([]string{
			point.Date,
			strconv.FormatInt(point.AmountMinorUnits, 10),
		}); writeErr != nil {
			logger.Ctx(ctx).Error("insights: csv row write failed", logger.Error(writeErr))
			return
		}
	}
}
