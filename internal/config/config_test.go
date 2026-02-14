package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{Version: "v1.2.3"}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", loaded.Version)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestModuleName(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")

	content := "module github.com/example/my-tool\n\ngo 1.25.0\n"
	if err := os.WriteFile(gomod, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	name, err := ModuleName(dir)
	if err != nil {
		t.Fatalf("ModuleName failed: %v", err)
	}

	if name != "my-tool" {
		t.Errorf("expected my-tool, got %s", name)
	}
}

func TestModuleNameMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := ModuleName(dir)
	if err == nil {
		t.Fatal("expected error for missing go.mod, got nil")
	}
}

func TestModuleNameNoDirective(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")

	if err := os.WriteFile(gomod, []byte("go 1.25.0\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	_, err := ModuleName(dir)
	if err == nil {
		t.Fatal("expected error for go.mod without module directive, got nil")
	}
}

func TestResolvePaths(t *testing.T) {
	paths := ResolvePaths("/project")

	if paths.Root != "/project" {
		t.Errorf("expected root /project, got %s", paths.Root)
	}
	if paths.Changesets != "/project/.changesets" {
		t.Errorf("expected changesets /project/.changesets, got %s", paths.Changesets)
	}
	if paths.Config != "/project/.changesets/config.json" {
		t.Errorf("expected config /project/.changesets/config.json, got %s", paths.Config)
	}
	if paths.Changes != "/project/.changesets/changes" {
		t.Errorf("expected changes /project/.changesets/changes, got %s", paths.Changes)
	}
}
