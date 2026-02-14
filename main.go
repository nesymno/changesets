package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"

	"github.com/nesymno/changesets/internal/changeset"
	"github.com/nesymno/changesets/internal/config"
	"github.com/nesymno/changesets/internal/git"
	"github.com/nesymno/changesets/internal/words"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = cmdInit()
	case "add":
		err = cmdAdd()
	case "next":
		err = cmdNext()
	case "release":
		err = cmdRelease()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`changesets - Manage changelogs with semantic versioning

Usage:
  changesets <command>

Commands:
  init        Initialize .changesets directory
  add         Create a new changeset
  next        Calculate and print the next version
  release     Bump version, update CHANGELOG.md, and clean up changesets`)
}

// cmdInit creates the .changesets directory structure.
func cmdInit() error {
	root, err := config.Root()
	if err != nil {
		return err
	}

	paths := config.ResolvePaths(root)

	// Check if .changesets already exists
	if _, err := os.Stat(paths.Changesets); err == nil {
		fmt.Print(".changesets already exists. Recreate? (y/n): ")
		answer := readLine()
		if !strings.EqualFold(strings.TrimSpace(answer), "y") {
			fmt.Println("Aborted.")
			return nil
		}

		// Remove existing directory
		if err := os.RemoveAll(paths.Changesets); err != nil {
			return fmt.Errorf("failed to remove existing .changesets: %w", err)
		}
	}

	// Create directories
	if err := os.MkdirAll(paths.Changes, 0755); err != nil {
		return fmt.Errorf("failed to create changes directory: %w", err)
	}

	// Write config.json
	cfg := &config.Config{Version: "v0.0.0"}
	if err := config.Save(paths.Config, cfg); err != nil {
		return err
	}

	// Write README.md
	readme := `# Changesets

This directory is used by [go-changesets](https://github.com/nesymno/go-changesets) to manage versioning and changelogs.

## How to add a changeset

Run ` + "`go-changesets changeset`" + ` to create a new changeset file describing your change.

## How to release

Run ` + "`go-changesets release`" + ` to bump the version, update CHANGELOG.md, and clean up changeset files.
`
	if err := os.WriteFile(paths.Readme, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Write .gitkeep
	if err := os.WriteFile(paths.Gitkeep, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write .gitkeep: %w", err)
	}

	fmt.Println("Initialized .changesets directory.")
	return nil
}

// cmdAdd interactively creates a new changeset file.
func cmdAdd() error {
	root, err := config.Root()
	if err != nil {
		return err
	}

	paths := config.ResolvePaths(root)
	if err := ensureChangesetsExist(paths); err != nil {
		return err
	}

	repoName, err := config.ModuleName(root)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)

	// 1. Select bump type
	fmt.Println("What kind of change is this?")
	fmt.Println("  1) patch")
	fmt.Println("  2) minor")
	fmt.Println("  3) major")
	fmt.Print("Select [1/2/3]: ")

	var bump changeset.BumpType
	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	choice := strings.TrimSpace(scanner.Text())
	switch choice {
	case "1", "patch":
		bump = changeset.Patch
	case "2", "minor":
		bump = changeset.Minor
	case "3", "major":
		bump = changeset.Major
	default:
		return fmt.Errorf("invalid selection: %q", choice)
	}

	// 2. Enter summary
	fmt.Print("Summary: ")
	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	summary := strings.TrimSpace(scanner.Text())
	if summary == "" {
		return fmt.Errorf("summary cannot be empty")
	}

	// 3. Preview and confirm
	content := changeset.ChangesetContent(repoName, bump, summary)
	fmt.Println()
	fmt.Println("--- Preview ---")
	fmt.Println()
	fmt.Print(content)
	fmt.Println()
	fmt.Println("--- End Preview ---")
	fmt.Println()
	fmt.Print("Confirm? (y/n): ")

	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	confirm := strings.TrimSpace(scanner.Text())
	if !strings.EqualFold(confirm, "y") {
		fmt.Println("Aborted.")
		return nil
	}

	// 4. Generate slug and write file
	slug, err := words.Generate(paths.Changes)
	if err != nil {
		return err
	}

	filename := words.SlugToFilename(slug)
	filePath := filepath.Join(paths.Changes, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write changeset file: %w", err)
	}

	fmt.Printf("Created changeset: .changesets/changes/%s\n", filename)
	return nil
}

// cmdNext calculates and prints the next version.
func cmdNext() error {
	root, err := config.Root()
	if err != nil {
		return err
	}

	paths := config.ResolvePaths(root)
	if err := ensureChangesetsExist(paths); err != nil {
		return err
	}

	nextVer, _, err := calculateNextVersion(paths)
	if err != nil {
		return err
	}

	fmt.Println(nextVer)
	return nil
}

