package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// bumpType represents a semantic version bump level.
type bumpType string

const (
	patch bumpType = "patch"
	minor bumpType = "minor"
	major bumpType = "major"
)

// changeset represents a parsed changeset file.
type changeset struct {
	filepath string   // absolute path to the .md file
	repoName string   // repo name from frontmatter
	bump     bumpType // patch, minor, or major
	summary  string   // the message body
}

// parseFile reads and parses a changeset markdown file.
// Expected format:
//
//	---
//	repo-name: patch
//	---
//
//	Summary text here
func parseFile(path string) (*changeset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read changeset %s: %w", path, err)
	}

	return parseChangeset(string(data), path)
}

// parseChangeset parses changeset content from a string.
func parseChangeset(content, filePath string) (*changeset, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("changeset missing opening frontmatter delimiter (---)")
	}

	// Remove the leading "---" and require the closing "---" on its own line.
	// This prevents a horizontal rule in the body from being misinterpreted.
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, fmt.Errorf("changeset missing closing frontmatter delimiter (---)")
	}

	frontmatter := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:])

	// Parse frontmatter: "repo-name: bump-type"
	parts := strings.SplitN(frontmatter, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid frontmatter format, expected 'name: bump-type'")
	}

	repoName := strings.TrimSpace(parts[0])
	bumpStr := strings.TrimSpace(parts[1])

	b, err := parseBumpType(bumpStr)
	if err != nil {
		return nil, err
	}

	return &changeset{
		filepath: filePath,
		repoName: repoName,
		bump:     b,
		summary:  body,
	}, nil
}

// changesetContent produces the markdown content for a changeset file.
func changesetContent(repoName string, bump bumpType, summary string) string {
	return fmt.Sprintf("---\n%s: %s\n---\n\n%s\n", repoName, bump, summary)
}

// listChangesets reads all .md files in the changes directory and parses them.
func listChangesets(changesDir string) ([]*changeset, error) {
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, fmt.Errorf("read changes directory: %w", err)
	}

	var result []*changeset
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(changesDir, entry.Name())
		cs, err := parseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		result = append(result, cs)
	}

	return result, nil
}

// highestBump returns the highest bump type among changesets.
// major > minor > patch
func highestBump(changes []*changeset) bumpType {
	highest := patch
	for _, cs := range changes {
		if bumpPriority(cs.bump) > bumpPriority(highest) {
			highest = cs.bump
		}
	}

	return highest
}

func parseBumpType(s string) (bumpType, error) {
	switch bumpType(s) {
	case patch, minor, major:
		return bumpType(s), nil
	default:
		return "", fmt.Errorf("invalid bump type %q, expected patch, minor, or major", s)
	}
}

func bumpPriority(b bumpType) int {
	switch b {
	case patch:
		return 1
	case minor:
		return 2
	case major:
		return 3
	default:
		return 0
	}
}
