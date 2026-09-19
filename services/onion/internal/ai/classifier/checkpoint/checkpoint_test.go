package checkpoint

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureDir = "../testdata/forms"

func TestTamperedCheckpoint(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"config.json", "model.safetensors"} {
		raw, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
			t.Fatalf("stage fixture: %v", err)
		}
	}
	wPath := filepath.Join(dir, "model.safetensors")
	raw, _ := os.ReadFile(wPath)
	raw[len(raw)-1] ^= 0xff
	if err := os.WriteFile(wPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("Load accepted a tampered checkpoint")
	}
}

func TestLoadFixture(t *testing.T) {
	f, err := Load(fixtureDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if f.Params != 14720 {
		t.Fatalf("Params = %d, want 14720 (reference parameter_count)", f.Params)
	}
	if f.Config.Encoder != "tinyx" || f.Config.Width != 16 {
		t.Fatalf("unexpected config: %+v", f.Config)
	}
}

func TestPyFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0.0, "0.0"},
		{0.1, "0.1"},
		{1e-05, "1e-05"},
		{0.0001, "0.0001"},
		{2.5e-05, "2.5e-05"},
		{1e15, "1000000000000000.0"},
		{1e16, "1e+16"},
		{-0.5, "-0.5"},
		{0.00012345, "0.00012345"},
	}
	for _, c := range cases {
		if got := pyFloat(c.in); got != c.want {
			t.Errorf("pyFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCanonical(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"é", `"\u00e9"`},
		{"🧑", `"\ud83e\uddd1"`},
		{"\x07", `"\u0007"`},
		{"a\"b\\c\nd", `"a\"b\\c\nd"`},
		{json.Number("32"), "32"},
		{json.Number("0.0"), "0.0"},
		{json.Number("1e-05"), "1e-05"},
		{[]any{json.Number("1"), "x"}, `[1,"x"]`},
		{map[string]any{"b": json.Number("2"), "a": json.Number("1")}, `{"a":1,"b":2}`},
	}
	for _, c := range cases {
		got, err := canonical(c.in)
		if err != nil {
			t.Fatalf("canonical(%v): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("canonical(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestTamperedConfigFailsClosed(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"config.json", "model.safetensors"} {
		raw, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
			t.Fatalf("stage fixture: %v", err)
		}
	}
	cPath := filepath.Join(dir, "config.json")
	raw, err := os.ReadFile(cPath)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), `"rank": 4`, `"rank": 5`, 1)
	if tampered == string(raw) {
		t.Fatal("fixture config does not contain the expected rank field")
	}
	if err := os.WriteFile(cPath, []byte(tampered), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("Load accepted a tampered config")
	}
}

func TestOversizedHeaderRejected(t *testing.T) {
	dir := t.TempDir()
	copyFixtureFile(t, dir, "config.json")
	var buf bytes.Buffer
	header := make([]byte, 8)
	binary.LittleEndian.PutUint64(header, maxHeaderBytes+1)
	buf.Write(header)
	buf.WriteString("{}")
	path := filepath.Join(dir, "model.safetensors")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load accepted an oversized header")
	}
}

func copyFixtureFile(t *testing.T, dir, name string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
		t.Fatalf("stage fixture: %v", err)
	}
}

func TestKindClassification(t *testing.T) {
	if got := Kind(nil); got != "" {
		t.Fatalf("Kind(nil) = %q", got)
	}
	if _, err := Load("does-not-exist"); Kind(err) != "not_found" {
		t.Fatalf("missing checkpoint kind = %q, want not_found", Kind(err))
	}
	dir := t.TempDir()
	copyFixtureFile(t, dir, "config.json")
	copyFixtureFile(t, dir, "model.safetensors")
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(raw), `"rank": 4`, `"rank": 5`, 1)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(tampered), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); Kind(err) != "signature_mismatch" {
		t.Fatalf("tampered checkpoint kind = %q, want signature_mismatch", Kind(err))
	}
}
