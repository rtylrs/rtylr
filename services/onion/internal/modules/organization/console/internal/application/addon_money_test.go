package application

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestAddonDTOPricesUseExactMoneyAndDecimalJSON(t *testing.T) {
	addon := toAddonDTO(AddonCatalogRow{
		Price: NullString{String: "1234.56", Valid: true},
	})
	if addon.Price != money.Amount(123456) {
		t.Fatalf("addon price = %s, want 1234.56", addon.Price.String())
	}

	grant := toAddonGrantDTO(AddonGrantRow{
		AddonPrice: NullString{String: "987.65", Valid: true},
	})
	if grant.AddonPrice != money.Amount(98765) {
		t.Fatalf("grant addon price = %s, want 987.65", grant.AddonPrice.String())
	}

	payload, err := json.Marshal(struct {
		Addon AddonDTO      `json:"addon"`
		Grant AddonGrantDTO `json:"grant"`
	}{Addon: addon, Grant: grant})
	if err != nil {
		t.Fatalf("marshal addon payload: %v", err)
	}
	if string(payload) == "" || !strings.Contains(string(payload), `"price":1234.56`) || !strings.Contains(string(payload), `"addon_price":987.65`) {
		t.Fatalf("addon payload = %s, want decimal JSON prices", payload)
	}
}
