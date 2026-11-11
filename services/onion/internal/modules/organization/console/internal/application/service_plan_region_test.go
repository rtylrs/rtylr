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
		"long region":        {Region: "CAN", Currency: "CAD"},
		"digit region":       {Region: "C1", Currency: "CAD"},
		"numeric region":     {Region: "01", Currency: "CAD"},
		"unassigned region":  {Region: "ZZ", Currency: "CAD"},
		"private-use region": {Region: "XA", Currency: "CAD"},
		"empty region":       {Region: "", Currency: "CAD"},
		"short currency":     {Region: "CA", Currency: "CA"},
		"unknown currency":   {Region: "CA", Currency: "ABC"},
		"no-currency code":   {Region: "CA", Currency: "XXX"},
		"testing code":       {Region: "CA", Currency: "XTS"},
		"negative price":     {Region: "CA", Currency: "CAD", PricePerSeatCents: -1},
		"negative annual":    {Region: "CA", Currency: "CAD", AnnualPricePerSeatCents: -1},
	} {
		if _, err := normalizeRegionPrice(row); !errors.Is(err, ErrInvalidRegionPrice) {
			t.Errorf("%s: err = %v, want ErrInvalidRegionPrice", name, err)
		}
	}
}

func TestNormalizeRegion(t *testing.T) {
	for in, want := range map[string]string{"ca": "CA", " US ": "US", "GB": "GB", "ZZ": "", "CAN": "", "": "", "1A": ""} {
		if got := normalizeRegion(in); got != want {
			t.Errorf("normalizeRegion(%q) = %q, want %q", in, got, want)
		}
	}
}
