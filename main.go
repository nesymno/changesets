package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
)

const startVersion = "v0.0.0"

func resolvedVersion() string {
	p, err := resolvePaths()
	if err != nil {
		return startVersion
	}
	cfg, err := loadConfig(p.config)
	if err != nil {
		return startVersion
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

	scanner := newLineScanner(stdin)

	switch args[1] {
	case "init":
		err = cmdInit(p, scanner)
	case "add":
		err = cmdAdd(p, scanner)
	case "next":
		err = cmdNext(p)
	case "release":
		err = cmdRelease(p)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[1])
		printUsage()
		return 1
	}
	if err != nil {
		if errors.Is(err, ErrNoChangesets) {
			fmt.Fprint(os.Stderr, err)
			return 2
		}

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
  release     Update CHANGELOG, bump version, and clean up changesets
  version     Show the CLI version
`)
}

func resolvePaths() (config, error) {
	root, err := findRoot()
	if err != nil {
		return config{}, err
	}

	return newConfig(root), nil
}

// lineScanner is the input interface used by interactive commands.
// *bufio.Scanner satisfies it directly, so tests need no changes.
type lineScanner interface {
	Scan() bool
	Text() string
}

type readerScanner struct {
	r    *bufio.Reader
	line string
}

func (s *readerScanner) Scan() bool {
	line, err := s.r.ReadString('\n')
	s.line = strings.TrimRight(line, "\r\n")
	return len(line) > 0 && (err == nil || err == io.EOF)
}

func (s *readerScanner) Text() string { return s.line }

func newLineScanner(stdin io.Reader) lineScanner {
	return &readerScanner{r: bufio.NewReader(stdin)}
}

func cmdInit(p config, scanner lineScanner) error {
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

	cfg := &Config{Version: startVersion}
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
func cmdAdd(p config, scanner lineScanner) error {
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

	var bump bump
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

var ErrNoChangesets = fmt.Errorf("no changesets found")

// cmdRelease updates CHANGELOG.md, bumps the version in config, and removes processed changesets.
func cmdRelease(p config) error {
	if err := ensureChangesetsExist(p); err != nil {
		return err
	}

	ver, err := buildNext(p)
	if err != nil {
		return err
	}

	if len(ver.changes) == 0 {
		return ErrNoChangesets
	}

	section := buildChangelogSection(ver.version, ver.changes)

	if err := prependChangelog(p.changelog, section); err != nil {
		return err
	}

	if err := saveConfig(p.config, ver.config); err != nil {
		return err
	}

	if err := cleanupChanges(p.changes); err != nil {
		return err
	}

	fmt.Println(ver.version)
	return nil
}

// cmdNext calculates and prints the next version.
func cmdNext(p config) error {
	if err := ensureChangesetsExist(p); err != nil {
		return err
	}

	ver, err := buildNext(p)
	if err != nil {
		return err
	}

	fmt.Println(ver.version)
	return nil
}

func currentVersion(currentVersion string) (*semver.Version, error) {
	ver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("parse current version %q: %w", currentVersion, err)
	}
	return ver, nil
}

type version struct {
	version string
	changes []*changeset
	config  *Config
}

// buildNext prepares the version struct by loading config and changesets.
func buildNext(p config) (version, error) {
	var ver version

	cfg, err := loadConfig(p.config)
	if err != nil {
		return ver, err
	}
	ver.config = cfg

	changes, err := listChangesets(p.changes)
	if err != nil {
		return ver, err
	}
	if len(changes) == 0 {
		ver.version = cfg.Version
		return ver, nil
	}
	ver.changes = changes

	cur, err := currentVersion(cfg.Version)
	if err != nil {
		return ver, err
	}

	bump := highestBump(changes)

	var next semver.Version
	switch bump {
	case major:
		next = cur.IncMajor()
	case minor:
		next = cur.IncMinor()
	case patch:
		next = cur.IncPatch()
	}
	ver.version = "v" + next.String()
	ver.config.Version = ver.version

	return ver, nil
}

// buildChangelogSection produces the markdown section for a release.
func buildChangelogSection(ver string, changes []*changeset) string {
	var sb strings.Builder

	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(&sb, "## %s - %s\n", ver, date)

	groups := map[bump][]*changeset{
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
func ensureChangesetsExist(p config) error {
	if _, err := os.Stat(p.changesets); os.IsNotExist(err) {
		return fmt.Errorf(".changesets directory not found. Run 'changesets init' first")
	}
	return nil
}
