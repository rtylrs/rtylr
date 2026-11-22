package main

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/team"
	"github.com/voxire/onion/internal/modules/pager"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

func TestJobModulesRegistersTerminatedAccessReconciliation(t *testing.T) {
	registry := platformjobs.NewRegistry()
	modules := &jobModules{Team: team.NewModule(context.Background(), nil)}

	platformjobs.AutoRegisterInto(context.Background(), modules, registry)

	if _, ok := registry.Get("team.terminated_access_reconciliation"); !ok {
		t.Fatal("worker job modules did not register team.terminated_access_reconciliation")
	}
}

func TestJobModulesRegistersPagerSweep(t *testing.T) {
	registry := platformjobs.NewRegistry()
	modules := &jobModules{Pager: pager.NewModule(context.Background(), nil, nil)}

	platformjobs.AutoRegisterInto(context.Background(), modules, registry)

	if _, ok := registry.Get("pager.sweep"); !ok {
		t.Fatal("worker job modules did not register pager.sweep")
	}
}
