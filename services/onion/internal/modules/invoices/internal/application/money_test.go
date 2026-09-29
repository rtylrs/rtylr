package application

import (
	"context"
	"math"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

// These cases assert whole cents, not approximate amounts, because the point is
// that the stored total, the emailed total and the tax report total are the
// same number. Every expectation below is hand-checked in the comment.

func TestComputeInvoiceTotals_ExactCents(t *testing.T) {
	tests := []struct {
		name          string
		items         []domain.InvoiceItemInput
		headerTaxRate float64
		discountType  string
		discountValue float64
		want          invoiceTotals
	}{
		{
			// 3 x 12.50 = 37.50, no tax anywhere.
			name:  "single line, no tax",
			items: []domain.InvoiceItemInput{{Quantity: 3, UnitPrice: 12.50}},
			want:  invoiceTotals{Subtotal: 37.50, Total: 37.50},
		},
		{
			// 2 x 19.99 = 39.98; 20% item tax = 7.996 -> 8.00 (half away from
			// zero is not needed, 7.996 rounds down); total 47.98.
			name:  "single line with item tax",
			items: []domain.InvoiceItemInput{{Quantity: 2, UnitPrice: 19.99, TaxRate: 20}},
			want:  invoiceTotals{Subtotal: 39.98, TaxAmount: 8.00, Total: 47.98},
		},
		{
			// 4 x 25.00 = 100.00; 20% header tax = 20.00; total 120.00.
			name:          "single line with header tax",
			items:         []domain.InvoiceItemInput{{Quantity: 4, UnitPrice: 25}},
			headerTaxRate: 20,
			want:          invoiceTotals{Subtotal: 100.00, TaxAmount: 20.00, Total: 120.00},
		},
		{
			// 10% header discount on 100.00 leaves 90.00 payable. The line tax
			// and the header tax are each 10% of 90.00 -> 9.00 each. Taxing the
			// item on the undiscounted base would declare 19.00 of tax on 90.00
			// of revenue.
			name:          "multiple lines, percentage discount then tax",
			items:         []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 100, TaxRate: 10}},
			headerTaxRate: 10,
			discountType:  "percentage",
			discountValue: 10,
			want:          invoiceTotals{Subtotal: 100.00, DiscountAmount: 10.00, TaxAmount: 18.00, Total: 108.00},
		},
		{
			// $20.00 off $150.00 leaves $130.00 payable, split pro-rata by base
			// as $86.67 and $43.33. Only the second line is taxed: 10% of
			// 43.33 = 4.333 -> 4.33. Total 130.00 + 4.33 = 134.33.
			name: "multiple lines, fixed discount then tax, mixed rates",
			items: []domain.InvoiceItemInput{
				{Quantity: 1, UnitPrice: 100},
				{Quantity: 1, UnitPrice: 50, TaxRate: 10},
			},
			discountValue: 20,
			want: invoiceTotals{
				Subtotal: 150.00, DiscountAmount: 20.00, TaxAmount: 4.33, Total: 134.33,
				LineTaxable: []float64{86.67, 43.33}, LineTax: []float64{0, 4.33},
			},
		},
		{
			// A 100% header discount leaves nothing payable, so nothing is
			// taxable and the tax must be 0.00, not the full 100.00 of item tax
			// that the pre-fix arithmetic produced.
			name:          "100% header discount zeroes the tax",
			items:         []domain.InvoiceItemInput{{Quantity: 10, UnitPrice: 50, TaxRate: 20}},
			headerTaxRate: 20,
			discountType:  "percentage",
			discountValue: 100,
			want:          invoiceTotals{Subtotal: 500.00, DiscountAmount: 500.00, Total: 0.00},
		},
		{
			// A 100% line discount already zeroes that line's base, so its tax
			// is 0.00 even with both rates set.
			name: "100% line discount zeroes that line's tax",
			items: []domain.InvoiceItemInput{
				{Quantity: 3, UnitPrice: 100, TaxRate: 20, DiscountPct: 100},
			},
			headerTaxRate: 20,
			want:          invoiceTotals{Subtotal: 0, Total: 0},
		},
		{
			// Zero rates everywhere: the taxable base is the subtotal.
			name:          "zero tax rates",
			items:         []domain.InvoiceItemInput{{Quantity: 2, UnitPrice: 50, TaxRate: 0}},
			headerTaxRate: 0,
			want:          invoiceTotals{Subtotal: 100.00, TaxAmount: 0, Total: 100.00},
		},
		{
			// A tax_rate of null/absent decodes to the zero value, so it must
			// produce 0.00 of tax rather than some inherited or default rate.
			name: "null/absent tax rate adds no tax",
			items: []domain.InvoiceItemInput{
				{Quantity: 1, UnitPrice: 100},
				{Quantity: 1, UnitPrice: 40, TaxRate: 0},
			},
			headerTaxRate: 0,
			want:          invoiceTotals{Subtotal: 140.00, TaxAmount: 0, Total: 140.00},
		},
		{
			// Zero quantity and a zero-priced line both contribute 0.00 and
			// must not drag the subtotal negative.
			name: "zero quantity and zero price",
			items: []domain.InvoiceItemInput{
				{Quantity: 0, UnitPrice: 99.99, TaxRate: 10},
				{Quantity: 5, UnitPrice: 0, TaxRate: 10},
				{Quantity: 1, UnitPrice: 20, TaxRate: 10},
			},
			headerTaxRate: 10,
			want:          invoiceTotals{Subtotal: 20.00, TaxAmount: 4.00, Total: 24.00},
		},
		{
			// 3 x 33.335 = 100.005 -> 100.01 (rounded at the line, not left to
			// bleed into the total). 7% of 100.01 = 7.0007 -> 7.00. Header 5%
			// of 100.01 = 5.0005 -> 5.00. Total 100.01 + 12.00 = 112.01. Before
			// the fix the stored total was 112.0056.
			name:          "fractional price rounds per line, not at display time",
			items:         []domain.InvoiceItemInput{{Quantity: 3, UnitPrice: 33.335, TaxRate: 7}},
			headerTaxRate: 5,
			want:          invoiceTotals{Subtotal: 100.01, TaxAmount: 12.00, Total: 112.01},
		},
		{
			// 40 x 25.00 at 19% and 12 x 99.99 at 7%. Subtotal 2199.88. A 10%
			// header discount leaves 1979.892 -> 1979.89 payable, so the lines
			// split to 900.00 and 1079.89. Line tax 171.00 + 75.59 = 246.59.
			// Header tax 20% of 1979.89 = 395.98. Total 1979.89 + 642.57.
			name:          "high quantity, high rate, discount and header tax",
			items:         []domain.InvoiceItemInput{{Quantity: 40, UnitPrice: 25, TaxRate: 19}, {Quantity: 12, UnitPrice: 99.99, TaxRate: 7}},
			headerTaxRate: 20,
			discountType:  "percentage",
			discountValue: 10,
			want: invoiceTotals{
				Subtotal: 2199.88, DiscountAmount: 219.99, TaxAmount: 642.57, Total: 2622.46,
				LineTaxable: []float64{900.00, 1079.89}, LineTax: []float64{171.00, 75.59},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeInvoiceTotals(tt.items, tt.headerTaxRate, tt.discountType, tt.discountValue)
			if err != nil {
				t.Fatalf("computeInvoiceTotals() error = %v", err)
			}

			if got.Subtotal != tt.want.Subtotal {
				t.Errorf("Subtotal = %v, want %v", got.Subtotal, tt.want.Subtotal)
			}
			if got.DiscountAmount != tt.want.DiscountAmount {
				t.Errorf("DiscountAmount = %v, want %v", got.DiscountAmount, tt.want.DiscountAmount)
			}
			if got.TaxAmount != tt.want.TaxAmount {
				t.Errorf("TaxAmount = %v, want %v", got.TaxAmount, tt.want.TaxAmount)
			}
			if got.Total != tt.want.Total {
				t.Errorf("Total = %v, want %v", got.Total, tt.want.Total)
			}

			// The documented invariant, asserted exactly: total is the taxable
			// base plus the tax, never re-derived.
			taxableBase := got.Subtotal - got.DiscountAmount
			if got.Total != taxableBase+got.TaxAmount {
				t.Errorf("Total = %v, want subtotal-discount+tax = %v", got.Total, taxableBase+got.TaxAmount)
			}

			lineBase, lineTax := 0.0, 0.0
			for i := range tt.items {
				// Per-line expectations are optional; the aggregate invariants
				// below are asserted either way.
				if len(tt.want.LineTaxable) > 0 && got.LineTaxable[i] != tt.want.LineTaxable[i] {
					t.Errorf("LineTaxable[%d] = %v, want %v", i, got.LineTaxable[i], tt.want.LineTaxable[i])
				}
				if len(tt.want.LineTax) > 0 && got.LineTax[i] != tt.want.LineTax[i] {
					t.Errorf("LineTax[%d] = %v, want %v", i, got.LineTax[i], tt.want.LineTax[i])
				}
				if got.LineTaxable[i] < 0 {
					t.Errorf("LineTaxable[%d] = %v, want a non-negative base", i, got.LineTaxable[i])
				}
				lineBase += got.LineTaxable[i]
				lineTax += got.LineTax[i]
			}
			if lineBase != taxableBase {
				t.Errorf("line bases sum to %v, want the taxable base %v", lineBase, taxableBase)
			}
			if got.TaxAmount < lineTax {
				t.Errorf("TaxAmount = %v is less than the line tax %v", got.TaxAmount, lineTax)
			}

			// The rows foot: the line totals add up to the header less the
			// header-only tax layer, to the cent. Summed in integer cents so
			// the assertion is not a float comparison.
			rowCents, lineTaxCents := int64(0), int64(0)
			for i := range tt.items {
				rowCents += money.DollarsToCents(got.LineTotal(i))
				lineTaxCents += money.DollarsToCents(got.LineTax[i])
			}
			wantRowCents := money.DollarsToCents(taxableBase) + lineTaxCents
			if rowCents != wantRowCents {
				t.Errorf("line rows sum to %d cents, want %d cents", rowCents, wantRowCents)
			}
		})
	}
}

