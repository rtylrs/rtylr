package dbclient

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

// knownProductionHosts are this application's production database endpoints.
//
// Matched exactly, not by an "*.rds.amazonaws.com" suffix: that suffix covers
// every standard AWS RDS endpoint in existence, so a staging, dev or self-host
// deployment running its own RDS instance would trip a guard that claims to be
// protecting *this* production database. Keep the regional endpoint alongside
// the legacy alias while both are supported.
var knownProductionHosts = map[string]struct{}{
	"voxire-db.rds.amazonaws.com":                        {},
	"voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com": {},
}

// envAllowProductionDB opts a run into the production database without renaming
// it "production". APP_ENV is not the only legitimate way to mean production:
// a staging tier or a release candidate may be deliberately pointed at the real
// database, and the guard cannot be taught every tier name in advance.
//
// Fail-closed: unset, or any value other than "true", leaves the guard armed.
const envAllowProductionDB = "ALLOW_PRODUCTION_DB"

// envAllowRemoteDB opts a development-labelled run into an off-box database.
// Sharing a development database between a team is legitimate; doing it by
// accident is what this guards. Fail-closed, parsed like envAllowProductionDB.
const envAllowRemoteDB = "ALLOW_REMOTE_DB"

// isProductionHost reports whether host (with or without a ":port" suffix) is
// the production database endpoint.
func isProductionHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	_, known := knownProductionHosts[strings.ToLower(strings.TrimSpace(host))]
	return known
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
	metadata, parseErr := parseDSNMetadata(dsn)
	if parseErr != nil || strings.TrimSpace(metadata.host) == "" {
		return fmt.Errorf(
			"refusing to connect: could not safely parse %s DSN; refusing an unknown database endpoint",
			role,
		)
	}
	host := metadata.host
	if !isProductionHost(host) {
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

// isLocalDBHost reports whether host (with or without a ":port" suffix) is on
// this machine or on the same container network.
//
// Local means one of: the name "localhost"; a loopback or unspecified IP; or a
// single-label hostname. That last case is load-bearing rather than lenient.
// self-host/docker-compose.yml sets DB_HOST: db for api, upload and worker, so
// a Docker service name has to count as local or every containerised deployment
// would be refused. A single label cannot be a public DNS name.
//
// Remote therefore means a dotted name or a routable IP, which is exactly the
// shape of the managed endpoint this guard exists to catch.
func isLocalDBHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	host = strings.ToLower(strings.TrimSpace(strings.Trim(host, "[]")))
	if host == "" {
		// An unknown endpoint is not evidence that it is local. The DSN parser
		// now rejects malformed input before this helper is reached, but keep
		// this branch fail-closed for direct callers as well.
		return false
	}

	if host == "localhost" {
		return true
	}

	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsUnspecified()
	}

	return !strings.Contains(host, ".")
}

// isDevelopmentEnv reports whether APP_ENV explicitly says this is development.
//
// Only an explicit label counts. An unset APP_ENV is deliberately not treated as
// development: it is the state in CI and in ad-hoc tooling, and firing there
// would refuse connections this guard has no opinion about. The mistake being
// caught is a file that says "development" while pointing somewhere else, so the
// declaration is the trigger.
func isDevelopmentEnv() bool {
	env := strings.ToLower(strings.TrimSpace(currentAppEnv()))

	return env == "development" || env == "dev"
}

// remoteDBAllowed reports whether this run has declared that a
// development-labelled process means to reach an off-box database.
func remoteDBAllowed() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(envAllowRemoteDB)), "true")
}

// checkDevelopmentSafety refuses to proceed when APP_ENV says development but
// dsn resolves to a host that is not on this machine.
//
// checkProductionSafety only fires on one exact known endpoint, so that a
// staging or self-host tier on its own RDS instance is not blocked by a guard
// claiming to protect *this* production database. The consequence is that a
// development-labelled config pointing at any other managed database connects
// without comment. This closes that gap from the other direction: rather than
// asking which host it is, it asks whether a run that calls itself development
// should be leaving the machine at all.
func checkDevelopmentSafety(role, dsn string) error {
	if !isDevelopmentEnv() {
		return nil
	}

	metadata, parseErr := parseDSNMetadata(dsn)
	if parseErr != nil || strings.TrimSpace(metadata.host) == "" {
		return fmt.Errorf(
			"refusing to connect: could not safely parse %s DSN; refusing an unknown database endpoint",
			role,
		)
	}
	host := metadata.host
	if isLocalDBHost(host) {
		return nil
	}

	if remoteDBAllowed() {
		return nil
	}

	return fmt.Errorf(
		"refusing to connect: APP_ENV=%q declares development, but %s host %q is not "+
			"on this machine.\n"+
			"If you meant to run against the local container, run `make local-env-up` and "+
			"copy .env.example (DB_HOST=localhost, DB_NAME=onion_dev).\n"+
			"If this tier is not development, set APP_ENV to the tier it is.\n"+
			"If a development run is genuinely meant to reach this database, set %s=true",
		currentAppEnv(), role, host, envAllowRemoteDB,
	)
}

// checkDSNSafety runs every guard that applies to a resolved DSN. Connect calls
// this rather than the individual checks so a new guard cannot be added to one
// of the three connections and forgotten on the other two.
func checkDSNSafety(role, dsn string) error {
	if err := checkProductionSafety(role, dsn); err != nil {
		return err
	}

	return checkDevelopmentSafety(role, dsn)
}
