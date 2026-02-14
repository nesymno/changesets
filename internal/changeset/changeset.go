package changeset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BumpType represents a semantic version bump level.
type BumpType string

const (
	Patch BumpType = "patch"
	Minor BumpType = "minor"
	Major BumpType = "major"
)

// Changeset represents a parsed changeset file.
type Changeset struct {
	Filepath string   // absolute path to the .md file
	RepoName string   // repo name from frontmatter
	Bump     BumpType // patch, minor, or major
	Summary  string   // the message body
}

// ParseFile reads and parses a changeset markdown file.
// Expected format:
//
//	---
//	repo-name: patch
//	---
//
//	Summary text here
func ParseFile(path string) (*Changeset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read changeset %s: %w", path, err)
	}

	return Parse(string(data), path)
}

// Parse parses changeset content from a string.
func Parse(content string, filepath string) (*Changeset, error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("changeset missing opening frontmatter delimiter (---)")
	}

	// Remove the leading ---
	content = content[3:]

	// Find closing ---
	idx := strings.Index(content, "---")
	if idx < 0 {
		return nil, fmt.Errorf("changeset missing closing frontmatter delimiter (---)")
	}

	frontmatter := strings.TrimSpace(content[:idx])
	body := strings.TrimSpace(content[idx+3:])

	// Parse frontmatter: "repo-name: bump-type"
	parts := strings.SplitN(frontmatter, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid frontmatter format, expected 'name: bump-type'")
	}

	repoName := strings.TrimSpace(parts[0])
	bumpStr := strings.TrimSpace(parts[1])

	bump, err := parseBumpType(bumpStr)
	if err != nil {
		return nil, err
	}

	return &Changeset{
		Filepath: filepath,
		RepoName: repoName,
		Bump:     bump,
		Summary:  body,
	}, nil
}

// ChangesetContent produces the markdown content for a changeset file.
func ChangesetContent(repoName string, bump BumpType, summary string) string {
	return fmt.Sprintf("---\n%s: %s\n---\n\n%s\n", repoName, bump, summary)
}

// ListChangesets reads all .md files in the changes directory and parses them.
func ListChangesets(changesDir string) ([]*Changeset, error) {
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, fmt.Errorf("read changes directory: %w", err)
	}

	var changesets []*Changeset
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(changesDir, entry.Name())
		cs, err := ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		changesets = append(changesets, cs)
	}

	return changesets, nil
}

// HighestBump returns the highest bump type among changesets.
// major > minor > patch
func HighestBump(changesets []*Changeset) BumpType {
	highest := Patch
	for _, cs := range changesets {
		if bumpPriority(cs.Bump) > bumpPriority(highest) {
			highest = cs.Bump
		}
	}

	return highest
}

func parseBumpType(s string) (BumpType, error) {
	switch BumpType(s) {
	case Patch, Minor, Major:
		return BumpType(s), nil
	default:
		return "", fmt.Errorf("invalid bump type %q, expected patch, minor, or major", s)
	}
}

func bumpPriority(b BumpType) int {
	switch b {
	case Patch:
		return 1
	case Minor:
		return 2
	case Major:
		return 3
	default:
		return 0
	}
}
