package fonts

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	Release          string
	Destination      string
	Families         []string
	RefreshFontCache bool
	DryRun           bool
	Stdout           io.Writer
	Stderr           io.Writer
	HTTPClient       *http.Client
}

func Install(ctx context.Context, opts Options) error {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 10 * time.Minute}
	}
	if opts.Release == "" {
		opts.Release = "latest"
	}
	if len(opts.Families) == 0 {
		return fmt.Errorf("at least one Nerd Font family is required")
	}

	root, err := expandPath(opts.Destination)
	if err != nil {
		return err
	}
	if opts.DryRun {
		for _, family := range opts.Families {
			fmt.Fprintf(opts.Stdout, "Would install %s from %s into %s\n", family, ReleaseURL(opts.Release, family), filepath.Join(root, family))
		}
		if opts.RefreshFontCache {
			fmt.Fprintf(opts.Stdout, "Would refresh font cache for %s\n", root)
		}
		return nil
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for _, family := range opts.Families {
		if strings.TrimSpace(family) == "" {
			return fmt.Errorf("empty Nerd Font family")
		}
		if err := installFamily(ctx, opts.HTTPClient, opts.Release, family, root, opts.Stdout); err != nil {
			return err
		}
	}

	if opts.RefreshFontCache {
		return refreshFontCache(ctx, root, opts.Stdout, opts.Stderr)
	}
	return nil
}

func installFamily(ctx context.Context, client *http.Client, release, family, root string, stdout io.Writer) error {
	url := ReleaseURL(release, family)
	fmt.Fprintf(stdout, "Installing Nerd Font %s from %s\n", family, url)

	temp, err := os.CreateTemp("", "nerd-font-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download %s: %s", url, resp.Status)
	}
	if _, err := io.Copy(temp, resp.Body); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	destination := filepath.Join(root, family)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	return ExtractFontZip(temp.Name(), destination)
}

func ReleaseURL(release, family string) string {
	if release == "latest" {
		return fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/latest/download/%s.zip", family)
	}
	return fmt.Sprintf("https://github.com/ryanoasis/nerd-fonts/releases/download/%s/%s.zip", release, family)
}

func ExtractFontZip(path, destination string) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		if file.FileInfo().IsDir() || !isFontFile(file.Name) {
			continue
		}
		if err := extractZipFile(file, filepath.Join(destination, filepath.Base(file.Name))); err != nil {
			return err
		}
	}
	return nil
}

func isFontFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf", ".ttc", ".ttf":
		return true
	default:
		return false
	}
}

func extractZipFile(file *zip.File, destination string) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

func refreshFontCache(ctx context.Context, root string, stdout, stderr io.Writer) error {
	if _, err := exec.LookPath("fc-cache"); err != nil {
		fmt.Fprintln(stdout, "fc-cache is not available; skipping font cache refresh.")
		return nil
	}
	fmt.Fprintln(stdout, "Refreshing font cache...")
	cmd := exec.CommandContext(ctx, "fc-cache", "-f", root)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("destination is required")
	}
	if path == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}
