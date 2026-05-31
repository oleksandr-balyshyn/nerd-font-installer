package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fonts.yaml")
	if err := os.WriteFile(path, []byte("families: [JetBrainsMono]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Release != "latest" {
		t.Fatalf("Release = %q", cfg.Release)
	}
	if cfg.Destination != "~/.local/share/fonts/NerdFonts" {
		t.Fatalf("Destination = %q", cfg.Destination)
	}
}

func TestValidateRejectsDuplicateFamilies(t *testing.T) {
	cfg := Config{Release: "latest", Destination: "/tmp/fonts", Families: []string{"Hack", "Hack"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want duplicate error")
	}
}