func TestComputeInvoiceTotals_RejectsUnusableAmounts(t *testing.T) {
	okItem := domain.InvoiceItemInput{Quantity: 1, UnitPrice: 100, TaxRate: 10}

	tests := []struct {
		name          string
		items         []domain.InvoiceItemInput
		headerTaxRate float64
		discountType  string
		discountValue float64
	}{
		{name: "negative quantity", items: []domain.InvoiceItemInput{{Quantity: -2, UnitPrice: 50}}},
		{name: "negative unit price", items: []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: -50}}},
		{name: "line discount over 100", items: []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 50, DiscountPct: 150}}},
		{name: "line discount under 0", items: []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 50, DiscountPct: -10}}},
		{name: "line tax over 100", items: []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 50, TaxRate: 120}}},
		{name: "negative line tax", items: []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 50, TaxRate: -5}}},
		{name: "negative header tax", items: []domain.InvoiceItemInput{okItem}, headerTaxRate: -5},
		{name: "header tax over 100", items: []domain.InvoiceItemInput{okItem}, headerTaxRate: 101},
		{name: "negative discount value", items: []domain.InvoiceItemInput{okItem}, discountValue: -5},
		{name: "unknown discount type", items: []domain.InvoiceItemInput{okItem}, discountType: "giftwrap"},
		{
			name: "non-finite unit price",
			items: []domain.InvoiceItemInput{{
				Quantity: 1, UnitPrice: math.Inf(1), TaxRate: 10,
			}},
		},
		{
			name: "NaN quantity",
			items: []domain.InvoiceItemInput{{
				Quantity: math.NaN(), UnitPrice: 10,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := computeInvoiceTotals(tt.items, tt.headerTaxRate, tt.discountType, tt.discountValue); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

// A header discount larger than the subtotal is still rejected by the existing
// guard, and the service turns it into ErrDiscountExceedsTotal.
func TestCreateInvoice_RejectsDiscountAboveSubtotal(t *testing.T) {
	s := &Service{repo: &mockRepo{}}

	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate:     "2024-06-15",
		DueDate:       "2024-06-20",
		DiscountType:  "fixed",
		DiscountValue: 500,
		Items:         []domain.InvoiceItemInput{{Quantity: 1, UnitPrice: 100}},
	})
	if err != domain.ErrDiscountExceedsTotal {
		t.Fatalf("error = %v, want %v", err, domain.ErrDiscountExceedsTotal)
	}
}

// A negative line amount must be refused rather than persisted as a negative
// total with a negative amount due.
func TestCreateInvoice_RejectsNegativeLineAmount(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}

	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-20",
		Items:     []domain.InvoiceItemInput{{Quantity: -2, UnitPrice: 50}},
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if repo.created != nil {
		t.Fatal("a rejected invoice must not reach the repository")
	}
}

