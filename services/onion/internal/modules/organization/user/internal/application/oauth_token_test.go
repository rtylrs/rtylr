package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

func oauthTestService() *Service {
	return &Service{
		JWT:      jwtutils.NewJWTHelperWithDuration("test-secret", "onion-platform", time.Hour),
		OAuthJWT: jwtutils.NewJWTHelperWithDuration("test-secret", OAuthTokenIssuer, time.Hour),
	}
}

func TestValidateOAuthTokenReadsClientAndScope(t *testing.T) {
	s := oauthTestService()
	token, err := s.OAuthJWT.CreateJWT(&AccessTokenClaims{UserID: 7, BusinessID: 3, OAuthClientID: "rtylr-mcp", Scope: "crm:read"})
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	claims, err := s.validateOAuthToken(context.Background(), token)
	if err != nil {
		t.Fatalf("validateOAuthToken: %v", err)
	}
	if claims.UserID != 7 || claims.OAuthClientID != "rtylr-mcp" || claims.Scope != "crm:read" {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestValidateOAuthTokenWithoutClientClaimStillRestricts(t *testing.T) {
	s := oauthTestService()
	token, err := s.OAuthJWT.CreateJWT(&AccessTokenClaims{UserID: 7})
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	claims, err := s.validateOAuthToken(context.Background(), token)
	if err != nil {
		t.Fatalf("validateOAuthToken: %v", err)
	}
	if claims.OAuthClientID == "" {
		t.Fatal("an OAuth-issuer token without a client claim must not validate as an unrestricted session")
	}
}

func TestValidateOAuthTokenRejectsSessionTokens(t *testing.T) {
	s := oauthTestService()
	session, err := s.JWT.CreateJWT(&AccessTokenClaims{UserID: 7})
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	if _, err := s.validateOAuthToken(context.Background(), session); err == nil {
		t.Fatal("a session token validated through the OAuth path")
	}
}

func TestValidateOAuthTokenRejectsForeignSignature(t *testing.T) {
	s := oauthTestService()
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": OAuthTokenIssuer, "user_id": 7, "cid": "rtylr-mcp", "scope": "crm:write",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := forged.SignedString([]byte("another-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.validateOAuthToken(context.Background(), signed); err == nil {
		t.Fatal("forged OAuth token validated")
	}
}

func TestValidateOAuthTokenUnavailableWithoutHelper(t *testing.T) {
	s := &Service{}
	if _, err := s.validateOAuthToken(context.Background(), "x"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
}
