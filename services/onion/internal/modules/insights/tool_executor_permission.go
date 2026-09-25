package insights

import (
	"errors"
	"net/mail"
	"strings"

	insightsapp "github.com/voxire/onion/internal/modules/insights/internal/application"
)

func insightsToolPermission(name string) string {
	switch name {
	case scheduleInsightsReportToolName:
		return "insights:reports:create"
	case getInsightsCustomerKPIsToolName,
		listInsightsTopCustomersToolName,
		getInsightsCustomerChurnToolName,
		getInsightsLTVDistributionToolName:
		return "insights:customers:view"
	case getInsightsCashFlowToolName,
		getInsightsExpensesByCategoryToolName,
		getInsightsInvoiceAgingToolName:
		return "insights:financial:view"
	default:
		return "insights:dashboard:view"
	}
}

func parseScheduleReportToolInput(input map[string]any) (insightsapp.ScheduledReportInput, error) {
	if input == nil {
		input = map[string]any{}
	}

	templateID := normalizeScheduleReportTemplate(
		insightsToolFirstString(input, "template_id", "templateId", "report_id", "report", "template", "name"),
		input,
	)
	cadence := normalizeScheduleReportCadence(
		insightsToolFirstString(input, "cadence", "frequency", "interval", "period", "repeat"),
	)
	channel := normalizeScheduleReportChannel(
		insightsToolFirstString(input, "channel", "delivery_channel", "deliveryChannel", "via"),
	)
	recipient := strings.TrimSpace(insightsToolFirstString(input, "recipient", "to", "email", "slack_channel", "slackChannel", "channel_id", "channelId"))
	if recipient == "" {
		recipient = firstScheduleReportRecipient(input["recipients"])
	}
	if channel == "" {
		channel = inferScheduleReportChannel(recipient)
	}
	if cadence == "" {
		cadence = inferScheduleReportCadence(templateID)
	}

	if err := validateScheduleReportRecipient(channel, recipient); err != nil {
		return insightsapp.ScheduledReportInput{}, err
	}

	return insightsapp.ScheduledReportInput{
		TemplateID: templateID,
		Cadence:    cadence,
		Channel:    channel,
		Recipient:  recipient,
		IsActive:   scheduleReportOptionalBool(input["is_active"], input["isActive"]),
	}, nil
}

func normalizeScheduleReportTemplate(value string, input map[string]any) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	slug := scheduleReportSlug(value)
	switch slug {
	case "revenue", "sales", "revenue-report", "sales-report":
		cadence := normalizeScheduleReportCadence(
			insightsToolFirstString(input, "cadence", "frequency", "interval", "period", "repeat"),
		)
		if cadence == "daily" {
			return "revenue-daily"
		}
		return "revenue-monthly"
	case "daily-revenue", "daily-sales", "revenue-daily", "sales-daily":
		return "revenue-daily"
	case "monthly-revenue", "monthly-sales", "revenue-monthly", "sales-monthly":
		return "revenue-monthly"
	case "top-customer", "top-customers", "customers", "customer-leaderboard":
		return "top-customers"
	case "invoice-aging", "ar-aging", "accounts-receivable", "receivables":
		return "invoice-aging"
	case "expenses", "expense", "expenses-category", "expenses-by-category", "expense-category":
		return "expenses-category"
	case "best-seller", "best-sellers", "bestsellers", "best-selling-products":
		return "best-sellers"
	case "slow-mover", "slow-movers", "slow-moving-products":
		return "slow-movers"
	case "shift-coverage", "shifts", "shift-report":
		return "shift-coverage"
	case "headcount", "staff-count", "employee-count":
		return "headcount"
	default:
		return slug
	}
}

func normalizeScheduleReportCadence(value string) string {
	slug := scheduleReportSlug(value)
	switch slug {
	case "day", "daily", "every-day", "each-day":
		return "daily"
	case "week", "weekly", "every-week", "each-week":
		return "weekly"
	case "month", "monthly", "every-month", "each-month":
		return "monthly"
	default:
		return slug
	}
}

func normalizeScheduleReportChannel(value string) string {
	slug := scheduleReportSlug(value)
	switch slug {
	case "mail", "e-mail", "email":
		return "email"
	case "slack", "slack-channel":
		return "slack"
	default:
		return slug
	}
}

func inferScheduleReportCadence(templateID string) string {
	switch templateID {
	case "revenue-daily":
		return "daily"
	case "revenue-monthly":
		return "monthly"
	default:
		return ""
	}
}

func inferScheduleReportChannel(recipient string) string {
	recipient = strings.TrimSpace(recipient)
	if recipient == "" {
		return ""
	}
	if strings.HasPrefix(recipient, "#") {
		return "slack"
	}
	return "email"
}

func firstScheduleReportRecipient(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []string:
		if len(v) == 0 {
			return ""
		}
		return strings.TrimSpace(v[0])
	case []any:
		for _, item := range v {
			if recipient := strings.TrimSpace(insightsToolString(item)); recipient != "" {
				return recipient
			}
		}
	}
	return ""
}

func validateScheduleReportRecipient(channel, recipient string) error {
	if strings.TrimSpace(recipient) == "" {
		return errors.New("recipient is required")
	}
	switch strings.ToLower(strings.TrimSpace(recipient)) {
	case "me", "my email", "current user", "current_user":
		return errors.New("recipient must be an explicit email address or Slack channel")
	}
	switch channel {
	case "email":
		if _, err := mail.ParseAddress(recipient); err != nil {
			return errors.New("recipient must be a valid email address")
		}
	case "slack":
		if !strings.HasPrefix(recipient, "#") && !strings.HasPrefix(strings.ToUpper(recipient), "C") {
			return errors.New("recipient must be a Slack channel name or channel id")
		}
	}
	return nil
}

func scheduleReportOptionalBool(primary, fallback any) *bool {
	for _, value := range []any{primary, fallback} {
		switch v := value.(type) {
		case nil:
			continue
		case bool:
			result := v
			return &result
		case string:
			trimmed := strings.TrimSpace(strings.ToLower(v))
			if trimmed == "" {
				continue
			}
			result := trimmed != "false" && trimmed != "0" && trimmed != "no" && trimmed != "inactive"
			return &result
		}
	}
	return nil
}
