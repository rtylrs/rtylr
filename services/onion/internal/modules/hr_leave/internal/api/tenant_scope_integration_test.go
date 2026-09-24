package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestLeaveTypeHandlersResolveBusinessIDBeforeReadingOrWriting(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	stamp := time.Now().UnixNano()
	callerUserID := int64(980000 + stamp%1000)

	insertBusiness := func(label string) int64 {
		result, err := db.ExecContext(ctx, `INSERT INTO businesses (name, owner_id, is_active) VALUES (?, ?, TRUE)`, fmt.Sprintf("leave-scope-%s-%d", label, stamp), callerUserID)
		if err != nil {
			t.Fatalf("insert %s business: %v", label, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("read %s business id: %v", label, err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", id) })
		return id
	}
	ownBusinessID := insertBusiness("own")
	otherBusinessID := insertBusiness("other")
	if callerUserID == ownBusinessID {
		t.Fatal("test setup must use a user id distinct from the business id")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'admin')`, callerUserID, ownBusinessID); err != nil {
		t.Fatalf("insert own business membership: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?", callerUserID, ownBusinessID)
	})

	repo := infrastructure.NewRepository(db)
	ownType, err := repo.CreateLeaveType(ctx, ownBusinessID, domain.LeaveTypeRecord{
		Name: fmt.Sprintf("scope own %d", stamp), IsPaid: true, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create own leave type: %v", err)
	}
	foreignType, err := repo.CreateLeaveType(ctx, otherBusinessID, domain.LeaveTypeRecord{
		Name: fmt.Sprintf("scope foreign %d", stamp), IsPaid: true, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create foreign leave type: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id IN (?, ?)", ownType.ID, foreignType.ID)
	})

	service := application.NewService(ctx, db, nil, nil)
	access := &authPkg.AuthUser{
		UserID: callerUserID, BusinessID: ownBusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: ownBusinessID, IsAdmin: true, AuthorizationVersion: 1,
		},
	}
	authCtx := context.WithValue(ctx, authPkg.AuthUserKey, access)

	listed, err := (&ListLeaveTypesHandler{Service: service}).Serve(authCtx, &ListLeaveTypesRequest{})
	if err != nil {
		t.Fatalf("list leave types: %v", err)
	}
	if len(listed.Types) != 1 || listed.Types[0].Id != ownType.ID {
		t.Fatalf("listed leave types = %+v, want only own type %d", listed.Types, ownType.ID)
	}

	_, err = (&UpdateLeaveTypeHandler{Service: service}).Serve(authCtx, &UpdateLeaveTypeRequest{ID: foreignType.ID, Name: stringPtr("hijacked")})
	if oneKitStatus(err) != http.StatusNotFound {
		t.Fatalf("foreign update status = %d (%v), want 404", oneKitStatus(err), err)
	}
	got, getErr := repo.GetLeaveType(ctx, otherBusinessID, foreignType.ID)
	if getErr != nil {
		t.Fatalf("read foreign type after denied update: %v", getErr)
	}
	if got.Name == "hijacked" {
		t.Fatal("foreign leave type was modified")
	}
}

func stringPtr(value string) *string { return &value }
