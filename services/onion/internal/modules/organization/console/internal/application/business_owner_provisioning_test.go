package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateBusinessOwnerProvisionsEmployeeThroughHRCore(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	core := hrcore.NewModule(ctx, db)
	service := application.NewService(ctx, repo).WithEmployeeProvisioner(core)
	stamp := time.Now().UnixNano()
	input := application.CreateBusinessOwnerInput{
		Email:        fmt.Sprintf("console-owner-%d@example.test", stamp),
		Password:     "test-password-123",
		FullName:     "Grace Hopper",
		BusinessName: fmt.Sprintf("Console Bootstrap %d", stamp),
	}

	created, err := service.CreateBusinessOwner(ctx, input)
	if err != nil {
		t.Fatalf("create business owner: %v", err)
	}
	t.Cleanup(func() {
		if err := core.DeleteEmployeeForUser(context.Background(), created.Owner.ID, created.Business.ID, created.Owner.ID); err != nil {
			t.Errorf("delete HR employee: %v", err)
		}
		if err := repo.DeleteCreatedBusinessOwner(context.Background(), created.Owner.ID, created.Business.ID); err != nil {
			t.Errorf("delete created business owner: %v", err)
		}
	})

	var firstName, lastName, email string
	if err := db.QueryRowContext(ctx, `
		SELECT first_name, last_name, email
		FROM hr_employees WHERE business_id = ? AND user_id = ?`,
		created.Business.ID, created.Owner.ID).Scan(&firstName, &lastName, &email); err != nil {
		t.Fatalf("read owner employee: %v", err)
	}
	if firstName != "Grace" || lastName != "Hopper" || email != input.Email {
		t.Fatalf("owner employee = %q %q <%s>", firstName, lastName, email)
	}
}
