package application

import (
	"errors"
	"testing"
)

func TestNormalizeRegionPrice(t *testing.T) {
	got, err := normalizeRegionPrice(PlanRegionPriceRow{Region: " ca ", Currency: "cad", PaddlePriceID: " pri_1 ", PricePerSeatCents: 899})
	if err != nil {
		t.Fatal(err)
	}
	if got.Region != "CA" || got.Currency != "CAD" || got.PaddlePriceID != "pri_1" {
		t.Errorf("not normalized: %+v", got)
	}
	for name, row := range map[string]PlanRegionPriceRow{
		"long region":     {Region: "CAN", Currency: "CAD"},
		"digit region":    {Region: "C1", Currency: "CAD"},
		"short currency":  {Region: "CA", Currency: "CA"},
		"negative price":  {Region: "CA", Currency: "CAD", PricePerSeatCents: -1},
		"negative annual": {Region: "CA", Currency: "CAD", AnnualPricePerSeatCents: -1},
	} {
		if _, err := normalizeRegionPrice(row); !errors.Is(err, ErrInvalidRegionPrice) {
			t.Errorf("%s: err = %v, want ErrInvalidRegionPrice", name, err)
		}
	}
}
