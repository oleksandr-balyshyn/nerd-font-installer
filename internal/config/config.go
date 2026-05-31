package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Release          string   `yaml:"release"`
	Destination      string   `yaml:"destination"`
	RefreshFontCache bool     `yaml:"refresh_font_cache"`
	Families         []string `yaml:"families"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.ApplyDefaults()
	return cfg, cfg.Validate()
}

func (c *Config) ApplyDefaults() {
	if c.Release == "" {
		c.Release = "latest"
	}
	if c.Destination == "" {
		c.Destination = "~/.local/share/fonts/NerdFonts"
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Release) == "" {
		return fmt.Errorf("release is required")
	}
	if strings.TrimSpace(c.Destination) == "" {
		return fmt.Errorf("destination is required")
	}
	if len(c.Families) == 0 {
		return fmt.Errorf("at least one font family is required")
	}
	seen := map[string]bool{}
	for _, family := range c.Families {
		family = strings.TrimSpace(family)
		if family == "" {
			return fmt.Errorf("font family names cannot be empty")
		}
		if seen[family] {
			return fmt.Errorf("duplicate font family %q", family)
		}
		seen[family] = true
	}
	return nil
}
