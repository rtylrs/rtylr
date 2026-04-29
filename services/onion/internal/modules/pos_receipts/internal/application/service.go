package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

type repository interface {
	List(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]domain.Receipt, error)
	Get(ctx context.Context, ownerID, id int64) (*domain.Receipt, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]domain.Receipt, error) {
	return s.repo.List(ctx, ownerID, dateFrom, dateTo)
}

func (s *Service) Get(ctx context.Context, ownerID, id int64) (*domain.Receipt, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// EmailReceipt emails a previously-generated receipt to the provided address.
// Returns ErrReceiptNotFound when the receipt does not exist or is not owned by ownerID.
func (s *Service) EmailReceipt(ctx context.Context, ownerID, id int64, toEmail string) error {
	receipt, err := s.repo.Get(ctx, ownerID, id)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Your receipt for order %s", receipt.OrderNumber)
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
