package application

import (
	"context"
	"testing"
)

func TestListReportTemplatesFiltersSortsAndPaginates(t *testing.T) {
	service := &Service{}
	page, err := service.ListReportTemplates(context.Background(), ReportTemplateListOptions{
		Page: 1, Limit: 2, Type: "financial", SortBy: "name", SortDir: "asc",
	})
	if err != nil {
		t.Fatalf("ListReportTemplates: %v", err)
	}
	if page.Total != 2 || len(page.Rows) != 2 {
		t.Fatalf("page = %+v, want two financial templates", page)
	}
	if page.Rows[0].ID != "expenses-category" || page.Rows[1].ID != "invoice-aging" {
		t.Fatalf("template order = %q, %q", page.Rows[0].ID, page.Rows[1].ID)
	}

	descending, err := service.ListReportTemplates(context.Background(), ReportTemplateListOptions{
		Page: 1, Limit: 1, Type: "financial", SortBy: "name", SortDir: "desc",
	})
	if err != nil {
		t.Fatalf("ListReportTemplates descending: %v", err)
	}
	if len(descending.Rows) != 1 || descending.Rows[0].ID != "invoice-aging" {
		t.Fatalf("descending first row = %+v, want invoice-aging", descending.Rows)
	}
}

func TestReportListOptionsRejectUnknownSortAndType(t *testing.T) {
	if _, err := NormalizeReportTemplateListOptions(ReportTemplateListOptions{SortBy: "sql"}); err == nil {
		t.Fatal("expected unsupported template sort to fail")
	}
	if _, err := NormalizeScheduledReportListOptions(ScheduledReportListOptions{SortDir: "sideways"}); err == nil {
		t.Fatal("expected unsupported report sort direction to fail")
	}
}
