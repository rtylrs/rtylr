package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	pendingToolApprovalTTL  = 5 * time.Minute
	maxPendingToolApprovals = 512
)

var errPendingToolApproval = errors.New("tool approval is missing, expired, already used, or does not match the proposed call")

type pendingToolApproval struct {
	Name             string
	InputFingerprint string
	Mode             CaretMode
	BusinessID       int64
	UserID           int64
	ExpiresAt        time.Time
}

var pendingToolApprovals = struct {
	sync.Mutex
	items map[string]pendingToolApproval
}{items: make(map[string]pendingToolApproval)}

func RegisterPendingToolApproval(ctx context.Context, id string, tool ToolDef, input map[string]any, mode CaretMode) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errPendingToolApproval
	}
	businessID, userID := AuthenticatedIdentity(ctx)
	now := time.Now()
	pendingToolApprovals.Lock()
	defer pendingToolApprovals.Unlock()
	for key, item := range pendingToolApprovals.items {
		if now.After(item.ExpiresAt) {
			delete(pendingToolApprovals.items, key)
		}
	}
	if len(pendingToolApprovals.items) >= maxPendingToolApprovals {
		return errPendingToolApproval
	}
	pendingToolApprovals.items[id] = pendingToolApproval{
		Name:             tool.Name,
		InputFingerprint: ToolCallFingerprint(tool.Name, input),
		Mode:             mode,
		BusinessID:       businessID,
		UserID:           userID,
		ExpiresAt:        now.Add(pendingToolApprovalTTL),
	}
	return nil
}

func ConsumePendingToolApproval(ctx context.Context, id string, tool ToolDef, input map[string]any, mode CaretMode) error {
	id = strings.TrimSpace(id)
	now := time.Now()
	businessID, userID := AuthenticatedIdentity(ctx)
	pendingToolApprovals.Lock()
	defer pendingToolApprovals.Unlock()
	item, ok := pendingToolApprovals.items[id]
	if !ok || now.After(item.ExpiresAt) {
		delete(pendingToolApprovals.items, id)
		return errPendingToolApproval
	}
	if item.Name != tool.Name || item.InputFingerprint != ToolCallFingerprint(tool.Name, input) || (mode != "" && item.Mode != mode) {
		return errPendingToolApproval
	}
	// Bind on the user id unconditionally: a caller whose business_id is
	// unset (0) must still not be able to consume someone else's approval,
	// so businessID is an additional check, never a gate.
	if userID != item.UserID || (item.BusinessID > 0 && businessID != item.BusinessID) {
		return errPendingToolApproval
	}
	delete(pendingToolApprovals.items, id)
	return nil
}

func AuthenticatedIdentity(ctx context.Context) (businessID, userID int64) {
	if ctx == nil {
		return 0, 0
	}
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil {
		return device.User.BusinessID, device.User.UserID
	}
	if user, ok := authPkg.GetAuthUser(ctx); ok && user != nil {
		return user.BusinessID, user.UserID
	}
	return 0, 0
}
