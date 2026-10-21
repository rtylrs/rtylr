package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/receipts/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

// Receipts are effectively immutable once printed/paid; cache aggressively.
// Past-date list queries piggyback on the orders namespace so any order/
// refund mutation can invalidate them via OrderPrefix.
const receiptsCacheTTL = 5 * time.Minute

type repository interface {
	List(ctx context.Context, businessID int64, dateFrom, dateTo string) ([]domain.Receipt, error)
	Get(ctx context.Context, businessID, id int64) (*domain.Receipt, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64, dateFrom, dateTo string) ([]domain.Receipt, error) {
	key := fmt.Sprintf("%s|receipts|f=%s|t=%s", cacheKeys.OrdersList(businessID), dateFrom, dateTo)
	return cacher.Of[[]domain.Receipt]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.Receipt, error) {
			return s.repo.List(ctx, businessID, dateFrom, dateTo)
		},
		cacher.WithExpiry(ctx, receiptsCacheTTL),
	)
}

func (s *Service) Get(ctx context.Context, businessID, id int64) (*domain.Receipt, error) {
	key := fmt.Sprintf("%s|receipt=%d", cacheKeys.OrdersList(businessID), id)
	return cacher.Of[*domain.Receipt]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.Receipt, error) {
			return s.repo.Get(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, receiptsCacheTTL),
	)
}

// EmailReceipt emails a previously-generated receipt to the provided address.
// Returns ErrReceiptNotFound when the receipt does not exist or is not owned by businessID.
func (s *Service) EmailReceipt(ctx context.Context, businessID, id int64, toEmail string) error {
	receipt, err := s.repo.Get(ctx, businessID, id)
	if err != nil {
		return err
	}

	subject := "Your receipt for order " + receipt.OrderNumber
	html := fmt.Sprintf(
		"<p>Hello %s,</p><p>Thank you for your order <strong>%s</strong>.</p>"+
			"<p>Total: <strong>%.2f</strong></p>"+
			"<p>Paid on %s via %s.</p>",
		receipt.CustomerName, receipt.OrderNumber, receipt.Total,
		receipt.PaidAt.Format("2006-01-02 15:04"), receipt.PaymentMethod,
	)
	text := fmt.Sprintf(
		"Receipt for order %s\nTotal: %.2f\nPaid on %s via %s\n",
		receipt.OrderNumber, receipt.Total,
		receipt.PaidAt.Format("2006-01-02 15:04"), receipt.PaymentMethod,
	)

	_, sendErr := emailPkg.Send(ctx, emailPkg.SendParams{
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
	return sendErr
}
