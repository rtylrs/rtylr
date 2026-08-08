package dbclient

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

// productionHostSuffix matches this project's RDS naming convention
// (e.g. voxire-db.rds.amazonaws.com). Pattern-matched rather than allow-listing
// local hosts, so new local/self-host conventions never need to be taught here.
const productionHostSuffix = ".rds.amazonaws.com"

// looksLikeProductionHost reports whether host (with or without a ":port"
// suffix) matches the production database naming convention.
func looksLikeProductionHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.HasSuffix(strings.ToLower(host), productionHostSuffix)
}

// checkProductionSafety refuses to proceed when dsn resolves to what looks
// like the production database but nothing declares this run as production.
//
// APP_ENV alone isn't a safe signal: self-host deploys set APP_ENV=production
// while still pointing at a local Docker database. So the guard only fires on
// the combination that actually indicates a mistake: a production-shaped host
// with no matching declaration that this is really production. A real
// production deploy (APP_ENV=production) passes through untouched.
func checkProductionSafety(dsn string) error {
	host := hostFromDSN(dsn)
	if host == "" || !looksLikeProductionHost(host) {
		return nil
	}

	if config.IsProduction() {
		return nil
	}

	return fmt.Errorf(
		"refusing to connect: database host %q looks like the production database "+
			"(matches *%s) but APP_ENV=%q is not \"production\" — this looks like a "+
			"local/dev run about to hit the live database.\n"+
			"If this really is production, set APP_ENV=production.\n"+
			"If this is a local/dev run, point DATABASE_URL/DB_HOST at a local database (see .env.example)",
		host, productionHostSuffix, currentAppEnv(),
	)
}

func currentAppEnv() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.AppEnv
	}
	return os.Getenv("APP_ENV")
}
