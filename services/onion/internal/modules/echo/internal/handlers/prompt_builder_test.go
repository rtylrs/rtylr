package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestBuildCaretPromptIncludesBusinessContractModeAndContext(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAnalyst,
		ConversationPrompt: "Current request: revenue today",
		SystemContext:      "Route: /finance/dashboard\nVisible module: finance",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	mustContain := []string{
		caretSystemPreamble,
		"Prompt contract: caret-business-v1",
		"Today: 2026-06-18",
		"Act like a capable business operator",
		"Mode:\nanalyst",
		"server-provided analytics grounding",
		"Current application context:",
		"Route: /finance/dashboard",
		"Conversation, memory, and current request:",
		"Current request: revenue today",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q\nprompt:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("analyst prompt should not invite model-emitted tool calls:\n%s", got)
	}
}

func TestBuildCaretPromptAgentModeIncludesExecutableToolProposalsOnly(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAgent,
		ConversationPrompt: "Current request: create a customer",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	mustContain := []string{
		"Available business capability catalog:",
		"business.snapshot [read]",
		"business.get_kpis [read]",
		"insights.get_dashboard [read]",
		"insights.get_revenue_series [read]",
		"insights.get_invoice_aging [read]",
		"insights.get_operations_kpis [read]",
		"crm.list_customers [read]",
		"crm.list_campaigns [read]",
		"crm.get_campaign [read]",
		"crm.list_segments [read]",
		"loyalty.get_config [read]",
		"loyalty.list_members [read]",
		"loyalty.list_rewards [read]",
		"loyalty.list_transactions [read]",
		"pos.list_orders [read]",
		"pos.list_locations [read]",
		"pos.list_terminals [read]",
		"pos.list_users [read]",
		"pos.list_floors [read]",
		"pos.list_tables [read]",
		"pos.list_menu_items [read]",
		"pos.list_modifiers [read]",
		"pos.list_sessions [read]",
		"pos.list_cash_movements [read]",
		"pos.list_stock_movements [read]",
		"pos.list_branch_inventory [read]",
		"pos.list_low_stock_items [read]",
		"store.list_orders [read]",
		"store.get_order [read]",
		"store.list_products [read]",
		"store.list_categories [read]",
		"support.list_tickets [read]",
		"support.get_ticket [read]",
		"support.list_inbox [read]",
		"support.get_unread_count [read]",
		"support.list_customers [read]",
		"support.get_customer [read]",
		"support.list_canned_responses [read]",
		"support.list_articles [read]",
		"erp.list_products [read]",
		"erp.list_categories [read]",
		"erp.list_brands [read]",
		"erp.list_suppliers [read]",
		"erp.list_warehouses [read]",
		"erp.list_purchase_orders [read]",
		"erp.list_stock_transfers [read]",
		"hr.list_employees [read]",
		"hr.list_locations [read]",
		"hr.list_departments [read]",
		"hr.list_work_schedules [read]",
		"hr.list_roles [read]",
		"hr.get_setup_status [read]",
		"hr.get_config [read]",
		"hr.list_reviews [read]",
		"hr.list_leave_types [read]",
		"hr.list_leave_requests [read]",
		"hr.get_leave_balances [read]",
		"hr.get_leave_calendar [read]",
		"hr.list_overtime_requests [read]",
		"hr.get_overtime_config [read]",
		"hr.get_payroll_config [read]",
		"hr.list_payroll_runs [read]",
		"hr.get_payroll_run [read]",
		"hr.list_time_entries [read]",
		"hr.list_timesheets [read]",
		"hr.get_timesheet [read]",
		"hr.list_attendance [read]",
		"hr.get_location_attendance [read]",
		"flow.list_tasks [read]",
		"flow.get_insights [read]",
		"flow.get_team_workload [read]",
		"flow.get_roadmap [read]",
		"finance.list_invoices [read]",
		"finance.get_balance_sheet [read]",
		"finance.get_cash_flow_report [read]",
		"finance.list_expenses [read]",
		"finance.list_expense_categories [read]",
		"finance.list_bank_accounts [read]",
		"finance.get_bank_summary [read]",
		"finance.list_bank_feed_transactions [read]",
		"finance.get_reconciliation_overview [read]",
		"finance.list_journal_entries [read]",
		"finance.get_tax_config [read]",
		"finance.get_pl_report [read]",
		"finance.list_bills [read]",
		"booking.list_appointments [read]",
		"booking.get_dashboard [read]",
		"booking.get_calendar [read]",
		"booking.list_services [read]",
		"booking.list_resources [read]",
		"booking.list_waitlist [read]",
		"booking.list_reminders [read]",
		"pos.get_day_summary [read]",
		"crm.create_customer [write, approval required]",
		"Structured stream controls:",
		`<tool>{"name":"tool.name","input":{}}</tool>`,
		"Read tools run immediately",
		"never reveal private chain-of-thought",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("agent prompt missing %q\nprompt:\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"analytics.query [read]"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("agent prompt should not advertise non-executable read tool %q\nprompt:\n%s", forbidden, got)
		}
	}
}

func TestBuildCaretPromptSupportModeOmitsToolCatalog(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeSupport,
		ConversationPrompt: "Current request: where do I find invoices?",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	if strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("support prompt should not expose tool catalog:\n%s", got)
	}
	if !strings.Contains(got, "Mode:\nsupport") {
		t.Fatalf("support prompt missing mode:\n%s", got)
	}
}
