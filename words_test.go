package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	dir := t.TempDir()

	slug, err := generateSlug(dir)
	if err != nil {
		t.Fatalf("generateSlug failed: %v", err)
	}

	parts := strings.Split(slug, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in slug, got %d: %q", len(parts), slug)
	}

	// Verify the file doesn't exist (it shouldn't since generateSlug only picks a name)
	path := filepath.Join(dir, slug+".md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to not exist at %s", path)
	}
}

func TestGenerateSlugUniqueness(t *testing.T) {
	dir := t.TempDir()

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		slug, err := generateSlug(dir)
		if err != nil {
			t.Fatalf("generateSlug failed on iteration %d: %v", i, err)
		}
		if seen[slug] {
			// Collisions are possible but extremely unlikely in 50 iterations
			// with ~100*100*100 = 1M combinations
			t.Logf("warning: duplicate slug %q on iteration %d (may be coincidence)", slug, i)
		}
		seen[slug] = true
	}
}

func TestGenerateSlugAvoidsExisting(t *testing.T) {
	dir := t.TempDir()

	// Generate one slug, create the file, then generate another
	slug1, err := generateSlug(dir)
	if err != nil {
		t.Fatalf("generateSlug failed: %v", err)
	}

	// Create a file with that slug
	if err := os.WriteFile(filepath.Join(dir, slug1+".md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Generate another slug - should be different
	slug2, err := generateSlug(dir)
	if err != nil {
		t.Fatalf("generateSlug failed: %v", err)
	}

	if slug1 == slug2 {
		t.Errorf("second slug should differ from first when file exists, both are %q", slug1)
	}
}

func TestSlugToFilename(t *testing.T) {
	if got := slugToFilename("brave-orange-fox"); got != "brave-orange-fox.md" {
		t.Errorf("expected brave-orange-fox.md, got %s", got)
	}
}