// The create and update paths must produce byte-identical money for the same
// input; they used to duplicate the arithmetic and drift.
func TestUpdateInvoice_MatchesCreateTotalsAndFootsTheRows(t *testing.T) {
	repo := &mockRepo{invoice: &domain.Invoice{ID: 10, Status: domain.StatusDraft}}
	s := &Service{repo: repo}

	items := []domain.InvoiceItemInput{
		{Quantity: 40, UnitPrice: 25, TaxRate: 19},
		{Quantity: 12, UnitPrice: 99.99, TaxRate: 7},
	}

	created, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate:     "2024-06-15",
		DueDate:       "2024-06-20",
		DiscountType:  "percentage",
		DiscountValue: 10,
		TaxRate:       20,
		Items:         items,
	})
	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}

	updated, err := s.UpdateInvoice(context.Background(), 1, 10, domain.InvoiceUpdateInput{
		IssueDate:     "2024-06-15",
		DueDate:       "2024-06-20",
		DiscountType:  "percentage",
		DiscountValue: 10,
		TaxRate:       20,
		Items:         items,
	})
	if err != nil {
		t.Fatalf("UpdateInvoice() error = %v", err)
	}

	if updated.Subtotal != created.Subtotal ||
		updated.DiscountAmount != created.DiscountAmount ||
		updated.TaxAmount != created.TaxAmount ||
		updated.Total != created.Total {
		t.Errorf("update totals (%v/%v/%v/%v) differ from create (%v/%v/%v/%v)",
			updated.Subtotal, updated.DiscountAmount, updated.TaxAmount, updated.Total,
			created.Subtotal, created.DiscountAmount, created.TaxAmount, created.Total)
	}

	assertInvoiceFoots(t, updated, 20)
}

