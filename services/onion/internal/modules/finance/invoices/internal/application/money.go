package application

import (
	"fmt"
	"math"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

// Invoice money arithmetic, shared by the create and the update path so the
// two can never drift apart again.
//
// ORDER OF OPERATIONS - do not re-order. Each step feeds only the next one:
//
//	 1. line gross   = quantity x unit price
//	 2. line discount = line gross x line discount_pct / 100
//	 3. line base    = line gross - line discount      (step 2, so a line's own
//	                                                    tax never taxes a discount)
//	 4. subtotal     = sum of the line bases          (pre-tax, post line discount)
//	 5. header disc. = subtotal x discount_pct / 100, or a fixed amount
//	 6. taxable base = subtotal - header discount      (after discount, before tax)
//	 7. line taxable = each line base scaled down to its pro-rata share of the
//	                  taxable base, so a line is taxed only on what is charged
//	 8. line tax     = line taxable x line tax_rate / 100
//	 9. header tax   = taxable base x header tax_rate / 100
//	10. tax amount   = sum of the line taxes + the header tax
//	11. total        = taxable base + tax amount      (never grossed up again)
//
// The header-rate tax is a header-only layer: it has no line of its own, so
// the line rows foot to (taxable base + line tax) and the header adds the
// header tax on top. Every line is taxed exactly once, and no layer is ever
// applied to a base that already contains another layer.
//
// ROUNDING POLICY - per line and per header amount, to whole cents, half away
// from zero, applied where the amount is produced via the platform money
// helpers. This is the same rounding the payment path uses
// (infrastructure.paymentAmountToCents), which is what lets a payment of
// AmountDue settle an invoice to exactly zero. Nothing is deferred to display
// time, so the stored total, the emailed total and the total summed into tax
// reports are the same number.

// roundMoney snaps a value to whole cents using the platform money helpers.
func roundMoney(value float64) float64 {
	return money.CentsToDollars(money.DollarsToCents(value))
}

// invoiceTotals is a fully rounded invoice money summary. LineTaxable and
// LineTax are parallel to the input items.
type invoiceTotals struct {
	Subtotal       float64
	DiscountAmount float64
	TaxAmount      float64
	Total          float64
	LineTaxable    []float64
	LineTax        []float64
}

// LineTotal is the line's net charge: its taxable base plus its own tax. Both
// parts are already whole cents, so this needs no further rounding.
func (t invoiceTotals) LineTotal(i int) float64 {
	return roundMoney(t.LineTaxable[i] + t.LineTax[i])
}

// computeInvoiceTotals runs the order of operations above. It returns an error
// for input that would persist a meaningless amount rather than clamping it:
// negative money, non-finite numbers, or rates outside 0..100.
func computeInvoiceTotals(items []domain.InvoiceItemInput, headerTaxRate float64, discountType string, discountValue float64) (invoiceTotals, error) {
	if err := validateMoneyInput(items, headerTaxRate, discountType, discountValue); err != nil {
		return invoiceTotals{}, err
	}

	// Steps 1-4.
	taxable := make([]float64, len(items))
	subtotal := 0.0
	for i, item := range items {
		base := roundMoney(item.CalculateComponents().Base)
		taxable[i] = base
		subtotal += base
	}
	subtotal = roundMoney(subtotal)

	// Step 5. Reuse the existing guard so a discount cannot exceed the subtotal.
	discountAmount, err := applyHeaderDiscount(subtotal, discountType, discountValue)
	if err != nil {
		return invoiceTotals{}, err
	}
	discountAmount = roundMoney(discountAmount)

	// Step 6. A 100% discount leaves a zero taxable base, and therefore zero tax.
	taxableBase := roundMoney(subtotal - discountAmount)

	// Step 7. Scaling each base by taxableBase/subtotal is the same as removing
	// its pro-rata share of the header discount, but the shares then add up to
	// the taxable base by construction. Splitting the base also means a fully
	// discounted invoice taxes every line at zero.
	if subtotal > 0 {
		for i := range taxable {
			taxable[i] = roundMoney(taxable[i] * taxableBase / subtotal)
		}
		settleSplitRounding(taxable, taxableBase)
	} else {
		for i := range taxable {
			taxable[i] = 0
		}
	}

	// Steps 8-11. A line with an absent or zero rate is a zero-valued float in
	// Go, so a null tax_rate yields zero tax rather than a stray amount.
	lineTax := make([]float64, len(items))
	itemTax := 0.0
	for i, item := range items {
		lineTax[i] = roundMoney(taxable[i] * item.TaxRate / 100)
		itemTax += lineTax[i]
	}

	headerTax := roundMoney(taxableBase * headerTaxRate / 100)
	taxAmount := roundMoney(itemTax + headerTax)

	return invoiceTotals{
		Subtotal:       subtotal,
		DiscountAmount: discountAmount,
		TaxAmount:      taxAmount,
		Total:          roundMoney(taxableBase + taxAmount),
		LineTaxable:    taxable,
		LineTax:        lineTax,
	}, nil
}

// settleSplitRounding hands the cents lost to independent per-line rounding to
// the largest lines first, so the split taxable bases add up to the invoice
// taxable base exactly instead of drifting by up to half a cent per line.
// It is bounded and never drives a base below zero.
func settleSplitRounding(values []float64, target float64) {
	for attempt := 0; attempt <= 2*len(values); attempt++ {
		sum := 0.0
		largest := 0
		for i, value := range values {
			sum += value
			if value > values[largest] {
				largest = i
			}
		}
		residual := money.DollarsToCents(target - sum)
		if residual == 0 {
			return
		}
		if residual > 0 {
			values[largest] = roundMoney(values[largest] + 0.01)
		} else if values[largest] >= 0.01 {
			values[largest] = roundMoney(values[largest] - 0.01)
		} else {
			return
		}
	}
}

// validateMoneyInput is the last gate before an amount is persisted. The HTTP
// and tool entry points check names and dates but not money, and a negative or
// non-finite amount would otherwise be written as a negative total and due.
func validateMoneyInput(items []domain.InvoiceItemInput, headerTaxRate float64, discountType string, discountValue float64) error {
	if !isFiniteMoney(headerTaxRate) {
		return fmt.Errorf("tax_rate must be a finite number")
	}
	if headerTaxRate < 0 || headerTaxRate > 100 {
		return fmt.Errorf("tax_rate must be between 0 and 100")
	}
	if !isFiniteMoney(discountValue) {
		return fmt.Errorf("discount_value must be a finite number")
	}
	switch discountType {
	case "", "fixed", "percentage":
	default:
		return fmt.Errorf("discount_type must be 'fixed' or 'percentage'")
	}

	for i, item := range items {
		if !isFiniteMoney(item.Quantity) || !isFiniteMoney(item.UnitPrice) ||
			!isFiniteMoney(item.TaxRate) || !isFiniteMoney(item.DiscountPct) {
			return fmt.Errorf("item[%d]: quantity, unit_price, tax_rate and discount_pct must be finite", i)
		}
		if item.Quantity < 0 {
			return fmt.Errorf("item[%d]: quantity must not be negative", i)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("item[%d]: unit_price must not be negative", i)
		}
		if item.DiscountPct < 0 || item.DiscountPct > 100 {
			return fmt.Errorf("item[%d]: discount_pct must be between 0 and 100", i)
		}
		if item.TaxRate < 0 || item.TaxRate > 100 {
			return fmt.Errorf("item[%d]: tax_rate must be between 0 and 100", i)
		}
	}

	return nil
}

func isFiniteMoney(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
