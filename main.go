package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printUsage()
		return
	case "version", "--version", "-v":
		fmt.Println(version)
		return
	}

	p, err := resolvePaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	switch os.Args[1] {
	case "init":
		err = cmdInit(p, scanner)
	case "add":
		err = cmdAdd(p, scanner)
	case "next":
		err = cmdNext(p)
	case "release":
		err = cmdRelease(p)
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
  release     Bump version, update CHANGELOG.md, and clean up changesets
  version     Print the CLI version`)
}

func resolvePaths() (paths, error) {
	root, err := findRoot()
	if err != nil {
		return paths{}, err
	}

	return newPaths(root), nil
}

// cmdInit creates the .changesets directory structure.
func cmdInit(p paths, scanner *bufio.Scanner) error {
	// Check if .changesets already exists
	if _, err := os.Stat(p.changesets); err == nil {
		fmt.Print(".changesets already exists. Recreate? (y/n): ")
		if !scanner.Scan() {
			return fmt.Errorf("no input received")
		}
		answer := strings.TrimSpace(scanner.Text())
		if !strings.EqualFold(answer, "y") {
			fmt.Println("Aborted.")
			return nil
		}

		// Remove existing directory
		if err := os.RemoveAll(p.changesets); err != nil {
			return fmt.Errorf("failed to remove existing .changesets: %w", err)
		}
	}

	// Create directories
	if err := os.MkdirAll(p.changes, 0755); err != nil {
		return fmt.Errorf("failed to create changes directory: %w", err)
	}

	// Write config.json
	cfg := &config{Version: "v0.0.0"}
	if err := saveConfig(p.config, cfg); err != nil {
		return err
	}

	// Write README.md
	readme := `# Changesets

This directory is used by [go-changesets](https://github.com/nesymno/go-changesets) to manage versioning and changelogs.

## How to add a changeset

Run ` + "`changesets add`" + ` to create a new changeset file describing your change.

## How to release

Run ` + "`changesets release`" + ` to bump the version, update CHANGELOG.md, and clean up changeset files.
`
	if err := os.WriteFile(p.readme, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Write .gitkeep
	if err := os.WriteFile(p.gitkeep, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write .gitkeep: %w", err)
	}

	fmt.Println("Initialized .changesets directory.")
	return nil
}

// cmdAdd interactively creates a new changeset file.
func cmdAdd(p paths, scanner *bufio.Scanner) error {
	repoName, err := moduleName(p.root)
	if err != nil {
		return err
	}

	// 1. Select bump type
	fmt.Println("What kind of change is this?")
	fmt.Println("  1) patch")
	fmt.Println("  2) minor")
	fmt.Println("  3) major")
	fmt.Print("Select [1/2/3]: ")

	var bump bumpType
	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	choice := strings.TrimSpace(scanner.Text())
	switch choice {
	case "1", "patch":
		bump = patch
	case "2", "minor":
		bump = minor
	case "3", "major":
		bump = major
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
	content := changesetContent(repoName, bump, summary)
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
	slug, err := generateSlug(p.changes)
	if err != nil {
		return err
	}

	filename := slugToFilename(slug)
	filePath := filepath.Join(p.changes, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write changeset file: %w", err)
	}

	fmt.Printf("Created changeset: .changesets/changes/%s\n", filename)
	return nil
}

// cmdNext calculates and prints the next version.
func cmdNext(p paths) error {
	nextVer, _, _, err := calculateNextVersion(p)
	if err != nil {
		return err
	}

	fmt.Println(nextVer)
	return nil
}

// cmdRelease bumps the version, updates CHANGELOG.md, and cleans up changesets.
func cmdRelease(p paths) error {
	if err := ensureChangesetsExist(p); err != nil {
		return err
	}

	nextVerStr, changes, cfg, err := calculateNextVersion(p)
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return fmt.Errorf("no changesets found, nothing to release")
	}

	// Build changelog section
	changelogSection := buildChangelogSection(nextVerStr, changes)

	// Update CHANGELOG.md
	changelogPath := filepath.Join(p.root, "CHANGELOG.md")
	if err := prependChangelog(changelogPath, changelogSection); err != nil {
		return err
	}

	// Update config.json
	cfg.Version = nextVerStr
	if err := saveConfig(p.config, cfg); err != nil {
		return err
	}

	// Clean up changeset files
	if err := cleanupChanges(p.changes); err != nil {
		return err
	}

	fmt.Println(nextVerStr)
	return nil
}

// calculateNextVersion reads the current version and all changesets, then computes the next version.
func calculateNextVersion(p paths) (string, []*changeset, *config, error) {
	cfg, err := loadConfig(p.config)
	if err != nil {
		return "", nil, nil, err
	}

	changes, err := listChangesets(p.changes)
	if err != nil {
		return "", nil, nil, err
	}

	if len(changes) == 0 {
		return cfg.Version, nil, cfg, nil
	}

	// Parse current version
	currentVersion := strings.TrimPrefix(cfg.Version, "v")
	ver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to parse current version %q: %w", cfg.Version, err)
	}

	// Determine highest bump
	bump := highestBump(changes)

	// Apply bump
	var next semver.Version
	switch bump {
	case major:
		next = ver.IncMajor()
	case minor:
		next = ver.IncMinor()
	case patch:
		next = ver.IncPatch()
	}

	nextVerStr := "v" + next.String()
	return nextVerStr, changes, cfg, nil
}

// buildChangelogSection produces the markdown section for a release.
func buildChangelogSection(ver string, changes []*changeset) string {
	var sb strings.Builder

	date := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("## %s - %s\n", ver, date))

	// Group by bump type
	groups := map[bumpType][]*changeset{
		major: {},
		minor: {},
		patch: {},
	}
	for _, cs := range changes {
		groups[cs.bump] = append(groups[cs.bump], cs)
	}

	// Write each group in order: major, minor, patch
	writeGroup := func(title string, items []*changeset) {
		if len(items) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("\n### %s\n\n", title))
		for _, cs := range items {
			sha, _ := getFileCommitSHA(cs.filepath)
			if sha != "" {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", sha, cs.summary))
			} else {
				sb.WriteString(fmt.Sprintf("- %s\n", cs.summary))
			}
		}
	}

	writeGroup("Major Changes", groups[major])
	writeGroup("Minor Changes", groups[minor])
	writeGroup("Patch Changes", groups[patch])

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
func cleanupChanges(dir string) error {
	entries, err := os.ReadDir(dir)
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
		path := filepath.Join(dir, entry.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// ensureChangesetsExist checks that the .changesets directory exists.
func ensureChangesetsExist(p paths) error {
	if _, err := os.Stat(p.changesets); os.IsNotExist(err) {
		return fmt.Errorf(".changesets directory not found. Run 'changesets init' first")
	}
	return nil
}
