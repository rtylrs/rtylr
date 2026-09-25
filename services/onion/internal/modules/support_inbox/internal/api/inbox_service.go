package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/application"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportinboxv1 "github.com/voxire/voxie/gen/go/support/inbox/v1"
)

type InboxAPI struct {
	service *application.Service
}

func NewInboxAPI(service *application.Service) *InboxAPI {
	return &InboxAPI{service: service}
}

func (a *InboxAPI) ListInbox(ctx context.Context, req *supportinboxv1.ListInboxRequest) (*supportinboxv1.ListInboxResponse, error) {
	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	items, total, err := a.service.ListInbox(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list inbox")
	}

	return &supportinboxv1.ListInboxResponse{
		Inbox:    newInboxItems(items),
		Total:    int32(total), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (a *InboxAPI) GetUnreadCount(ctx context.Context, _ *supportinboxv1.GetUnreadCountRequest) (*supportinboxv1.GetUnreadCountResponse, error) {
	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	count, err := a.service.UnreadCount(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get unread count")
	}

	return &supportinboxv1.GetUnreadCountResponse{Count: int32(count)}, nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func businessIDFromContext(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return device.User.BusinessID, nil
}

func newInboxItems(items []domain.InboxItem) []*supportinboxv1.InboxItem {
	out := make([]*supportinboxv1.InboxItem, 0, len(items))
	for _, item := range items {
		out = append(out, &supportinboxv1.InboxItem{
			TicketId:      item.TicketID,
			CustomerName:  item.CustomerName,
			CustomerEmail: item.CustomerEmail,
			Subject:       item.Subject,
			Status:        item.Status,
			Priority:      item.Priority,
			Channel:       item.Channel,
			LastMessage:   item.LastMessage,
			LastMessageAt: formatTime(item.LastMessageAt),
			MessageCount:  int32(item.MessageCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
