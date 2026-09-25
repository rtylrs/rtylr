package application

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type ReportTemplatePayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type ReportTemplateListOptions struct {
	Page    int32
	Limit   int32
	SortBy  string
	SortDir string
	Type    string
}

type ReportTemplatePage struct {
	Rows  []ReportTemplatePayload
	Total int64
}

func reportTemplateCatalog() []ReportTemplatePayload {
	return []ReportTemplatePayload{
		{ID: "revenue-monthly", Name: "Monthly Revenue", Description: "Month-by-month paid revenue over the last 12 months.", Type: "revenue"},
		{ID: "revenue-daily", Name: "Daily Revenue", Description: "Paid revenue per day for a custom window.", Type: "revenue"},
		{ID: "top-customers", Name: "Top Customers", Description: "Leaderboard of customers by lifetime spend.", Type: "customers"},
		{ID: "invoice-aging", Name: "Invoice Aging", Description: "A/R aging buckets with outstanding balances.", Type: "financial"},
		{ID: "expenses-category", Name: "Expenses by Category", Description: "Expense spend grouped by category.", Type: "financial"},
		{ID: "best-sellers", Name: "Best Sellers", Description: "Top-selling SKUs over the last 30 days.", Type: "products"},
		{ID: "slow-movers", Name: "Slow Movers", Description: "SKUs with no sales in a trailing window.", Type: "products"},
		{ID: "shift-coverage", Name: "Shift Coverage", Description: "Scheduled vs completed shifts by week.", Type: "operations"},
		{ID: "headcount", Name: "Headcount", Description: "Active staff count and period change.", Type: "operations"},
	}
}

func NormalizeReportTemplateListOptions(options ReportTemplateListOptions) (ReportTemplateListOptions, error) {
	options.Page = normalizePage(options.Page)
	options.Limit = normalizeListLimit(options.Limit)
	options.SortBy = strings.ToLower(strings.TrimSpace(options.SortBy))
	switch options.SortBy {
	case "":
		options.SortBy = "name"
	case "name", "type", "id":
	default:
		return ReportTemplateListOptions{}, fmt.Errorf("unsupported sort_by %q", options.SortBy)
	}
	options.SortDir = strings.ToLower(strings.TrimSpace(options.SortDir))
	switch options.SortDir {
	case "":
		options.SortDir = "asc"
	case "asc", "desc":
	default:
		return ReportTemplateListOptions{}, fmt.Errorf("unsupported sort_dir %q", options.SortDir)
	}
	options.Type = strings.ToLower(strings.TrimSpace(options.Type))
	if options.Type != "" {
		switch options.Type {
		case "revenue", "customers", "products", "operations", "financial":
		default:
			return ReportTemplateListOptions{}, fmt.Errorf("unsupported type %q", options.Type)
		}
	}
	return options, nil
}

func (s *Service) ListReportTemplates(_ context.Context, options ReportTemplateListOptions) (ReportTemplatePage, error) {
	options, err := NormalizeReportTemplateListOptions(options)
	if err != nil {
		return ReportTemplatePage{}, err
	}
	rows := reportTemplateCatalog()
	if options.Type != "" {
		filtered := rows[:0]
		for _, row := range rows {
			if row.Type == options.Type {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left, right := rows[i].Name, rows[j].Name
		switch options.SortBy {
		case "type":
			left, right = rows[i].Type, rows[j].Type
		case "id":
			left, right = rows[i].ID, rows[j].ID
		}
		comparison := strings.Compare(left, right)
		if comparison == 0 {
			comparison = strings.Compare(rows[i].ID, rows[j].ID)
		}
		if options.SortDir == "desc" {
			comparison = -comparison
		}
		return comparison < 0
	})

	total := int64(len(rows))
	start := int(pageOffset(options.Page, options.Limit))
	if start >= len(rows) {
		return ReportTemplatePage{Rows: []ReportTemplatePayload{}, Total: total}, nil
	}
	end := start + int(options.Limit)
	if end > len(rows) {
		end = len(rows)
	}
	return ReportTemplatePage{Rows: rows[start:end], Total: total}, nil
}
