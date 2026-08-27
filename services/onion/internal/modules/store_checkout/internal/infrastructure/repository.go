package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/platform/logger"
)

// ensure domain is used.
var _ = domain.Checkout{}

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) Create(ctx context.Context, businessID, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin checkout transaction", logger.Error(err))
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	qtx := r.queries.WithTx(tx)
	input, err = r.priceCheckoutInput(ctx, qtx, ownerID, input)
	if err != nil {
		return 0, err
	}

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Email: input.CustomerEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync checkout customer", logger.Error(err))
		return 0, err
	}

	result, err := qtx.CreateCheckout(ctx, sqlc.CreateCheckoutParams{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		SessionID:     sql.NullString{String: input.SessionID, Valid: input.SessionID != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		Items:         nullJSON(input.Items),
		Subtotal:      strconv.FormatFloat(input.Subtotal, 'f', 2, 64),
		Tax:           strconv.FormatFloat(input.Tax, 'f', 2, 64),
		Total:         strconv.FormatFloat(input.Total, 'f', 2, 64),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create checkout", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read checkout insert id", logger.Error(err))
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit checkout transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

type checkoutItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
	Qty       int   `json:"qty"`
}

type pricedCheckoutItem struct {
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name"`
	Price     string `json:"price"`
	Total     string `json:"total"`
}

func (r *Repository) priceCheckoutInput(ctx context.Context, qtx *sqlc.Queries, ownerID int64, input domain.CheckoutCreateInput) (domain.CheckoutCreateInput, error) {
	if len([]byte(input.Items)) == 0 || input.Items == "null" || input.Items == "[]" {
		return input, nil
	}
	var requested []checkoutItem
	if err := json.Unmarshal([]byte(input.Items), &requested); err != nil {
		return domain.CheckoutCreateInput{}, fmt.Errorf("invalid checkout items: %w", err)
	}
	var subtotalCents, taxCents int64
	priced := make([]pricedCheckoutItem, 0, len(requested))
	for _, item := range requested {
		quantity := item.Quantity
		if quantity == 0 {
			quantity = item.Qty
		}
		if item.ProductID <= 0 || quantity <= 0 {
			return domain.CheckoutCreateInput{}, errors.New("checkout items must reference positive catalog products and quantities")
		}
		product, err := qtx.GetCheckoutProductPricing(ctx, sqlc.GetCheckoutProductPricingParams{OwnerID: ownerID, ID: item.ProductID})
		if err != nil {
			return domain.CheckoutCreateInput{}, fmt.Errorf("load checkout product %d: %w", item.ProductID, err)
		}
		unitCents, err := decimalCents(product.SalePrice)
		if err != nil || unitCents < 0 {
			return domain.CheckoutCreateInput{}, fmt.Errorf("invalid catalog price for product %d", item.ProductID)
		}
		taxRate, err := strconv.ParseFloat(product.TaxRate, 64)
		if err != nil || taxRate < 0 || math.IsNaN(taxRate) || math.IsInf(taxRate, 0) {
			return domain.CheckoutCreateInput{}, fmt.Errorf("invalid catalog tax for product %d", item.ProductID)
		}
		lineCents := unitCents * int64(quantity)
		subtotalCents += lineCents
		taxCents += int64(math.Round(float64(lineCents) * taxRate / 100))
		priced = append(priced, pricedCheckoutItem{
			ProductID: item.ProductID,
			Quantity:  quantity,
			Name:      product.Name,
			Price:     centsString(unitCents),
			Total:     centsString(lineCents),
		})
	}
	encoded, err := json.Marshal(priced)
	if err != nil {
		return domain.CheckoutCreateInput{}, err
	}
	input.Items = string(encoded)
	input.Subtotal = domain.CentsToFloat(subtotalCents)
	input.Tax = domain.CentsToFloat(taxCents)
	input.Total = domain.CentsToFloat(subtotalCents + taxCents)
	return input, nil
}

func decimalCents(value string) (int64, error) {
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, errors.New("invalid decimal amount")
	}
	return int64(math.Round(amount * 100)), nil
}

func centsString(cents int64) string {
	return strconv.FormatFloat(float64(cents)/100, 'f', 2, 64)
}

func (r *Repository) Complete(ctx context.Context, businessID, id, _ int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin checkout completion transaction", logger.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var customerEmail sql.NullString
	var totalString string
	checkout, err := r.queries.WithTx(tx).GetCheckoutForComplete(ctx, sqlc.GetCheckoutForCompleteParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// Completion is idempotent for missing or cross-business checkouts.
		// The scoped lookup deliberately reveals neither existence nor data.
		return nil
	} else if err != nil {
		logger.Ctx(ctx).Error("Failed to load checkout customer", logger.Error(err))
		return err
	} else {
		if checkout.Status.Valid && checkout.Status.StoreCheckoutsStatus == sqlc.StoreCheckoutsStatus("completed") {
			return nil
		}
		customerEmail = checkout.CustomerEmail
		totalString = checkout.Total
		total, parseErr := strconv.ParseFloat(totalString, 64)
		if parseErr != nil {
			logger.Ctx(ctx).Warn("failed to parse checkout total for CRM sync",
				logger.String("total", totalString),
				logger.Error(parseErr),
			)
		} else if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
			Email:      customerEmail.String,
			TotalSpent: total,
			CountVisit: true,
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to sync completed checkout customer", logger.Error(err))
			return err
		}
	}

	err = r.queries.WithTx(tx).CompleteCheckout(ctx, sqlc.CompleteCheckoutParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to complete checkout", logger.Error(err))
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit checkout completion transaction", logger.Error(err))
		return err
	}
	return nil
}

func nullJSON(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}
