package dbclient

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

// knownProductionHost is this application's production database endpoint.
//
// Matched exactly, not by an "*.rds.amazonaws.com" suffix: that suffix covers
// every standard AWS RDS endpoint in existence, so a staging, dev or self-host
// deployment running its own RDS instance would trip a guard that claims to be
// protecting *this* production database. An exact host is the only match whose
// meaning matches what the refusal says.
//
// If production ever moves, this constant moves with it. A stale value fails
// open (the guard stops firing) rather than blocking a healthy deploy.
const knownProductionHost = "voxire-db.rds.amazonaws.com"

// envAllowProductionDB opts a run into the production database without renaming
// it "production". APP_ENV is not the only legitimate way to mean production:
// a staging tier or a release candidate may be deliberately pointed at the real
// database, and the guard cannot be taught every tier name in advance.
//
// Fail-closed: unset, or any value other than "true", leaves the guard armed.
const envAllowProductionDB = "ALLOW_PRODUCTION_DB"

// isProductionHost reports whether host (with or without a ":port" suffix) is
// the production database endpoint.
func isProductionHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	return strings.EqualFold(strings.TrimSpace(host), knownProductionHost)
}

// productionDBAllowed reports whether this run has explicitly declared that it
// means to use the production database. Parsed the same way as the other
// boolean env overrides in internal/platform/config.
func productionDBAllowed() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(envAllowProductionDB)), "true")
}

// checkProductionSafety refuses to proceed when dsn resolves to the production
// database but nothing declares that this run means to use it. role names which
// connection is being checked ("database", "logging database", "read replica")
// so the refusal points at the right environment variable.
//
// APP_ENV alone isn't a safe signal in either direction. Self-host deploys set
// APP_ENV=production while pointing at a local Docker database, so it can't be
// required; and tiers that aren't named "production" may still legitimately
// target the real database, so it can't be the only way in. The guard fires
// only on the combination that actually indicates a mistake: the production
// host, with no declaration of intent from either signal.
func checkProductionSafety(role, dsn string) error {
	host := hostFromDSN(dsn)
	if host == "" || !isProductionHost(host) {
		return nil
	}

	if config.IsProduction() || productionDBAllowed() {
		return nil
	}

	return fmt.Errorf(
		"refusing to connect: %s host %q is the production database, but this run "+
			"has not declared that it means to use it (APP_ENV=%q).\n"+
			"If this is a real production deploy, set APP_ENV=production.\n"+
			"If this is another tier (staging, release candidate) that intentionally "+
			"targets the production database, set %s=true.\n"+
			"If this is a local/dev run, point DATABASE_URL/DB_HOST at a local database (see .env.example)",
		role, host, currentAppEnv(), envAllowProductionDB,
	)
}

func currentAppEnv() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.AppEnv
	}

	return os.Getenv("APP_ENV")
}
