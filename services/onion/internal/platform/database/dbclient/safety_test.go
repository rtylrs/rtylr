package dbclient

import "testing"

func TestLooksLikeProductionHost(t *testing.T) {
	cases := map[string]bool{
		"voxire-db.rds.amazonaws.com:3306": true,
		"voxire-db.rds.amazonaws.com":      true,
		"VOXIRE-DB.RDS.AMAZONAWS.COM:3306": true,
		"localhost:3306":                   false,
		"127.0.0.1:3306":                   false,
		"db:3306":                          false,
		"host.internal:3306":               false,
		"":                                 false,
	}
	for host, want := range cases {
		if got := looksLikeProductionHost(host); got != want {
			t.Errorf("looksLikeProductionHost(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestCheckProductionSafety(t *testing.T) {
	const prodDSN = "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true"

	t.Run("blocks prod host without APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		if err := checkProductionSafety(prodDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("blocks prod host with APP_ENV unset", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		if err := checkProductionSafety(prodDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allows prod host with APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		if err := checkProductionSafety(prodDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("allows local host regardless of APP_ENV", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		dsn := "user:pw@tcp(localhost:3306)/onion_dev"
		if err := checkProductionSafety(dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("allows self-host docker host with APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		dsn := "user:pw@tcp(db:3306)/onion_self_host"
		if err := checkProductionSafety(dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unparseable DSN is not blocked", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		if err := checkProductionSafety("not-a-dsn"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