// cmdRelease bumps the version, updates CHANGELOG.md, and cleans up changesets.
func cmdRelease() error {
	root, err := config.Root()
	if err != nil {
		return err
	}

	paths := config.ResolvePaths(root)
	if err := ensureChangesetsExist(paths); err != nil {
		return err
	}

	nextVerStr, changesets, err := calculateNextVersion(paths)
	if err != nil {
		return err
	}

	if len(changesets) == 0 {
		return fmt.Errorf("no changesets found, nothing to release")
	}

	// Build changelog section
	changelogSection := buildChangelogSection(nextVerStr, changesets)

	// Update CHANGELOG.md
	changelogPath := filepath.Join(root, "CHANGELOG.md")
	if err := prependChangelog(changelogPath, changelogSection); err != nil {
		return err
	}

	// Update config.json
	cfg, err := config.Load(paths.Config)
	if err != nil {
		return err
	}
	cfg.Version = nextVerStr
	if err := config.Save(paths.Config, cfg); err != nil {
		return err
	}

	// Clean up changeset files
	if err := cleanupChanges(paths.Changes); err != nil {
		return err
	}

	fmt.Println(nextVerStr)
	return nil
}

// calculateNextVersion reads the current version and all changesets, then computes the next version.
func calculateNextVersion(paths config.Paths) (string, []*changeset.Changeset, error) {
	cfg, err := config.Load(paths.Config)
	if err != nil {
		return "", nil, err
	}

	changesets, err := changeset.ListChangesets(paths.Changes)
	if err != nil {
		return "", nil, err
	}

	if len(changesets) == 0 {
		fmt.Fprintf(os.Stderr, "no changesets found\n")
		return cfg.Version, nil, nil
	}

	// Parse current version
	currentVersion := strings.TrimPrefix(cfg.Version, "v")
	ver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse current version %q: %w", cfg.Version, err)
	}

	// Determine highest bump
	bump := changeset.HighestBump(changesets)

	// Apply bump
	var next semver.Version
	switch bump {
	case changeset.Major:
		next = ver.IncMajor()
	case changeset.Minor:
		next = ver.IncMinor()
	case changeset.Patch:
		next = ver.IncPatch()
	}

	nextVerStr := "v" + next.String()
	return nextVerStr, changesets, nil
}

// buildChangelogSection produces the markdown section for a release.
func buildChangelogSection(version string, changesets []*changeset.Changeset) string {
	var sb strings.Builder

	date := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("## %s - %s\n", version, date))

	// Group by bump type
	groups := map[changeset.BumpType][]*changeset.Changeset{
		changeset.Major: {},
		changeset.Minor: {},
		changeset.Patch: {},
	}
	for _, cs := range changesets {
		groups[cs.Bump] = append(groups[cs.Bump], cs)
	}

	// Write each group in order: major, minor, patch
	writeGroup := func(title string, items []*changeset.Changeset) {
		if len(items) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("\n### %s\n\n", title))
		for _, cs := range items {
			sha := git.GetFileCommitSHA(cs.Filepath)
			if sha != "" {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", sha, cs.Summary))
			} else {
				sb.WriteString(fmt.Sprintf("- %s\n", cs.Summary))
			}
		}
	}

	writeGroup("Major Changes", groups[changeset.Major])
	writeGroup("Minor Changes", groups[changeset.Minor])
	writeGroup("Patch Changes", groups[changeset.Patch])

	return sb.String()
}

// prependChangelog prepends a new section to CHANGELOG.md.
func prependChangelog(path string, section string) error {
	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	var content string
	if existing == "" {
		content = "# Changelog\n\n" + section
	} else {
		// Insert after the first line (# Changelog header) if it exists
		if strings.HasPrefix(existing, "# ") {
			idx := strings.Index(existing, "\n")
			if idx >= 0 {
				header := existing[:idx+1]
				rest := existing[idx+1:]
				// Trim leading newlines from rest
				rest = strings.TrimLeft(rest, "\n")
				content = header + "\n" + section + "\n" + rest
			} else {
				content = existing + "\n\n" + section
			}
		} else {
			content = section + "\n" + existing
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
	}

	return nil
}

// cleanupChanges removes all .md files from the changes directory, keeping .gitkeep.
func cleanupChanges(changesDir string) error {
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return fmt.Errorf("failed to read changes directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(changesDir, entry.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// ensureChangesetsExist checks that the .changesets directory exists.
func ensureChangesetsExist(paths config.Paths) error {
	if _, err := os.Stat(paths.Changesets); os.IsNotExist(err) {
		return fmt.Errorf(".changesets directory not found. Run 'go-changesets init' first")
	}
	return nil
}

// readLine reads a single line from stdin.
func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
