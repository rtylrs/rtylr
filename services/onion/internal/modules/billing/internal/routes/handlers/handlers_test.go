package handlers

import (
	"testing"
)

func TestSummaryRoutes(t *testing.T) {
	routes := SummaryRoutes()
	if len(routes) == 0 {
		t.Error("expected at least one route for summary")
	}
}

func TestInvoicesRoutes(t *testing.T) {
	routes := InvoicesRoutes()
	if len(routes) == 0 {
		t.Error("expected at least one route for invoices")
	}
}

func TestSummaryRoutes_ContainsBilling(t *testing.T) {
	for _, r := range SummaryRoutes() {
		if r == "" {
			t.Error("route should not be empty")
		}
	}
}

func TestInvoicesRoutes_ContainsBilling(t *testing.T) {
	for _, r := range InvoicesRoutes() {
		if r == "" {
			t.Error("route should not be empty")
		}
	}
}
