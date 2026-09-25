package infrastructure

import (
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
)

// TestInvoiceAgingBucketsMatchContract pins B74-R08: the API must expose every
// bucket declared by the insights.financial.v1 InvoiceAgingBucket contract
// (current, 1-30, 31-60, 61-90, 90+) in order, including not-yet-due
// receivables in "current", so total outstanding covers current receivables.
func TestInvoiceAgingBucketsMatchContract(t *testing.T) {
	rows := []sqlc.GetInvoiceAgingRow{
		{Bucket: "current", InvoiceCount: 2, AmountDue: []byte("50000.00")},
		{Bucket: "31-60", InvoiceCount: 1, AmountDue: []byte("125.50")},
	}

	buckets := invoiceAgingBucketsFromRows(rows)
	wantOrder := []string{"current", "1-30", "31-60", "61-90", "90+"}
	if len(buckets) != len(wantOrder) {
		t.Fatalf("len(buckets) = %d, want %d", len(buckets), len(wantOrder))
	}
	for i, want := range wantOrder {
		if buckets[i].Bucket != want {
			t.Errorf("buckets[%d].Bucket = %q, want %q", i, buckets[i].Bucket, want)
		}
	}

	if buckets[0].InvoiceCount != 2 || buckets[0].AmountMinorUnits != 5000000 {
		t.Errorf("current bucket = %+v, want count 2 and 5000000 minor units", buckets[0])
	}
	if buckets[2].InvoiceCount != 1 || buckets[2].AmountMinorUnits != 12550 {
		t.Errorf("31-60 bucket = %+v, want count 1 and 12550 minor units", buckets[2])
	}
	if buckets[1].InvoiceCount != 0 || buckets[1].AmountMinorUnits != 0 {
		t.Errorf("1-30 bucket = %+v, want zero-filled bucket", buckets[1])
	}
}
