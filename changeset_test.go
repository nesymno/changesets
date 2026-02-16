package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	content := "---\nmy-repo: minor\n---\n\nAdded something cool"

	cs, err := parseChangeset(content, "test.md")
	if err != nil {
		t.Fatalf("parseChangeset failed: %v", err)
	}

	if cs.repoName != "my-repo" {
		t.Errorf("expected repo name my-repo, got %s", cs.repoName)
	}
	if cs.bump != minor {
		t.Errorf("expected bump minor, got %s", cs.bump)
	}
	if cs.summary != "Added something cool" {
		t.Errorf("expected summary 'Added something cool', got %q", cs.summary)
	}
}

func TestParseAllBumpTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected bumpType
	}{
		{"---\nrepo: patch\n---\n\nfix", patch},
		{"---\nrepo: minor\n---\n\nfeat", minor},
		{"---\nrepo: major\n---\n\nbreaking", major},
	}

	for _, tt := range tests {
		cs, err := parseChangeset(tt.input, "test.md")
		if err != nil {
			t.Fatalf("parseChangeset failed for %s: %v", tt.expected, err)
		}
		if cs.bump != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, cs.bump)
		}
	}
}

func TestParseMissingFrontmatter(t *testing.T) {
	_, err := parseChangeset("no frontmatter here", "test.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter, got nil")
	}
}

func TestParseMissingClosingDelimiter(t *testing.T) {
	_, err := parseChangeset("---\nrepo: patch\nno closing", "test.md")
	if err == nil {
		t.Fatal("expected error for missing closing delimiter, got nil")
	}
}

func TestParseInvalidBumpType(t *testing.T) {
	_, err := parseChangeset("---\nrepo: invalid\n---\n\nmessage", "test.md")
	if err == nil {
		t.Fatal("expected error for invalid bump type, got nil")
	}
}

func TestFormat(t *testing.T) {
	result := changesetContent("my-repo", minor, "Added feature")
	expected := "---\nmy-repo: minor\n---\n\nAdded feature\n"

	if result != expected {
		t.Errorf("Format mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestListChangesets(t *testing.T) {
	dir := t.TempDir()

	// Write two changeset files
	cs1 := "---\nrepo: patch\n---\n\nFix bug"
	cs2 := "---\nrepo: minor\n---\n\nAdd feature"

	if err := os.WriteFile(filepath.Join(dir, "one.md"), []byte(cs1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "two.md"), []byte(cs2), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a non-md file that should be ignored
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	changes, err := listChangesets(dir)
	if err != nil {
		t.Fatalf("listChangesets failed: %v", err)
	}

	if len(changes) != 2 {
		t.Fatalf("expected 2 changesets, got %d", len(changes))
	}
}

func TestParseWithHorizontalRuleInBody(t *testing.T) {
	content := "---\nmy-repo: minor\n---\n\nSome summary\n\n---\n\nMore details after a horizontal rule"

	cs, err := parseChangeset(content, "test.md")
	if err != nil {
		t.Fatalf("parseChangeset failed: %v", err)
	}

	if cs.bump != minor {
		t.Errorf("expected bump minor, got %s", cs.bump)
	}

	expected := "Some summary\n\n---\n\nMore details after a horizontal rule"
	if cs.summary != expected {
		t.Errorf("expected summary to include horizontal rule content.\nExpected:\n%s\nGot:\n%s", expected, cs.summary)
	}
}

func TestHighestBumpEmpty(t *testing.T) {
	result := highestBump(nil)
	if result != patch {
		t.Errorf("highestBump(nil) = %s, expected patch (default)", result)
	}

	result = highestBump([]*changeset{})
	if result != patch {
		t.Errorf("highestBump([]) = %s, expected patch (default)", result)
	}
}

func TestHighestBump(t *testing.T) {
	tests := []struct {
		bumps    []bumpType
		expected bumpType
	}{
		{[]bumpType{patch}, patch},
		{[]bumpType{patch, minor}, minor},
		{[]bumpType{patch, minor, major}, major},
		{[]bumpType{minor, patch}, minor},
		{[]bumpType{major, patch}, major},
	}

	for _, tt := range tests {
		var changes []*changeset
		for _, b := range tt.bumps {
			changes = append(changes, &changeset{bump: b})
		}
		result := highestBump(changes)
		if result != tt.expected {
			t.Errorf("highestBump(%v) = %s, expected %s", tt.bumps, result, tt.expected)
		}
	}
}
