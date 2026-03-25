package application

import (
	"context"
	"os"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/alerts"
)

type Service struct {
	alertsSvc *alerts.Service
}

func NewService(ctx context.Context) *Service {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	dashboardChannelID := os.Getenv("SLACK_DASHBOARD_CHANNEL_ID")
	alertsChannelID := os.Getenv("SLACK_ALERTS_CHANNEL_ID")

	alertsSvc := alerts.NewAlertsService(ctx, slackBotToken)
	if dashboardChannelID != "" {
		alertsSvc.AddChannel("dashboard-errors", dashboardChannelID)
	}

	if alertsChannelID != "" {
		alertsSvc.AddChannel("voxire-errors", alertsChannelID)
	}

	return &Service{
		alertsSvc: alertsSvc,
	}
}

func (s *Service) SendFormattedDashboardAlertsMessage(ctx context.Context, message string) {
	if message == "" {
		return
	}

	if err := s.alertsSvc.SendMessage(ctx, "dashboard-errors", message); err != nil {
		logger.Ctx(ctx).Error("failed to send dashboard alerts message", logger.Error(err))
	}
}

func (s *Service) SendFormattedVoxireAlertsMessage(ctx context.Context, message string) {
	if message == "" {
		return
	}

	if err := s.alertsSvc.SendMessage(ctx, "voxire-errors", message); err != nil {
		logger.Ctx(ctx).Error("failed to send voxire alerts message", logger.Error(err))
	}
}
