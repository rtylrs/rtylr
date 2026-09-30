package config

import (
	"os"
	"strings"
)

// #nosec G101 -- This is an environment variable name, not a secret value.
const AppEnvEnvVar = "APP_ENV"

// nonProductionEnvironments is the allowlist of APP_ENV values that mean
// "this is a development machine". It is deliberately an allowlist of the
// *non*-production values rather than a list of production values: every
// development-only surface in the codebase hangs off !IsProduction(), so
// naming production values instead leaves the unsafe direction as the default.
// An unset, misspelled or unexpected APP_ENV then resolves to production and
// those surfaces stay off.
//
// Keep in step with cdnDevelopmentEnvironment in internal/cdn/auth.go, which
// already gates on the same labels with the same fail-closed shape.
var nonProductionEnvironments = map[string]struct{}{
	"dev":         {},
	"development": {},
	"local":       {},
	"test":        {},
}

// productionEnvironments are the labels that positively name production.
// They are only used to decide how loudly to report the resolved environment
// at boot; they are never consulted to decide whether the run is production.
var productionEnvironments = map[string]struct{}{
	"production": {},
	"prod":       {},
}

// NormalizeAppEnv folds an APP_ENV value to the form the allowlists compare
// against. Surrounding whitespace comes from shell exports, compose files and
// Helm values and is never a deliberate distinction, and case is not either:
// PROD and Production mean production, DEV means development.
func NormalizeAppEnv(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// CurrentAppEnv returns the normalized APP_ENV this process resolved to.
// A loaded configuration wins over the environment when it actually carries a
// value, matching how the other getters in this package read GlobalConfig; an
// empty AppEnv falls back to the environment rather than masking it.
func CurrentAppEnv() string {
	if GlobalConfig != nil {
		if v := strings.TrimSpace(GlobalConfig.AppEnv); v != "" {
			return NormalizeAppEnv(v)
		}
	}
	return NormalizeAppEnv(os.Getenv(AppEnvEnvVar))
}

// IsNonProduction reports whether APP_ENV explicitly names a development
// environment. Only the allowlist above counts, so unset, empty, misspelled
// and unexpected values are all non-development.
func IsNonProduction() bool {
	return isNonProductionEnv(CurrentAppEnv())
}

// IsProduction (declared in paddle.go, where it has always lived) is the exact
// negation of IsNonProduction, so no caller can reintroduce a "not development
// means production" test of its own.

// IsDeclaredProduction reports whether APP_ENV positively names production.
// This is the opposite question to IsProduction and is for callers that need a
// declaration of intent rather than an absence of a development one, such as a
// guard that refuses to reach a production endpoint unless the run says it is
// production. IsProduction() is the wrong tool there: it answers "not
// development", which is not a statement that the operator meant this.
func IsDeclaredProduction() bool {
	return isDeclaredProductionEnv(CurrentAppEnv())
}

// UnrecognizedAppEnv reports whether the resolved APP_ENV is neither an
// allowlisted development label nor an allowlisted production label. It does
// not imply anything unsafe — the value still resolves to production — it
// means the operator wrote something this build does not recognise, and that
// deserves to be said out loud at boot.
func UnrecognizedAppEnv() bool {
	return isUnrecognizedEnv(CurrentAppEnv())
}

// The three predicates take an already-normalized value rather than reading
// the environment themselves, so the boot assertion can classify the exact
// value it was handed instead of a second lookup that could disagree.

func isNonProductionEnv(normalized string) bool {
	_, ok := nonProductionEnvironments[normalized]
	return ok
}

func isDeclaredProductionEnv(normalized string) bool {
	_, ok := productionEnvironments[normalized]
	return ok
}

func isUnrecognizedEnv(normalized string) bool {
	if normalized == "" {
		return true
	}
	if isNonProductionEnv(normalized) || isDeclaredProductionEnv(normalized) {
		return false
	}
	return true
}
