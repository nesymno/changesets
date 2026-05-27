package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
)

var version = "dev"

func resolvedVersion() string {
	if version != "dev" {
		return version
	}

	p, err := resolvePaths()
	if err != nil {
		return version
	}
	cfg, err := loadConfig(p.config)
	if err != nil {
		return version
	}
	return cfg.Version
}

func main() {
	os.Exit(run(os.Args, os.Stdin))
}

func run(args []string, stdin io.Reader) int {
	if len(args) < 2 {
		printUsage()
		return 0
	}

	switch args[1] {
	case "help", "--help", "-h":
		printUsage()
		return 0
	case "version", "--version", "-v":
		fmt.Println(resolvedVersion())
		return 0
	}

	p, err := resolvePaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\nAre you inside a Go project?\n", err)
		return 1
	}

	scanner := bufio.NewScanner(stdin)

	switch args[1] {
	case "init":
		err = cmdInit(p, scanner)
	case "add":
		err = cmdAdd(p, scanner)
	case "next":
		err = cmdNext(p)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[1])
		printUsage()
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Print(`changesets - Manage changelogs with semantic versioning

Usage:
  changesets <command>

Commands:
  init        Initialize .changesets directory
  add         Create a new changeset
  next        Calculate and print the next version
  version     Print the CLI version
`)
}

func resolvePaths() (paths, error) {
	root, err := findRoot()
	if err != nil {
		return paths{}, err
	}

	return newPaths(root), nil
}

func cmdInit(p paths, scanner *bufio.Scanner) error {
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

		if err := os.RemoveAll(p.changesets); err != nil {
			return fmt.Errorf("emove existing .changesets: %w", err)
		}
	}

	if err := os.MkdirAll(p.changes, 0755); err != nil {
		return fmt.Errorf("reate changes directory: %w", err)
	}

	cfg := &config{Version: "v0.0.1"}
	if err := saveConfig(p.config, cfg); err != nil {
		return err
	}

	readme := `# Changesets

This directory is used by [changesets](https://github.com/nesymno/changesets) to manage versioning and changelogs.

## How to add a changeset

Run ` + "`changesets add`" + ` to create a new changeset file describing your change.

## How to release

Run ` + "`changesets release`" + ` to bump the version, update [CHANGELOG.md](https://github.com/nesymno/changesets/CHANGELOG.md), and clean up changeset files.
`
	if err := os.WriteFile(p.readme, []byte(readme), 0644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	if err := os.WriteFile(p.gitkeep, []byte(""), 0644); err != nil {
		return fmt.Errorf("write .gitkeep: %w", err)
	}

	fmt.Println("Initialized .changesets directory.")
	return nil
}

// cmdAdd interactively creates a new changeset file.
func cmdAdd(p paths, scanner *bufio.Scanner) error {
	if err := ensureChangesetsExist(p); err != nil {
		return err
	}

	repoName, err := moduleName(p.root)
	if err != nil {
		return err
	}

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

	fmt.Print("Summary: ")
	if !scanner.Scan() {
		return fmt.Errorf("no input received")
	}
	summary := strings.TrimSpace(scanner.Text())
	if summary == "" {
		return fmt.Errorf("summary cannot be empty")
	}

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

	slug, err := generateSlug(p.changes)
	if err != nil {
		return err
	}

	filename := slugToFilename(slug)
	filePath := filepath.Join(p.changes, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write changeset file: %w", err)
	}

	fmt.Printf("Created changeset: .changesets/changes/%s\n", filename)
	return nil
}

// cmdNext calculates and prints the next version.
func cmdNext(p paths) error {
	if err := ensureChangesetsExist(p); err != nil {
		return err
	}

	nextVer, _, _, err := calculateNextVersion(p)
	if err != nil {
		return err
	}

	fmt.Println(nextVer)
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
		return "", nil, nil, fmt.Errorf("parse current version %q: %w", cfg.Version, err)
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
	fmt.Fprintf(&sb, "## %s - %s\n", ver, date)

	groups := map[bumpType][]*changeset{
		major: {},
		minor: {},
		patch: {},
	}
	for _, cs := range changes {
		groups[cs.bump] = append(groups[cs.bump], cs)
	}

	writeGroup := func(title string, items []*changeset) {
		if len(items) == 0 {
			return
		}
		fmt.Fprintf(&sb, "\n### %s\n\n", title)
		for _, cs := range items {
			sha, _ := getFileCommitSHA(cs.filePath)
			if sha != "" {
				fmt.Fprintf(&sb, "- %s: %s\n", sha, cs.summary)
			} else {
				fmt.Fprintf(&sb, "- %s\n", cs.summary)
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
		return fmt.Errorf("write CHANGELOG.md: %w", err)
	}

	return nil
}

// cleanupChanges removes all .md files from the changes directory, keeping .gitkeep.
func cleanupChanges(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read changes directory: %w", err)
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
			return fmt.Errorf("remove %s: %w", entry.Name(), err)
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