// The stored total must be whole cents so that paying AmountDue settles the
// invoice to exactly zero through the cents-based payment path. Before the fix
// this basket stored a total of 112.0056.
func TestCreateInvoice_TotalIsWholeCents(t *testing.T) {
	s := &Service{repo: &mockRepo{}}

	inv, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-20",
		TaxRate:   5,
		Items:     []domain.InvoiceItemInput{{Quantity: 3, UnitPrice: 33.335, TaxRate: 7}},
	})
	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}

	for _, amount := range []struct {
		name  string
		value float64
	}{
		{"Subtotal", inv.Subtotal},
		{"TaxAmount", inv.TaxAmount},
		{"Total", inv.Total},
		{"AmountDue", inv.AmountDue},
	} {
		if got := money.CentsToDollars(money.DollarsToCents(amount.value)); got != amount.value {
			t.Errorf("%s = %v, want a whole-cent value (%v)", amount.name, amount.value, got)
		}
	}

	assertInvoiceFoots(t, inv, 5)
}

// settleSplitRounding is the one piece of this that has to converge rather
// than just compute, so sweep a deterministic spread of awkward baskets and
// assert the invariants hold on every one of them.
func TestComputeInvoiceTotals_InvariantsOverAwkwardBaskets(t *testing.T) {
	// A fixed linear congruential sequence: same baskets on every run.
	seed := int64(20240615)
	next := func(limit int64) int64 {
		seed = (seed*6364136223846793005 + 1442695040888963407) % (1 << 62)
		value := seed % limit
		if value < 0 {
			value += limit
		}
		return value
	}

	for basket := 0; basket < 400; basket++ {
		lineCount := int(next(6)) + 1
		items := make([]domain.InvoiceItemInput, 0, lineCount)
		for i := 0; i < lineCount; i++ {
			quantity := float64(next(4))
			if i == 0 {
				// Keep every basket payable; later lines still exercise zero
				// quantities and zero prices.
				quantity = 1
			}
			items = append(items, domain.InvoiceItemInput{
				// Cents that do not divide evenly: .005, .335, .01, .125.
				Quantity:    quantity,
				UnitPrice:   float64(next(40000))/1000 + 0.005,
				TaxRate:     float64(next(3)) * 10,
				DiscountPct: float64(next(3)) * 20,
			})
		}
		discountType := "percentage"
		discountValue := float64(next(101)) // a percentage, 0..100
		if next(2) == 0 {
			// A fixed discount must stay within the subtotal or the guard
			// legitimately rejects it, so scale it to the lines.
			discountType = "fixed"
			subtotal := 0.0
			for _, item := range items {
				subtotal += roundMoney(item.CalculateComponents().Base)
			}
			discountValue = float64(next(50000)) / 100
			if limit := roundMoney(subtotal); discountValue > limit {
				discountValue = limit
			}
		}
		headerTaxRate := float64(next(3)) * 10

		got, err := computeInvoiceTotals(items, headerTaxRate, discountType, discountValue)
		if err != nil {
			t.Fatalf("basket %d: unexpected error: %v", basket, err)
		}

		taxableBaseCents := money.DollarsToCents(got.Subtotal) - money.DollarsToCents(got.DiscountAmount)
		lineBaseCents, lineTaxCents, rowCents := int64(0), int64(0), int64(0)
		for i := range items {
			base := money.DollarsToCents(got.LineTaxable[i])
			if base < 0 {
				t.Fatalf("basket %d line %d: negative taxable base %d cents", basket, i, base)
			}
			lineBaseCents += base
			lineTaxCents += money.DollarsToCents(got.LineTax[i])
			rowCents += money.DollarsToCents(got.LineTotal(i))
		}
		if lineBaseCents != taxableBaseCents {
			t.Fatalf("basket %d: line bases sum to %d cents, want the taxable base %d", basket, lineBaseCents, taxableBaseCents)
		}
		if rowCents != taxableBaseCents+lineTaxCents {
			t.Fatalf("basket %d: rows sum to %d cents, want %d", basket, rowCents, taxableBaseCents+lineTaxCents)
		}
		if got.Total < 0 {
			t.Fatalf("basket %d: negative total %v", basket, got.Total)
		}
		headerTaxCents := money.DollarsToCents(money.CentsToDollars(taxableBaseCents) * headerTaxRate / 100)
		if rowCents+headerTaxCents != money.DollarsToCents(got.Total) {
			t.Fatalf("basket %d: rows + header tax do not reach the total", basket)
		}
		if money.DollarsToCents(got.Total) != taxableBaseCents+money.DollarsToCents(got.TaxAmount) {
			t.Fatalf("basket %d: total does not foot to subtotal-discount+tax", basket)
		}
	}
}

