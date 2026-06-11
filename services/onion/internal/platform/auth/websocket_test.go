package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

func signedWebSocketToken(t *testing.T, claims jwtutils.Claims) string {
	t.Helper()

	token, err := jwtutils.NewJWTHelperWithDuration("ws-secret", webSocketJWTIssuer, time.Hour).CreateJWT(claims)
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	return token
}

func wsRequest(path string) *http.Request {
	ctx := context.WithValue(context.Background(), jwtutils.JWTSecretKey, "ws-secret")
	return httptest.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
}

func TestAuthUserFromWebSocketRequest_QueryToken(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{
		UserID:         7,
		BusinessID:     3,
		Role:           "staff",
		Platform:       "rtylr",
		PlatformUserID: 42,
		PlanID:         9,
	})

	req := wsRequest("/flow/ws?token=" + token)
	req.RemoteAddr = "1.2.3.4:12345"

	user, err := AuthUserFromWebSocketRequest(req)
	if err != nil {
		t.Fatalf("AuthUserFromWebSocketRequest: %v", err)
	}
	if user.UserID != 7 || user.BusinessID != 3 || user.Role != "staff" {
		t.Fatalf("unexpected auth user: %+v", user)
	}
	if user.Platform != "rtylr" || user.PlatformUserID != 42 || user.PlanID != 9 {
		t.Fatalf("platform claims were not preserved: %+v", user)
	}
	if user.Token != token || user.IP != "1.2.3.4:12345" {
		t.Fatalf("token/IP not preserved: %+v", user)
	}
}

func TestAuthUserFromWebSocketRequest_BearerFallback(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{UserID: 8, BusinessID: 4, Role: "admin"})
	req := wsRequest("/dash/ws")
	req.Header.Set("Authorization", "Bearer "+token)

	user, err := AuthUserFromWebSocketRequest(req)
	if err != nil {
		t.Fatalf("AuthUserFromWebSocketRequest: %v", err)
	}
	if user.UserID != 8 || user.BusinessID != 4 || user.Role != "admin" {
		t.Fatalf("unexpected auth user: %+v", user)
	}
}

func TestAuthUserFromWebSocketRequest_MissingToken(t *testing.T) {
	_, err := AuthUserFromWebSocketRequest(wsRequest("/pos/ws"))
	if !errors.Is(err, ErrWebSocketTokenMissing) {
		t.Fatalf("expected ErrWebSocketTokenMissing, got %v", err)
	}
}

func TestAuthUserFromWebSocketRequest_MissingSecret(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{UserID: 1})
	req := httptest.NewRequest(http.MethodGet, "/pos/ws?token="+token, http.NoBody)

	_, err := AuthUserFromWebSocketRequest(req)
	if !errors.Is(err, ErrJWTSecretMissing) {
		t.Fatalf("expected ErrJWTSecretMissing, got %v", err)
	}
}

func TestAuthUserFromWebSocketRequest_MissingUser(t *testing.T) {
	token := signedWebSocketToken(t, &webSocketClaims{})
	req := wsRequest("/pos/ws?token=" + token)

	_, err := AuthUserFromWebSocketRequest(req)
	if !errors.Is(err, ErrWebSocketUserMissing) {
		t.Fatalf("expected ErrWebSocketUserMissing, got %v", err)
	}
}
