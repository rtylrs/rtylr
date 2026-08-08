package dbclient

import (
	"strings"
	"testing"
)

func TestIsProductionHost(t *testing.T) {
	cases := map[string]bool{
		// the real production database, with and without a port, any casing
		"voxire-db.rds.amazonaws.com:3306": true,
		"voxire-db.rds.amazonaws.com":      true,
		"VOXIRE-DB.RDS.AMAZONAWS.COM:3306": true,

		// other RDS endpoints are NOT this app's production database: a staging
		// or self-host deployment on its own RDS instance must not be blocked
		"staging-db.rds.amazonaws.com:3306":        false,
		"voxire-db-replica.rds.amazonaws.com:3306": false,
		"some-other-app.rds.amazonaws.com":         false,
		"rds.amazonaws.com:3306":                   false,

		// a host that merely contains the prod host name is not the prod host
		"not-voxire-db.rds.amazonaws.com":               false,
		"voxire-db.rds.amazonaws.com.evil.example:3306": false,

		"localhost:3306":     false,
		"127.0.0.1:3306":     false,
		"db:3306":            false,
		"host.internal:3306": false,
		"":                   false,
	}
	for host, want := range cases {
		if got := isProductionHost(host); got != want {
			t.Errorf("isProductionHost(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestCheckProductionSafety(t *testing.T) {
	const prodDSN = "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true"

	t.Run("blocks prod host without APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", prodDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("blocks prod host with APP_ENV unset", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", prodDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allows prod host with APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", prodDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("allows local host regardless of APP_ENV", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		dsn := "user:pw@tcp(localhost:3306)/onion_dev"
		if err := checkProductionSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("allows self-host docker host with APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv(envAllowProductionDB, "")
		dsn := "user:pw@tcp(db:3306)/onion_self_host"
		if err := checkProductionSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unparseable DSN is not blocked", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", "not-a-dsn"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// Codex review: the *.rds.amazonaws.com suffix matched every standard AWS
	// RDS endpoint. A staging tier on its own RDS instance is not this app's
	// production database and must connect without any declaration.
	t.Run("allows a different RDS host outside production", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		t.Setenv(envAllowProductionDB, "")
		dsn := "user:pw@tcp(staging-db.rds.amazonaws.com:3306)/voxire_staging"
		if err := checkProductionSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// Mohammad's review: APP_ENV=production is not the only legitimate way to mean
// production. A staging tier or release candidate pointed at the real prod
// database needs a way to say so without being renamed "production".
func TestCheckProductionSafetyOverride(t *testing.T) {
	const prodDSN = "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true"

	t.Run("blocked when override is unset", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", prodDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allowed when override is set", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		t.Setenv(envAllowProductionDB, "true")
		if err := checkProductionSafety("database", prodDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("override is case-insensitive and space-tolerant", func(t *testing.T) {
		t.Setenv("APP_ENV", "release-candidate")
		t.Setenv(envAllowProductionDB, "  TRUE  ")
		if err := checkProductionSafety("database", prodDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// Fail-closed: only an explicit "true" opts in. Anything else stays blocked.
	t.Run("stays blocked for non-true values", func(t *testing.T) {
		for _, v := range []string{"false", "0", "1", "yes", "no", "maybe"} {
			t.Setenv("APP_ENV", "staging")
			t.Setenv(envAllowProductionDB, v)
			if err := checkProductionSafety("database", prodDSN); err == nil {
				t.Fatalf("expected error for %s=%q, got nil", envAllowProductionDB, v)
			}
		}
	})
}

// The refusal must show both ways out, so whoever hits it can tell which one
// applies to them without reading the source.
func TestCheckProductionSafetyErrorDocumentsBothOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv(envAllowProductionDB, "")

	err := checkProductionSafety("database", "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, want := range []string{"APP_ENV=production", envAllowProductionDB + "=true"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error message missing %q; got: %v", want, err)
		}
	}
}

// Codex review: Connect() opened LOGGING_DATABASE_URL with no guard at all, so
// a local main DSN plus a prod logging DSN silently reached production.
func TestLoggingDSNIsChecked(t *testing.T) {
	t.Run("blocks prod logging host outside production", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv("LOGGING_DATABASE_URL", "user:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire_logs")

		if err := checkProductionSafety("logging database", resolveLoggingDSN()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allows local logging host", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv("LOGGING_DATABASE_URL", "user:pw@tcp(localhost:3306)/onion_logs")

		if err := checkProductionSafety("logging database", resolveLoggingDSN()); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("names the logging database in the refusal", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv("LOGGING_DATABASE_URL", "user:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire_logs")

		err := checkProductionSafety("logging database", resolveLoggingDSN())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "logging database") {
			t.Errorf("refusal should name the logging database; got: %v", err)
		}
	})
}

// The AI / NL-to-SQL replica DSN is assembled from DB_REPLICA_HOST rather than
// read from a single URL, so it needs the same guard on the assembled result.
func TestReplicaDSNIsChecked(t *testing.T) {
	setReplicaEnv := func(t *testing.T, host string) {
		t.Helper()
		t.Setenv(envReplicaHost, host)
		t.Setenv(envAIUser, "ai_ro")
		t.Setenv(envAIPassword, "pw")
		t.Setenv("DB_NAME", "voxire")
	}

	t.Run("blocks prod replica host outside production", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		setReplicaEnv(t, "voxire-db.rds.amazonaws.com")

		if !replicaConfigured() {
			t.Fatal("expected replica to be configured")
		}
		if err := checkProductionSafety("read replica", buildReplicaDSN()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allows a non-production replica host", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		setReplicaEnv(t, "staging-db.rds.amazonaws.com")

		if err := checkProductionSafety("read replica", buildReplicaDSN()); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("allows prod replica host with the override set", func(t *testing.T) {
		t.Setenv("APP_ENV", "staging")
		t.Setenv(envAllowProductionDB, "true")
		setReplicaEnv(t, "voxire-db.rds.amazonaws.com")

		if err := checkProductionSafety("read replica", buildReplicaDSN()); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