// assertInvoiceFoots checks the money invariants in whole cents: the header
// foots to subtotal - discount + tax, the rows foot to the header less exactly
// the header-rate tax layer, and the amount due is what is left after payments.
func assertInvoiceFoots(t *testing.T, inv *domain.Invoice, headerTaxRate float64) {
	t.Helper()

	taxableBase := inv.Subtotal - inv.DiscountAmount
	if got, want := money.DollarsToCents(inv.Total), money.DollarsToCents(taxableBase)+money.DollarsToCents(inv.TaxAmount); got != want {
		t.Errorf("Total = %d cents, want subtotal-discount+tax = %d cents", got, want)
	}
	if got, want := money.DollarsToCents(inv.AmountDue), money.DollarsToCents(inv.Total)-money.DollarsToCents(inv.AmountPaid); got != want {
		t.Errorf("AmountDue = %d cents, want total-paid = %d cents", got, want)
	}

	rowCents := int64(0)
	for _, item := range inv.Items {
		rowCents += money.DollarsToCents(item.Total)
	}
	headerTaxCents := money.DollarsToCents(taxableBase * headerTaxRate / 100)
	if got, want := rowCents+headerTaxCents, money.DollarsToCents(inv.Total); got != want {
		t.Errorf("line rows (%d cents) plus the header-only tax (%d cents) = %d cents, want the total %d cents",
			rowCents, headerTaxCents, got, want)
	}
}
