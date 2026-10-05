package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

func TestSalesMoneyFieldsPreserveExactDecimalJSON(t *testing.T) {
	mrrPotential := money.Amount(123456)
	payload, err := json.Marshal(struct {
		Lead     leadPayload            `json:"lead"`
		Contract contractPayload        `json:"contract"`
		Pipeline pipelineSummaryPayload `json:"pipeline"`
	}{
		Lead:     payloadLead(domain.Lead{MRRPotential: &mrrPotential}),
		Contract: payloadContract(domain.Contract{MRR: money.Amount(98765)}),
		Pipeline: payloadPipelineSummary(domain.PipelineSummary{TotalMRR: money.Amount(11122)}),
	})
	if err != nil {
		t.Fatalf("marshal sales payload: %v", err)
	}

	encoded := string(payload)
	for _, want := range []string{`"mrr_potential":1234.56`, `"mrr":987.65`, `"total_mrr":111.22`} {
		if !strings.Contains(encoded, want) {
			t.Fatalf("sales payload = %s, missing %s", encoded, want)
		}
	}
}
