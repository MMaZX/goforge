// Package config loads GoForge's project configuration from goforge.yaml,
// a .env file and the process environment.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/MMaZX/goforge/internal/providers"
)

// Config is the parsed content of goforge.yaml, with ${VAR} references
// expanded against the environment (after .env has been loaded into it).
type Config struct {
	Database struct {
		Driver string `yaml:"driver"`
		URL    string `yaml:"url"`
	} `yaml:"database"`
	Migrations struct {
		Path string `yaml:"path"`
	} `yaml:"migrations"`
}

// Load reads goforge.yaml from path, first loading a sibling .env file (if
// present) into the process environment so ${VAR} references in the YAML
// resolve. It does not overwrite variables already set in the environment.
func Load(path string) (*Config, error) {
	if err := loadDotEnv(DotEnvPath(path)); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading %s: %w", path, err)
	}

	var raw Config
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(data))), &raw); err != nil {
		return nil, fmt.Errorf("config: parsing %s: %w", path, err)
	}

	desc, err := providers.Resolve(raw.Database.Driver)
	if err != nil {
		return nil, fmt.Errorf("config: database.driver: %w", err)
	}
	raw.Database.Driver = desc.Driver

	if raw.Database.URL == "" {
		return nil, fmt.Errorf("config: database.url is required")
	}
	if raw.Migrations.Path == "" {
		raw.Migrations.Path = "./migrations"
	}

	return &raw, nil
}

// DotEnvPath returns the .env path Load() reads relative to configPath:
// the same directory, named ".env".
func DotEnvPath(configPath string) string {
	dir := "."
	if idx := strings.LastIndexByte(configPath, '/'); idx >= 0 {
		dir = configPath[:idx]
	}
	return dir + "/.env"
}

// loadDotEnv parses simple KEY=VALUE lines (# comments and blank lines
// ignored) and sets them in the process environment, without overwriting
// variables that are already set.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("config: reading .env: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
