package api

import (
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/crm/campaigns/internal/infrastructure"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func campaignWriteContext(businessID, userID, id int64) context.Context {
	user := &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: businessID,
			IsOwner:    true,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppCRM,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"crm:campaigns:create", "crm:campaigns:send"},
			}},
		},
	}

	rctx := map[string]string{}
	rctx["id"] = (strconv.FormatInt(id, 10))
	return authPkg.WithAuthUser(serverHttp.WithPathParams(context.Background(), rctx), user)
}

func campaignHTTPStatus(err error) int {
	var httpErr *onekithttp.Error
	if errors.As(err, &httpErr) {
		return int(httpErr.GetStatusCode())
	}
	return 0
}

// A retried update that re-sends the stored values changes zero rows; it must
// answer success instead of failing the request.
func TestUpdateCampaignHandler_ResentPayload_Succeeds(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(983001)
		businessID = int64(983002)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	service := application.NewService(ctx, moduleinfra.NewRepository(db))
	id, err := service.Create(ctx, domain.CampaignInput{
		OwnerID:    userID,
		BusinessID: businessID,
		Name:       "Handler Idempotent Campaign",
		Type:       "email",
		Subject:    "Hello",
		Body:       "Body content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_campaigns WHERE id = ?", id)
	})

	handler := &UpdateCampaignHandler{Service: service}
	req := &UpdateCampaignRequest{
		Name:    "Handler Idempotent Campaign",
		Type:    "email",
		Subject: "Hello",
		Body:    "Body content",
	}
	reqCtx := campaignWriteContext(businessID, userID, id)
	if _, err := handler.Serve(reqCtx, req); err != nil {
		t.Fatalf("first Serve: %v", err)
	}
	if _, err := handler.Serve(reqCtx, req); err != nil {
		t.Fatalf("re-sent Serve: %v", err)
	}
}

func TestUpdateCampaignHandler_MissingCampaign_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)

	const (
		userID     = int64(983011)
		businessID = int64(983012)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &UpdateCampaignHandler{Service: application.NewService(context.Background(), moduleinfra.NewRepository(db))}
	_, err := handler.Serve(
		campaignWriteContext(businessID, userID, 999999983),
		&UpdateCampaignRequest{Name: "Ghost Campaign", Type: "email"},
	)
	if got := campaignHTTPStatus(err); got != http.StatusNotFound {
		t.Fatalf("Serve on a missing campaign = %d, want 404 (err=%v)", got, err)
	}
}

func TestUpdateCampaignStatusHandler_TrimsStatusAndMapsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(983021)
		businessID = int64(983022)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	service := application.NewService(ctx, moduleinfra.NewRepository(db))
	id, err := service.Create(ctx, domain.CampaignInput{
		OwnerID:    userID,
		BusinessID: businessID,
		Name:       "Handler Status Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_campaigns WHERE id = ?", id)
	})

	handler := &UpdateCampaignStatusHandler{Service: service}
	reqCtx := campaignWriteContext(businessID, userID, id)
	if _, err := handler.Serve(reqCtx, &UpdateCampaignStatusRequest{Status: "paused"}); err != nil {
		t.Fatalf("Serve(paused): %v", err)
	}
	// The handler validates the trimmed status, so the domain validator must
	// see the trimmed value too; an untrimmed status previously surfaced as 400.
	if _, err := handler.Serve(reqCtx, &UpdateCampaignStatusRequest{Status: " paused "}); err != nil {
		t.Fatalf("Serve(\" paused \"): %v", err)
	}

	_, err = handler.Serve(
		campaignWriteContext(businessID, userID, 999999983),
		&UpdateCampaignStatusRequest{Status: "paused"},
	)
	if got := campaignHTTPStatus(err); got != http.StatusNotFound {
		t.Fatalf("Serve on a missing campaign = %d, want 404 (err=%v)", got, err)
	}
}
