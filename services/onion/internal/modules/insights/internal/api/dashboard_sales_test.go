package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

func TestToSalesSummaryResponse(t *testing.T) {
	got := toSalesSummaryResponse(domain.SalesSummary{
		OpenPipelineValue: 4_500_000,
		OpenDeals:         12,
		Won:               domain.MoneyCard{Value: 300_000, DeltaPct: -25},
		RecentWonDeals:    4,
		MRR:               domain.MoneyCard{Value: 33_383, DeltaPct: 66.9},
		ActiveContracts:   4,
	})
	if got.OpenPipelineValue != 4_500_000 || got.OpenDealCount != 12 || got.RecentWonDealCount != 4 {
		t.Errorf("response = %+v", got)
	}
	if got.Won == nil || got.Won.Value != 300_000 || got.Won.DeltaPct != -25 {
		t.Errorf("Won = %+v, want 300000 down 25%%", got.Won)
	}
	if got.Mrr == nil || got.Mrr.Value != 33_383 || got.Mrr.DeltaPct != 66.9 {
		t.Errorf("Mrr = %+v, want 33383 up 66.9%%", got.Mrr)
	}
	if got.ActiveContractCount != 4 {
		t.Errorf("ActiveContractCount = %d, want 4", got.ActiveContractCount)
	}
}

// A business without deals still gets won and MRR cards, so a client reading
// won.value or mrr.value never dereferences a missing object.
func TestToSalesSummaryResponseWithoutDeals(t *testing.T) {
	got := toSalesSummaryResponse(domain.SalesSummary{})
	if got.Won == nil {
		t.Fatal("Won = nil, want an empty card")
	}
	if got.Mrr == nil {
		t.Fatal("Mrr = nil, want an empty card")
	}
	if got.OpenPipelineValue != 0 || got.OpenDealCount != 0 || got.RecentWonDealCount != 0 || got.Won.Value != 0 || got.Won.DeltaPct != 0 {
		t.Errorf("response = %+v, want zeros", got)
	}
	if got.Mrr.Value != 0 || got.Mrr.DeltaPct != 0 || got.ActiveContractCount != 0 {
		t.Errorf("Mrr = %+v with %d contracts, want zeros", got.Mrr, got.ActiveContractCount)
	}
}
