package keys

import "testing"

func TestUserMeForBusinessScopesByUserAndBusiness(t *testing.T) {
	got := UserMeForBusiness(42, 7)
	want := "cache:user:42:business:7:me"
	if got != want {
		t.Fatalf("UserMeForBusiness = %q, want %q", got, want)
	}
}
