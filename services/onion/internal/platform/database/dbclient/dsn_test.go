package dbclient

import (
	"strings"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
)

func TestHostFromDSN(t *testing.T) {
	cases := map[string]string{
		// native go-sql-driver form
		"bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true": "voxire-db.rds.amazonaws.com:3306",
		"user:pw@tcp(localhost:3306)/onion_dev":                                 "localhost:3306",
		// URL form, with and without credentials
		"mysql://bolted:pw@voxire-db.rds.amazonaws.com:3306/voxire?tls=true": "voxire-db.rds.amazonaws.com:3306",
		"mysql://host.internal:3306/db":                                      "host.internal:3306",
		// unparseable → empty so callers fall back to env
		"not-a-dsn": "",
	}
	for dsn, want := range cases {
		if got := hostFromDSN(dsn); got != want {
			t.Errorf("hostFromDSN(%q) = %q, want %q", dsn, got, want)
		}
	}
}

func TestDBNameFromDSN(t *testing.T) {
	cases := map[string]string{
		"bolted:pw@tcp(host:3306)/voxire?parseTime=true": "voxire",
		"mysql://bolted:pw@host:3306/voxire?tls=true":    "voxire",
		"user:pw@tcp(localhost:3306)/onion_dev":          "onion_dev",
	}
	for dsn, want := range cases {
		if got := dbNameFromDSN(dsn); got != want {
			t.Errorf("dbNameFromDSN(%q) = %q, want %q", dsn, got, want)
		}
	}
}

func TestNormalizeDSNPinsMySQLSessionTimezoneToUTC(t *testing.T) {
	for _, dsn := range []string{
		"user:pw@tcp(localhost:3306)/onion_dev",
		"user:pw@tcp(localhost:3306)/onion_dev?parseTime=true",
		"mysql://user:pw@localhost:3306/onion_dev?parseTime=true",
	} {
		got := normalizeDSN(dsn)
		if !strings.Contains(got, "time_zone=%27UTC%27") {
			t.Errorf("normalizeDSN(%q) missing UTC session timezone: %q", dsn, got)
		}
	}
}

func TestDSNMetadataDoesNotTreatCredentialPunctuationAsEndpoint(t *testing.T) {
	cases := map[string]string{
		"u:tcp(foo)@tcp(host.example:3306)/db":            "host.example:3306",
		"u:p?ss@tcp(host.example:3306)/db?parseTime=true": "host.example:3306",
		"u:p@ss@tcp(host.example:3306)/db":                "host.example:3306",
		"u:p/ss@tcp(host.example:3306)/db":                "host.example:3306",
		"u:mysql://ss@tcp(host.example:3306)/db":          "host.example:3306",
	}
	for dsn, wantHost := range cases {
		if got := hostFromDSN(dsn); got != wantHost {
			t.Errorf("hostFromDSN(%q) = %q, want %q", dsn, got, wantHost)
		}
		if got := dbNameFromDSN(dsn); got != "db" {
			t.Errorf("dbNameFromDSN(%q) = %q, want db", dsn, got)
		}
	}
}

func TestNormalizeDSNPreservesCredentialPunctuation(t *testing.T) {
	const password = "tcp(foo)/p?ss@word"
	normalized := normalizeDSN("u:" + password + "@tcp(host.example:3306)/db?parseTime=true")
	cfg, err := mysql.ParseDSN(normalized)
	if err != nil {
		t.Fatalf("ParseDSN(normalized) error = %v; DSN = %q", err, normalized)
	}
	if cfg.Passwd != password {
		t.Fatalf("normalized password = %q, want %q", cfg.Passwd, password)
	}
	if cfg.Addr != "host.example:3306" || cfg.DBName != "db" {
		t.Fatalf("normalized endpoint = %q/%q, want host.example:3306/db", cfg.Addr, cfg.DBName)
	}
}

func TestNormalizeURLDSNDoesNotInventCredentials(t *testing.T) {
	cfg, err := mysql.ParseDSN(normalizeDSN("mysql://host.example:3306/db?parseTime=true"))
	if err != nil {
		t.Fatalf("ParseDSN(normalized URL) error = %v", err)
	}
	if cfg.User != "" || cfg.Passwd != "" || cfg.Addr != "host.example:3306" || cfg.DBName != "db" {
		t.Fatalf("normalized URL config = user %q pass %q addr %q db %q", cfg.User, cfg.Passwd, cfg.Addr, cfg.DBName)
	}
}
