package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &config{Version: "v1.2.3"}
	if err := saveConfig(path, cfg); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	loaded, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
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

	_, err := loadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := loadConfig("/nonexistent/config.json")
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

	name, err := moduleName(dir)
	if err != nil {
		t.Fatalf("moduleName failed: %v", err)
	}

	if name != "my-tool" {
		t.Errorf("expected my-tool, got %s", name)
	}
}

func TestModuleNameMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := moduleName(dir)
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

	_, err := moduleName(dir)
	if err == nil {
		t.Fatal("expected error for go.mod without module directive, got nil")
	}
}

func TestNewPaths(t *testing.T) {
	p := newPaths("/project")

	if p.root != "/project" {
		t.Errorf("expected root /project, got %s", p.root)
	}
	if p.changesets != "/project/.changesets" {
		t.Errorf("expected changesets /project/.changesets, got %s", p.changesets)
	}
	if p.config != "/project/.changesets/config.json" {
		t.Errorf("expected config /project/.changesets/config.json, got %s", p.config)
	}
	if p.changes != "/project/.changesets/changes" {
		t.Errorf("expected changes /project/.changesets/changes, got %s", p.changes)
	}
}

func TestFindRoot(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatalf("findRoot failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Errorf("go.mod not found in returned root %s", root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	_, err = findRoot()
	if err == nil {
		t.Fatal("expected error when no go.mod in parent chain, got nil")
	}
}

func TestSaveConfigWriteError(t *testing.T) {
	err := saveConfig("/nonexistent/deeply/nested/config.json", &config{Version: "v1.0.0"})
	if err == nil {
		t.Fatal("expected error writing to nonexistent path, got nil")
	}
}
