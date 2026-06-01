package fonts

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseURL(t *testing.T) {
	tests := []struct {
		name    string
		release string
		family  string
		want    string
	}{
		{
			name:    "latest",
			release: "latest",
			family:  "JetBrainsMono",
			want:    "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.zip",
		},
		{
			name:    "tagged release",
			release: "v3.4.0",
			family:  "Hack",
			want:    "https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/Hack.zip",
		},
		{
			name:    "escapes path segments",
			release: "release candidate",
			family:  "Symbols Nerd Font",
			want:    "https://github.com/ryanoasis/nerd-fonts/releases/download/release%20candidate/Symbols%20Nerd%20Font.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReleaseURL(tt.release, tt.family); got != tt.want {
				t.Fatalf("ReleaseURL(%q, %q) = %q, want %q", tt.release, tt.family, got, tt.want)
			}
		})
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

func TestExtractFontZipRejectsInvalidZip(t *testing.T) {
	temp := t.TempDir()
	archivePath := filepath.Join(temp, "font.zip")
	if err := os.WriteFile(archivePath, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := ExtractFontZip(archivePath, filepath.Join(temp, "out"))
	if err == nil {
		t.Fatal("ExtractFontZip() error = nil, want invalid zip error")
	}
	if !strings.Contains(err.Error(), "open font zip") {
		t.Fatalf("ExtractFontZip() error = %v, want open font zip context", err)
	}
}

func TestExtractFontZipRejectsArchiveWithoutFonts(t *testing.T) {
	temp := t.TempDir()
	archivePath := filepath.Join(temp, "font.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("README.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("docs")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	err = ExtractFontZip(archivePath, filepath.Join(temp, "out"))
	if err == nil {
		t.Fatal("ExtractFontZip() error = nil, want empty archive error")
	}
	if !strings.Contains(err.Error(), "no font files found") {
		t.Fatalf("ExtractFontZip() error = %v, want no font files found", err)
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

func TestInstallDryRunRejectsInvalidFamily(t *testing.T) {
	var stdout bytes.Buffer
	err := Install(t.Context(), Options{
		Release:     "latest",
		Destination: "/tmp/fonts",
		Families:    []string{"../Hack"},
		DryRun:      true,
		Stdout:      &stdout,
	})
	if err == nil {
		t.Fatal("Install() error = nil, want invalid family error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestInstallProgressWritesToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(bytes.NewReader(fontZip(t))),
			}, nil
		}),
	}

	err := Install(t.Context(), Options{
		Release:     "latest",
		Destination: filepath.Join(t.TempDir(), "fonts"),
		Families:    []string{"Hack"},
		Stdout:      &stdout,
		Stderr:      &stderr,
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if strings.Contains(stdout.String(), "Installing Nerd Font") {
		t.Fatalf("stdout = %q, want no progress message", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Installing Nerd Font Hack") {
		t.Fatalf("stderr = %q, want progress message", stderr.String())
	}
}

func TestInstallReplacesFamilyDirectoryAfterSuccessfulExtraction(t *testing.T) {
	var stderr bytes.Buffer
	destination := filepath.Join(t.TempDir(), "fonts")
	existingFamily := filepath.Join(destination, "Hack")
	if err := os.MkdirAll(existingFamily, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existingFamily, "old.ttf"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(bytes.NewReader(fontZip(t))),
			}, nil
		}),
	}

	err := Install(t.Context(), Options{
		Release:     "latest",
		Destination: destination,
		Families:    []string{"Hack"},
		Stderr:      &stderr,
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(existingFamily, "Hack.ttf")); err != nil {
		t.Fatalf("expected new font: %v", err)
	}
	if _, err := os.Stat(filepath.Join(existingFamily, "old.ttf")); !os.IsNotExist(err) {
		t.Fatalf("old font should be removed, stat err = %v", err)
	}
}

func TestInstallKeepsExistingFamilyDirectoryOnExtractionFailure(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "fonts")
	existingFamily := filepath.Join(destination, "Hack")
	if err := os.MkdirAll(existingFamily, 0o755); err != nil {
		t.Fatal(err)
	}
	existingFont := filepath.Join(existingFamily, "old.ttf")
	if err := os.WriteFile(existingFont, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(bytes.NewReader(noFontZip(t))),
			}, nil
		}),
	}

	err := Install(t.Context(), Options{
		Release:     "latest",
		Destination: destination,
		Families:    []string{"Hack"},
		HTTPClient:  client,
	})
	if err == nil {
		t.Fatal("Install() error = nil, want extraction error")
	}
	if data, err := os.ReadFile(existingFont); err != nil || string(data) != "old" {
		t.Fatalf("existing font = %q, %v; want old font preserved", data, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func fontZip(t *testing.T) []byte {
	t.Helper()

	var body bytes.Buffer
	writer := zip.NewWriter(&body)
	entry, err := writer.Create("Hack.ttf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("font")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

func noFontZip(t *testing.T) []byte {
	t.Helper()

	var body bytes.Buffer
	writer := zip.NewWriter(&body)
	entry, err := writer.Create("README.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("docs")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}
