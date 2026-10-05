package api

import (
	"math"

	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
)

// Page bounds for the developer listings. A delivery listing is something an
// owner scans for failures, so the default is small enough to read and the cap
// stops a client asking for the whole history in one request.
const (
	defaultDeliveryPageLimit = int32(50)
	maxDeliveryPageLimit     = int32(200)
)

// resolveDeliveryPagination normalises the requested page and limit.
//
// Same shape as the workspaces helper, including the offset clamp: repositories
// narrow the offset to int32 for sqlc, so an extreme page number could
// otherwise wrap negative and either alias to a wrong page or be rejected by
// the server.
func resolveDeliveryPagination(page, limit int32) (resolvedPage, resolvedLimit int32, offset int64) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = defaultDeliveryPageLimit
	}
	if limit > maxDeliveryPageLimit {
		limit = maxDeliveryPageLimit
	}
	offset = int64(page-1) * int64(limit)
	if offset > math.MaxInt32 {
		offset = math.MaxInt32
	}
	return page, limit, offset
}

func buildDeliveryPagination(page, limit int32, total int64, returned int) *commonpaginationv1.PaginationResponse {
	var totalPages int32
	if limit > 0 {
		totalPages = int32((total + int64(limit) - 1) / int64(limit)) // #nosec G115 -- bounded by the page limit above.
	}
	return &commonpaginationv1.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int32(total), // #nosec G115 -- bounded by the page limit above.
		TotalPages: totalPages,
		HasNext:    int64(page-1)*int64(limit)+int64(returned) < total,
		HasPrev:    page > 1,
	}
}
