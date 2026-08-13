package handlers

import "testing"

func TestAssessCaretAnswerAllowsNormalBusinessLanguage(t *testing.T) {
	got := assessCaretAnswer("Revenue was $12,000 this month. Focus next on the three overdue invoices.")
	if !got.Cacheable || got.Score < 90 || len(got.Findings) != 0 {
		t.Fatalf("quality = %+v", got)
	}
}

func TestAssessCaretAnswerBlocksProtocolLeakAndSyntheticRows(t *testing.T) {
	for _, response := range []string{
		`<tool>{"name":"crm.list_customers"}</tool>`,
		"Your top customer is John Smith, customer ID C001.",
		"Internal business snapshot evidence: row count 3",
	} {
		got := assessCaretAnswer(response)
		if got.Cacheable || got.Score >= 100 || len(got.Findings) == 0 {
			t.Fatalf("response %q quality = %+v", response, got)
		}
	}
}
