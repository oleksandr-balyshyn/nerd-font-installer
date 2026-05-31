package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/w0rxbend/nerd-font-installer/internal/config"
	"github.com/w0rxbend/nerd-font-installer/internal/fonts"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	configPath := flag.String("config", "fonts.yaml", "YAML config file with release, destination, refresh_font_cache, and families")
	dryRun := flag.Bool("dry-run", false, "print planned downloads without installing fonts")
	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "nerdfont-install %s (%s, %s)\n", version, commit, date)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if err := fonts.Install(ctx, fonts.Options{
		Release:          cfg.Release,
		Destination:      cfg.Destination,
		Families:         cfg.Families,
		RefreshFontCache: cfg.RefreshFontCache,
		DryRun:           *dryRun,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "install fonts: %v\n", err)
		os.Exit(1)
	}
}
