package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
)

type deliveryRepository interface {
	ListDueDeliveries(context.Context, int32) ([]domain.Delivery, error)
	ClaimDelivery(context.Context, uint64) error
	MarkDelivered(context.Context, uint64, int) error
	MarkFailed(context.Context, uint64, int64, int, string) error
	MarkDead(context.Context, uint64, string) error
	RecoverAbandoned(context.Context, int64) (int64, error)
	GetRegistrationForDelivery(context.Context, int64, int64) (domain.Registration, error)
	RecordTriggered(context.Context, int64, int64) error
}

type eventReader interface {
	GetEvent(context.Context, uint64, int64) (eventbus.Event, error)
}

type notificationRepository interface {
	ListFailingRegistrations(context.Context, int64, int32, time.Time) ([]domain.FailingRegistration, error)
	GetNotificationOwner(context.Context, int64, int64) (int64, error)
	RecordDeadDeliveryNotice(context.Context, int64, int64, time.Time) error
}

type notificationContactResolver func(context.Context, int64, int64) (string, string, bool, error)

type NotificationContactResolver = notificationContactResolver

type NotificationEmail struct {
	To        string
	Subject   string
	HTML      string
	Text      string
	DedupeKey string
}

type notificationQueue interface {
	Enqueue(context.Context, NotificationEmail) error
}
