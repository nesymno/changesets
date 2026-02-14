package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ChangesetsDir = ".changesets"
	ConfigFile    = "config.json"
	ChangesDir    = "changes"
	ReadmeFile    = "README.md"
	GitkeepFile   = ".gitkeep"
)

// Config represents the .changesets/config.json file.
type Config struct {
	Version string `json:"version"`
}

// Paths holds resolved absolute paths for the changesets directory structure.
type Paths struct {
	Root       string // project root (where go.mod lives)
	Changesets string // .changesets/
	Config     string // .changesets/config.json
	Changes    string // .changesets/changes/
	Readme     string // .changesets/README.md
	Gitkeep    string // .changesets/changes/.gitkeep
}

// Root walks up from the current directory to find the project root
// (the directory containing go.mod).
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod in any parent directory")
		}

		dir = parent
	}
}

// ResolvePaths returns all changesets-related paths relative to the given root.
func ResolvePaths(root string) Paths {
	cs := filepath.Join(root, ChangesetsDir)
	return Paths{
		Root:       root,
		Changesets: cs,
		Config:     filepath.Join(cs, ConfigFile),
		Changes:    filepath.Join(cs, ChangesDir),
		Readme:     filepath.Join(cs, ReadmeFile),
		Gitkeep:    filepath.Join(cs, ChangesDir, GitkeepFile),
	}
}

// Load reads and parses the config.json file.
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the config back to disk with indentation.
func Save(configPath string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ModuleName reads go.mod and extracts the last segment of the module path.
// For example, "github.com/nesymno/changesets" returns "changesets".
func ModuleName(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)
			parts := strings.Split(modulePath, "/")
			return parts[len(parts)-1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	return "", fmt.Errorf("module directive not found in go.mod")
}
