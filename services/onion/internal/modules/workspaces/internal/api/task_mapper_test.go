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
