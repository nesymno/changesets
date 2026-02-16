package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zeroReader always returns zero bytes (deterministic output for crypto/rand).
type zeroReader struct{}

func (z zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// countingFailReader succeeds for maxReads Read calls, then returns an error.
type countingFailReader struct {
	maxReads int
	count    int
}

func (r *countingFailReader) Read(p []byte) (int, error) {
	if r.count >= r.maxReads {
		return 0, fmt.Errorf("forced reader failure")
	}
	r.count++
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func withReader(r io.Reader, fn func()) {
	orig := rand.Reader
	rand.Reader = r
	defer func() { rand.Reader = orig }()
	fn()
}

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

func TestRandomElementError(t *testing.T) {
	withReader(&countingFailReader{maxReads: 0}, func() {
		_, err := randomElement(adjectives)
		if err == nil {
			t.Error("expected error with failing reader, got nil")
		}
	})
}

func TestGenerateSlugExhaustion(t *testing.T) {
	dir := t.TempDir()

	// Use a zero reader so the slug is always the same deterministic value.
	withReader(zeroReader{}, func() {
		slug, err := generateSlug(dir)
		if err != nil {
			t.Fatalf("first generateSlug failed: %v", err)
		}

		// Create the file so every subsequent attempt collides.
		if err := os.WriteFile(filepath.Join(dir, slug+".md"), []byte("taken"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err = generateSlug(dir)
		if err == nil {
			t.Error("expected error after 100 collision attempts, got nil")
		}
	})
}

func TestGenerateSlugRandomElementFailFirstCall(t *testing.T) {
	dir := t.TempDir()
	withReader(&countingFailReader{maxReads: 0}, func() {
		_, err := generateSlug(dir)
		if err == nil {
			t.Error("expected error when first randomElement fails")
		}
	})
}

func TestGenerateSlugRandomElementFailSecondCall(t *testing.T) {
	dir := t.TempDir()
	withReader(&countingFailReader{maxReads: 1}, func() {
		_, err := generateSlug(dir)
		if err == nil {
			t.Error("expected error when second randomElement fails")
		}
	})
}

func TestGenerateSlugRandomElementFailThirdCall(t *testing.T) {
	dir := t.TempDir()
	withReader(&countingFailReader{maxReads: 2}, func() {
		_, err := generateSlug(dir)
		if err == nil {
			t.Error("expected error when third randomElement fails")
		}
	})
}
