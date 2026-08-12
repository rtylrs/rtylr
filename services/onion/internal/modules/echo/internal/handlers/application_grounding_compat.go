package handlers

import (
	"context"
	"time"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
)

type analystGroundingResult = echoapp.AnalystGroundingResult

func buildBusinessSnapshotGrounding(ctx context.Context, analyst AnalystService, question string, businessID, userID int64, now time.Time) (*analystGroundingResult, bool) {
	return echoapp.BuildBusinessSnapshotGrounding(ctx, analyst, question, businessID, userID, now)
}
