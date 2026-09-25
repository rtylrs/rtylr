package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

func TestTaskMappersIncludeDependencies(t *testing.T) {
	task := domain.Task{ID: 42, Dependencies: []int64{7, 9}}

	v1 := toTask(task)
	if len(v1.Dependencies) != 2 || v1.Dependencies[0] != 7 || v1.Dependencies[1] != 9 {
		t.Fatalf("v1 dependencies = %v, want [7 9]", v1.Dependencies)
	}

	v2 := toTaskV2(task)
	if len(v2.Dependencies) != 2 || v2.Dependencies[0] != 7 || v2.Dependencies[1] != 9 {
		t.Fatalf("v2 dependencies = %v, want [7 9]", v2.Dependencies)
	}
}

func TestTaskMapperV2CarriesTodayFields(t *testing.T) {
	minutes := int32(45)
	task := domain.Task{ID: 42, EffortMinutes: &minutes, FocusPin: true}

	mapped := toTaskV2(task)
	if mapped.EffortMinutes == nil || *mapped.EffortMinutes != minutes {
		t.Fatalf("v2 effort_minutes = %v, want %d", mapped.EffortMinutes, minutes)
	}
	if !mapped.FocusPin {
		t.Fatal("v2 focus_pin = false, want true")
	}
}
