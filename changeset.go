package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bumpType string

const (
	patch bumpType = "patch"
	minor bumpType = "minor"
	major bumpType = "major"
)

type changeset struct {
	filePath string
	repoName string
	summary  string
	bump     bumpType
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
		return nil, fmt.Errorf("changeset missing opening delimiter (---)")
	}

	rest := content[3:]
	repoData, changesContent, found := strings.Cut(rest, "\n---")
	if !found {
		return nil, fmt.Errorf("changeset missing closing delimiter (---)")
	}

	repoData = strings.TrimSpace(repoData)
	changesContent = strings.TrimSpace(changesContent)

	parts := strings.SplitN(repoData, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid changeset format, expected 'repo-name: bump-type'")
	}

	repoName := strings.TrimSpace(parts[0])
	bt := bumpType(strings.TrimSpace(parts[1]))

	switch bt {
	case patch, minor, major:
		return &changeset{
			filePath: filePath,
			repoName: repoName,
			bump:     bt,
			summary:  changesContent,
		}, nil
	default:
		return nil, fmt.Errorf("invalid bump type %q, expected patch, minor, or major", bt)
	}
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
