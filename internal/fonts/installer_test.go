package fonts

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseURL(t *testing.T) {
	tests := []struct {
		release string
		family  string
		want    string
	}{
		{"latest", "JetBrainsMono", "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.zip"},
		{"v3.4.0", "Hack", "https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/Hack.zip"},
	}
	for _, tt := range tests {
		if got := ReleaseURL(tt.release, tt.family); got != tt.want {
			t.Fatalf("ReleaseURL(%q, %q) = %q, want %q", tt.release, tt.family, got, tt.want)
		}
	}
}

func TestExtractFontZipOnlyExtractsFonts(t *testing.T) {
	temp := t.TempDir()
	archivePath := filepath.Join(temp, "font.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, body := range map[string]string{
		"Font.ttf":        "font",
		"nested/Font.otf": "font",
		"README.md":       "docs",
	} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(temp, "out")
	if err := ExtractFontZip(archivePath, destination); err != nil {
		t.Fatalf("ExtractFontZip() error = %v", err)
	}
	for _, name := range []string{"Font.ttf", "Font.otf"} {
		if _, err := os.Stat(filepath.Join(destination, name)); err != nil {
			t.Fatalf("expected extracted font %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(destination, "README.md")); !os.IsNotExist(err) {
		t.Fatalf("README.md should not be extracted, stat err = %v", err)
	}
}

func TestInstallDryRun(t *testing.T) {
	var stdout bytes.Buffer
	err := Install(t.Context(), Options{
		Release:          "latest",
		Destination:      "/tmp/fonts",
		Families:         []string{"Hack"},
		RefreshFontCache: true,
		DryRun:           true,
		Stdout:           &stdout,
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Would install Hack") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Would refresh font cache") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
