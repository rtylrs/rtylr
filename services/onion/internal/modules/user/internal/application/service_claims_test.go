package application

import (
	"encoding/json"
	"testing"
)

func TestAccessTokenClaimsUnmarshalAcceptsNumericAndStringIDs(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "numeric claims",
			json: `{"user_id":10,"business_id":20,"platform_user_id":30,"plan_id":40}`,
		},
		{
			name: "string claims",
			json: `{"user_id":"10","business_id":"20","platform_user_id":"30","plan_id":"40"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var claims AccessTokenClaims
			if err := json.Unmarshal([]byte(tt.json), &claims); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if claims.UserID != 10 || claims.BusinessID != 20 || claims.PlatformUserID != 30 || claims.PlanID != 40 {
				t.Fatalf("unexpected claims: %+v", claims)
			}
		})
	}
}

func TestAccessTokenClaimsUnmarshalRejectsNonIntegerIDs(t *testing.T) {
	var claims AccessTokenClaims
	if err := json.Unmarshal([]byte(`{"user_id":"not-an-id"}`), &claims); err == nil {
		t.Fatal("expected invalid user_id to be rejected")
	}
}
