package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	changesetsDir = ".changesets"
	configFile    = "config.json"
	changesDir    = "changes"
	readmeFile    = "README.md"
	gitkeepFile   = ".gitkeep"
)

// config represents the .changesets/config.json file.
type config struct {
	Version string `json:"version"`
}

// paths holds resolved absolute paths for the changesets directory structure.
type paths struct {
	root       string // project root (where go.mod lives)
	changesets string // .changesets/
	config     string // .changesets/config.json
	changes    string // .changesets/changes/
	readme     string // .changesets/README.md
	gitkeep    string // .changesets/changes/.gitkeep
}

// findRoot walks up from the current directory to find the project root
// (the directory containing go.mod).
func findRoot() (string, error) {
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

// newPaths returns all changesets-related paths relative to the given root.
func newPaths(root string) paths {
	cs := filepath.Join(root, changesetsDir)
	return paths{
		root:       root,
		changesets: cs,
		config:     filepath.Join(cs, configFile),
		changes:    filepath.Join(cs, changesDir),
		readme:     filepath.Join(cs, readmeFile),
		gitkeep:    filepath.Join(cs, changesDir, gitkeepFile),
	}
}

// loadConfig reads and parses the config.json file.
func loadConfig(configPath string) (*config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// saveConfig writes the config back to disk with indentation.
func saveConfig(configPath string, cfg *config) error {
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

// moduleName reads go.mod and extracts the last segment of the module path.
// For example, "github.com/nesymno/changesets" returns "changesets".
func moduleName(root string) (string, error) {
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
