package api

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
)

func resolveBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func resolveBusinessIDFor(ctx context.Context, permission string) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppInsights, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	return resolveBusinessID(ctx)
}

func resolvePeriodRange(period string) (string, string, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		return "", "", nil
	}
	parts := strings.SplitN(period, ",", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
	}
	return period, "", nil
}

func resolveDateRangeFromPeriod(period string) (time.Time, time.Time, error) {
	fromStr, toStr, err := resolvePeriodRange(period)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	from, to, err := application.ResolveDateRange(fromStr, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, onekiterrors.New(http.StatusBadRequest, "Invalid period")
	}
	return from, to, nil
}

func minorToMajor(v int64) float64 {
	return float64(v) / 100
}

func normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizePage(page int32) int32 {
	if page <= 0 {
		return 1
	}
	return page
}

func paginationResponse(page, limit int32, total int64) *commonpaginationv1.PaginationResponse {
	page = normalizePage(page)
	limit = normalizeLimit(limit)
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}
	if total > math.MaxInt32 {
		total = math.MaxInt32
	}
	if totalPages > math.MaxInt32 {
		totalPages = math.MaxInt32
	}
	return &commonpaginationv1.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int32(total),
		TotalPages: int32(totalPages),
		HasNext:    int64(page)*int64(limit) < total,
		HasPrev:    page > 1,
	}
}

func notImplemented(feature string) error {
	return onekiterrors.New(http.StatusNotImplemented, feature+" is not available on the OneKit path yet")
}
