package changeset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	content := "---\nmy-repo: minor\n---\n\nAdded something cool"

	cs, err := Parse(content, "test.md")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cs.RepoName != "my-repo" {
		t.Errorf("expected repo name my-repo, got %s", cs.RepoName)
	}
	if cs.Bump != Minor {
		t.Errorf("expected bump minor, got %s", cs.Bump)
	}
	if cs.Summary != "Added something cool" {
		t.Errorf("expected summary 'Added something cool', got %q", cs.Summary)
	}
}

func TestParseAllBumpTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected BumpType
	}{
		{"---\nrepo: patch\n---\n\nfix", Patch},
		{"---\nrepo: minor\n---\n\nfeat", Minor},
		{"---\nrepo: major\n---\n\nbreaking", Major},
	}

	for _, tt := range tests {
		cs, err := Parse(tt.input, "test.md")
		if err != nil {
			t.Fatalf("Parse failed for %s: %v", tt.expected, err)
		}
		if cs.Bump != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, cs.Bump)
		}
	}
}

func TestParseMissingFrontmatter(t *testing.T) {
	_, err := Parse("no frontmatter here", "test.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter, got nil")
	}
}

func TestParseMissingClosingDelimiter(t *testing.T) {
	_, err := Parse("---\nrepo: patch\nno closing", "test.md")
	if err == nil {
		t.Fatal("expected error for missing closing delimiter, got nil")
	}
}

func TestParseInvalidBumpType(t *testing.T) {
	_, err := Parse("---\nrepo: invalid\n---\n\nmessage", "test.md")
	if err == nil {
		t.Fatal("expected error for invalid bump type, got nil")
	}
}

func TestFormat(t *testing.T) {
	result := ChangesetContent("my-repo", Minor, "Added feature")
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

	changesets, err := ListChangesets(dir)
	if err != nil {
		t.Fatalf("ListChangesets failed: %v", err)
	}

	if len(changesets) != 2 {
		t.Fatalf("expected 2 changesets, got %d", len(changesets))
	}
}

func TestHighestBump(t *testing.T) {
	tests := []struct {
		bumps    []BumpType
		expected BumpType
	}{
		{[]BumpType{Patch}, Patch},
		{[]BumpType{Patch, Minor}, Minor},
		{[]BumpType{Patch, Minor, Major}, Major},
		{[]BumpType{Minor, Patch}, Minor},
		{[]BumpType{Major, Patch}, Major},
	}

	for _, tt := range tests {
		var changesets []*Changeset
		for _, b := range tt.bumps {
			changesets = append(changesets, &Changeset{Bump: b})
		}
		result := HighestBump(changesets)
		if result != tt.expected {
			t.Errorf("HighestBump(%v) = %s, expected %s", tt.bumps, result, tt.expected)
		}
	}
}
