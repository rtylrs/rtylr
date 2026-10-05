package application

import "context"

type InsightsMemberProfile struct {
	UserID int64
	Name   string
}

// InsightsMemberProfileSource resolves display names through the identity owner.
type InsightsMemberProfileSource interface {
	GetInsightsMemberProfiles(ctx context.Context, userIDs []int64) (map[int64]InsightsMemberProfile, error)
}
