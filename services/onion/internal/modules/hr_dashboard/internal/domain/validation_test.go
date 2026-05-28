package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
)

func TestIsValidWidgetType_AllValid(t *testing.T) {
	for _, wt := range []string{
		domain.WidgetTypeHeadcount,
		domain.WidgetTypeAttendanceToday,
		domain.WidgetTypeOpenLeaveRequests,
		domain.WidgetTypeUpcomingShifts,
		domain.WidgetTypePayrollSummary,
		domain.WidgetTypeOvertimeStats,
		domain.WidgetTypeRecentHires,
		domain.WidgetTypeTasksDue,
	} {
		if !domain.IsValidWidgetType(wt) {
			t.Errorf("expected %q to be valid", wt)
		}
	}
}

func TestIsValidWidgetType_Invalid(t *testing.T) {
	if domain.IsValidWidgetType("calendar") {
		t.Error("expected 'calendar' to be invalid")
	}
}

func TestValidateWidget_Valid(t *testing.T) {
	w := &domain.DashboardWidget{
		WidgetType: domain.WidgetTypeHeadcount,
		PositionX:  0,
		PositionY:  0,
		Width:      2,
		Height:     2,
	}
	if err := domain.ValidateWidget(w); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateWidget_InvalidType(t *testing.T) {
	w := &domain.DashboardWidget{
		WidgetType: "unknown_widget",
		Width:      2,
		Height:     2,
	}
	if err := domain.ValidateWidget(w); !errors.Is(err, domain.ErrInvalidWidgetType) {
		t.Errorf("expected ErrInvalidWidgetType, got %v", err)
	}
}

func TestValidateWidget_EmptyTypeAllowed(t *testing.T) {
	w := &domain.DashboardWidget{
		WidgetType: "",
		Width:      1,
		Height:     1,
	}
	if err := domain.ValidateWidget(w); err != nil {
		t.Errorf("empty widget type should be allowed, got %v", err)
	}
}

func TestValidateWidget_NegativePositionX(t *testing.T) {
	w := &domain.DashboardWidget{
		WidgetType: domain.WidgetTypeHeadcount,
		PositionX:  -1,
		PositionY:  0,
		Width:      2,
		Height:     2,
	}
	if err := domain.ValidateWidget(w); !errors.Is(err, domain.ErrInvalidWidgetPosition) {
		t.Errorf("expected ErrInvalidWidgetPosition, got %v", err)
	}
}

func TestValidateWidget_NegativePositionY(t *testing.T) {
	w := &domain.DashboardWidget{
		WidgetType: domain.WidgetTypeHeadcount,
		PositionX:  0,
		PositionY:  -5,
		Width:      2,
		Height:     2,
	}
	if err := domain.ValidateWidget(w); !errors.Is(err, domain.ErrInvalidWidgetPosition) {
		t.Errorf("expected ErrInvalidWidgetPosition, got %v", err)
	}
}

func TestValidateWidget_NegativeWidth(t *testing.T) {
	w := &domain.DashboardWidget{
		PositionX: 0,
		PositionY: 0,
		Width:     -1,
		Height:    2,
	}
	if err := domain.ValidateWidget(w); !errors.Is(err, domain.ErrInvalidWidgetDimension) {
		t.Errorf("expected ErrInvalidWidgetDimension, got %v", err)
	}
}
