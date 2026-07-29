package dbclient

import "testing"

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
