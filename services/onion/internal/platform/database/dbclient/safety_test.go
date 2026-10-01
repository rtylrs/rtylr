package dbclient

import (
	"strings"
	"testing"
)

func TestIsProductionHost(t *testing.T) {
	cases := map[string]bool{
		// the real production database, with and without a port, any casing
		"voxire-db.rds.amazonaws.com:3306":                        true,
		"voxire-db.rds.amazonaws.com":                             true,
		"voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com:3306": true,
		"VOXIRE-DB.RDS.AMAZONAWS.COM:3306":                        true,

		// other RDS endpoints are NOT this app's production database: a staging
		// or isolated deployment on its own RDS instance must not be blocked
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

	t.Run("allows local docker host with APP_ENV=production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv(envAllowProductionDB, "")
		dsn := "user:pw@tcp(db:3306)/onion_local"
		if err := checkProductionSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unparseable DSN is blocked instead of treated as local", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		if err := checkProductionSafety("database", "not-a-dsn"); err == nil {
			t.Fatal("expected malformed DSN to be rejected")
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

// isLocalDBHost decides whether checkDevelopmentSafety fires at all, so the
// single-label case matters as much as the loopback ones: local compose files
// set DB_HOST: db for api, upload and worker, and a guard that called a Docker
// service name "remote" would refuse every containerised deployment.
func TestIsLocalDBHost(t *testing.T) {
	local := []string{
		"localhost",
		"localhost:3306",
		"LocalHost",
		"127.0.0.1",
		"127.0.0.1:3306",
		"127.5.5.5",
		"::1",
		"[::1]:3306",
		"0.0.0.0",
		"db",          // local compose service name
		"db:3306",     //
		"mysql",       // the same shape under another name
		"onion_mysql", // the local-env volume/container naming
	}
	for _, h := range local {
		if !isLocalDBHost(h) {
			t.Errorf("isLocalDBHost(%q) = false, want true", h)
		}
	}

	remote := []string{
		"voxire-db.rds.amazonaws.com",
		"voxire-db.rds.amazonaws.com:3306",
		"staging-db.rds.amazonaws.com",
		"db.internal",
		"10.0.0.4",
		"192.168.1.50",
		"192.168.1.50:3306",
		"", // unknown endpoints must not be treated as local
	}
	for _, h := range remote {
		if isLocalDBHost(h) {
			t.Errorf("isLocalDBHost(%q) = true, want false", h)
		}
	}
}

// The combination this exists for: a file that says development while pointing
// at a managed database. checkProductionSafety cannot catch it unless the host
// happens to be the one known production endpoint.
func TestCheckDevelopmentSafety(t *testing.T) {
	const remoteDSN = "user:pw@tcp(some-other-db.rds.amazonaws.com:3306)/voxire"

	t.Run("blocks a remote host when APP_ENV=development", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowRemoteDB, "")
		if err := checkDevelopmentSafety("database", remoteDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("blocks a remote host when APP_ENV=dev", func(t *testing.T) {
		t.Setenv("APP_ENV", "dev")
		t.Setenv(envAllowRemoteDB, "")
		if err := checkDevelopmentSafety("database", remoteDSN); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("allows the local container in development", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowRemoteDB, "")
		dsn := "onion_user:onion_password@tcp(localhost:3306)/onion_dev"
		if err := checkDevelopmentSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	// Regression guard: a containerised development run reaches its database by
	// service name. Refusing that would break local compose setups.
	t.Run("allows a docker service name in development", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowRemoteDB, "")
		dsn := "user:pw@tcp(db:3306)/onion_dev"
		if err := checkDevelopmentSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("says nothing about a remote host outside development", func(t *testing.T) {
		for _, env := range []string{"production", "staging", "release-candidate"} {
			t.Setenv("APP_ENV", env)
			t.Setenv(envAllowRemoteDB, "")
			if err := checkDevelopmentSafety("database", remoteDSN); err != nil {
				t.Errorf("APP_ENV=%q: expected no error, got %v", env, err)
			}
		}
	})

	// Deliberate: unset is the state in CI and ad-hoc tooling. The trigger is a
	// declaration of development, not the absence of one.
	t.Run("does not fire when APP_ENV is unset", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		t.Setenv(envAllowRemoteDB, "")
		if err := checkDevelopmentSafety("database", remoteDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("names the connection being refused", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowRemoteDB, "")
		err := checkDevelopmentSafety("read replica", remoteDSN)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "read replica") {
			t.Errorf("refusal should name the connection; got: %v", err)
		}
		if !strings.Contains(err.Error(), envAllowRemoteDB) {
			t.Errorf("refusal should name the way out; got: %v", err)
		}
	})
}

func TestCheckDevelopmentSafetyOverride(t *testing.T) {
	const remoteDSN = "user:pw@tcp(shared-dev-db.example.com:3306)/onion_dev"

	t.Run("allows a remote host once declared", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowRemoteDB, "true")
		if err := checkDevelopmentSafety("database", remoteDSN); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("accepts any casing of true", func(t *testing.T) {
		for _, v := range []string{"TRUE", "True", " true "} {
			t.Setenv("APP_ENV", "development")
			t.Setenv(envAllowRemoteDB, v)
			if err := checkDevelopmentSafety("database", remoteDSN); err != nil {
				t.Errorf("%s=%q: expected no error, got %v", envAllowRemoteDB, v, err)
			}
		}
	})

	// Fail-closed, matching envAllowProductionDB: only an explicit "true" opts in.
	t.Run("stays armed for anything that is not true", func(t *testing.T) {
		for _, v := range []string{"1", "yes", "on", "false", "maybe"} {
			t.Setenv("APP_ENV", "development")
			t.Setenv(envAllowRemoteDB, v)
			if err := checkDevelopmentSafety("database", remoteDSN); err == nil {
				t.Errorf("%s=%q: expected error, got nil", envAllowRemoteDB, v)
			}
		}
	})
}

// Connect calls checkDSNSafety, not the individual guards. These assert the
// wrapper actually runs both, since a guard wired into one of the three
// connections and not the others is the failure this wrapper prevents.
func TestCheckDSNSafetyRunsBothGuards(t *testing.T) {
	t.Run("catches the known production host", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv(envAllowRemoteDB, "")
		dsn := "user:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire"
		if err := checkDSNSafety("database", dsn); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	// The gap this issue is about: checkProductionSafety alone passes this,
	// because the host is not the one endpoint it knows.
	t.Run("catches a remote host the production guard allows", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv(envAllowRemoteDB, "")
		dsn := "user:pw@tcp(some-other-db.rds.amazonaws.com:3306)/voxire"

		if err := checkProductionSafety("database", dsn); err != nil {
			t.Fatalf("precondition: the production guard should allow this, got %v", err)
		}
		if err := checkDSNSafety("database", dsn); err == nil {
			t.Fatal("expected checkDSNSafety to refuse, got nil")
		}
	})

	t.Run("allows the documented local setup", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv(envAllowProductionDB, "")
		t.Setenv(envAllowRemoteDB, "")
		dsn := "onion_user:onion_password@tcp(localhost:3306)/onion_dev"
		if err := checkDSNSafety("database", dsn); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
