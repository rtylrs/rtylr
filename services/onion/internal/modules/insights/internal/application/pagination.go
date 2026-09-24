package application

import "math"

func normalizeListLimit(limit int32) int32 {
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

func pageOffset(page, limit int32) int32 {
	if page > (math.MaxInt32-int32(limit-1))/int32(limit) {
		return math.MaxInt32 - int32(limit) + 1
	}
	return (page - 1) * limit
}
