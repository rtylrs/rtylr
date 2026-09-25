package cdn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckMediaEngine(t *testing.T) {
	t.Run("unconfigured", func(t *testing.T) {
		c := &CDN{mediaEngine: ""}
		if err := c.CheckMediaEngine(); err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("missing", func(t *testing.T) {
		c := &CDN{mediaEngine: filepath.Join(t.TempDir(), "nope")}
		if err := c.CheckMediaEngine(); err == nil {
			t.Error("expected error for missing binary")
		}
	})

	t.Run("directory", func(t *testing.T) {
		c := &CDN{mediaEngine: t.TempDir()}
		if err := c.CheckMediaEngine(); err == nil {
			t.Error("expected error for a directory")
		}
	})

	t.Run("non-executable", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "plain")
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		c := &CDN{mediaEngine: f}
		if err := c.CheckMediaEngine(); err == nil {
			t.Error("expected error for non-executable file")
		}
	})

	t.Run("executable", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "bin")
		if err := os.WriteFile(f, []byte("#!/bin/sh\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(f, 0o700); err != nil {
			t.Fatal(err)
		}
		c := &CDN{mediaEngine: f}
		if err := c.CheckMediaEngine(); err != nil {
			t.Errorf("expected no error for executable, got %v", err)
		}
	})
}
