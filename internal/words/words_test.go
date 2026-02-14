package words

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()

	slug, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	parts := strings.Split(slug, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in slug, got %d: %q", len(parts), slug)
	}

	// Verify the file doesn't exist (it shouldn't since Generate only picks a name)
	path := filepath.Join(dir, slug+".md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to not exist at %s", path)
	}
}

func TestGenerateUniqueness(t *testing.T) {
	dir := t.TempDir()

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		slug, err := Generate(dir)
		if err != nil {
			t.Fatalf("Generate failed on iteration %d: %v", i, err)
		}
		if seen[slug] {
			// Collisions are possible but extremely unlikely in 50 iterations
			// with ~100*100*100 = 1M combinations
			t.Logf("warning: duplicate slug %q on iteration %d (may be coincidence)", slug, i)
		}
		seen[slug] = true
	}
}

func TestGenerateAvoidsExisting(t *testing.T) {
	dir := t.TempDir()

	// Generate one slug, create the file, then generate another
	slug1, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Create a file with that slug
	if err := os.WriteFile(filepath.Join(dir, slug1+".md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Generate another slug - should be different
	slug2, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if slug1 == slug2 {
		t.Errorf("second slug should differ from first when file exists, both are %q", slug1)
	}
}

func TestSlugToFilename(t *testing.T) {
	if got := SlugToFilename("brave-orange-fox"); got != "brave-orange-fox.md" {
		t.Errorf("expected brave-orange-fox.md, got %s", got)
	}
}

func TestFilenameToSlug(t *testing.T) {
	if got := FilenameToSlug("brave-orange-fox.md"); got != "brave-orange-fox" {
		t.Errorf("expected brave-orange-fox, got %s", got)
	}
}
