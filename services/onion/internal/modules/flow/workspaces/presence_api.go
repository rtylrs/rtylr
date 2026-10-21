package workspaces

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowpresencev1 "github.com/voxire/voxie/gen/go/flow/presence/v1"
)

// PresenceAPI serves the contracted Flow presence endpoint. The presence
// aggregate depends on HR modules owned by this package, so its adapter lives
// beside Module rather than in the workspace application API subpackage.
type PresenceAPI struct {
	Module *Module
}

func presenceEntriesToProto(entries []PresenceEntry) []*flowpresencev1.PresenceEntry {
	responseEntries := make([]*flowpresencev1.PresenceEntry, 0, len(entries))
	for _, entry := range entries {
		firstName := entry.FirstName
		lastName := entry.LastName
		responseEntries = append(responseEntries, &flowpresencev1.PresenceEntry{
			EmployeeId: strconv.FormatInt(entry.EmployeeID, 10),
			UserId:     strconv.FormatInt(entry.UserID, 10),
			FirstName:  &firstName,
			LastName:   &lastName,
			Status:     entry.Status,
		})
	}
	return responseEntries
}

func (a *PresenceAPI) GetPresence(
	ctx context.Context,
	_ *flowpresencev1.GetPresenceRequest,
) (*flowpresencev1.GetPresenceResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	resolvedOwnerID, err := tenant.ResolveOwnerID(ctx, businessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve business owner")
	}

	date := time.Now().UTC().Format("2006-01-02")
	entries, err := a.Module.ListPresenceForBusiness(ctx, resolvedOwnerID, businessID, device.User.UserID, date)
	if err != nil {
		if errors.Is(err, errPresenceAccessDenied) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load presence")
	}

	return &flowpresencev1.GetPresenceResponse{
		Presence: presenceEntriesToProto(entries),
		Date:     date,
	}, nil
}
