# Nerd Font Installer

[![CI](https://github.com/w0rxbend/nerd-font-installer/actions/workflows/ci.yml/badge.svg)](https://github.com/w0rxbend/nerd-font-installer/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/w0rxbend/nerd-font-installer)](https://go.dev/)

`nerdfont-install` is a small Go CLI that installs selected
[Nerd Fonts](https://github.com/ryanoasis/nerd-fonts) release archives from a
YAML config file. It is designed for repeatable workstation bootstrap scripts:
the font list lives in config, the destination is explicit, and `--dry-run`
shows the exact downloads before touching the filesystem.

## Features

- Installs only `.ttf`, `.otf`, and `.ttc` files from Nerd Font zip archives.
- Supports `latest` or a pinned Nerd Fonts release tag such as `v3.4.0`.
- Installs each family into its own directory under the configured destination.
- Expands `~` in destination paths.
- Optionally refreshes the font cache with `fc-cache`.
- Provides `--dry-run` for bootstrap verification and CI smoke checks.

## Getting Started

Build the binary:

```bash
go build -o bin/nerdfont-install ./cmd/nerdfont-install
```

Create a config file:

```bash
cp config.example.yaml fonts.yaml
```

Run a dry-run first:

```bash
./bin/nerdfont-install --config fonts.yaml --dry-run
```

Install the configured fonts:

```bash
./bin/nerdfont-install --config config.example.yaml
```

Print build metadata:

```bash
./bin/nerdfont-install --version
```

## Configuration

The config file is YAML:

```yaml
release: latest
destination: ~/.local/share/fonts/NerdFonts
refresh_font_cache: true
families:
  - JetBrainsMono
  - Hack
```

### Fields

| Field | Required | Default | Description |
| --- | --- | --- | --- |
| `release` | No | `latest` | Nerd Fonts release to download from. Use `latest` or a tag such as `v3.4.0`. |
| `destination` | No | `~/.local/share/fonts/NerdFonts` | Root directory for installed font families. |
| `refresh_font_cache` | No | `false` | Runs `fc-cache -f <destination>` after installation when available. |
| `families` | Yes | none | Nerd Font archive names, for example `JetBrainsMono`, `Hack`, `FiraCode`, or `Meslo`. |

The family names must match the archive names published by the Nerd Fonts
release. For example, `JetBrainsMono` maps to:

```text
https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.zip
```

## Development

Run the local quality checks:

```bash
go mod tidy
go test ./...
go build -trimpath -o bin/nerdfont-install ./cmd/nerdfont-install
./bin/nerdfont-install --config config.example.yaml --dry-run
```

## CI and Releases

The repository includes two GitHub Actions workflows:

- `.github/workflows/ci.yml` runs `go mod tidy`, `go vet`, race-enabled tests,
  a build, and a dry-run smoke test on pull requests and pushes to `main` or
  `master`.
- `.github/workflows/release.yml` runs GoReleaser when a tag matching `v*` is
  pushed.

Create a GitHub Release by pushing a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser builds archives for Linux, macOS, and Windows on `amd64` and `arm64`,
uploads checksums, and injects version metadata into `nerdfont-install
--version`.

## Operational Notes

- Network access to `github.com` is required.
- `fc-cache` is optional. If it is missing, the installer skips cache refresh.
- Existing font files with the same names are overwritten.
- The installer writes only inside the configured destination directory.
