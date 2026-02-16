package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupProject creates a temporary project directory with .changesets structure.
func setupProject(t *testing.T, version string, changesetContents ...string) paths {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.25.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	p := newPaths(dir)
	if err := os.MkdirAll(p.changes, 0755); err != nil {
		t.Fatal(err)
	}
	if err := saveConfig(p.config, &config{Version: version}); err != nil {
		t.Fatal(err)
	}

	for i, content := range changesetContents {
		name := fmt.Sprintf("change-%d.md", i)
		if err := os.WriteFile(filepath.Join(p.changes, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(p.gitkeep, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	return p
}

func newScanner(input string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(input))
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunNoArgs(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets"}, strings.NewReader(""))
	})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunHelp(t *testing.T) {
	var code int
	output := captureStdout(func() {
		code = run([]string{"changesets", "help"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "changesets") {
		t.Error("help output missing")
	}
}

func TestRunHelpLong(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "--help"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunHelpShort(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "-h"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunVersion(t *testing.T) {
	var code int
	output := captureStdout(func() {
		code = run([]string{"changesets", "version"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected version 'dev', got %q", strings.TrimSpace(output))
	}
}

func TestRunVersionLong(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "--version"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunVersionShort(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "-v"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "foobar"}, strings.NewReader(""))
	})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunResolvePathsError(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "init"}, strings.NewReader(""))
	})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "init"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunAdd(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	captureStdout(func() { run([]string{"changesets", "init"}, strings.NewReader("")) })

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "add"}, strings.NewReader("1\nFix bug\ny\n"))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunNext(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	captureStdout(func() { run([]string{"changesets", "init"}, strings.NewReader("")) })

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "next"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunRelease(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	captureStdout(func() { run([]string{"changesets", "init"}, strings.NewReader("")) })
	captureStdout(func() { run([]string{"changesets", "add"}, strings.NewReader("1\nFix\ny\n")) })

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "release"}, strings.NewReader(""))
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunCommandError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	var code int
	captureStdout(func() {
		code = run([]string{"changesets", "next"}, strings.NewReader(""))
	})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestPrintUsage(t *testing.T) {
	output := captureStdout(printUsage)
	if !strings.Contains(output, "changesets - Manage changelogs") {
		t.Error("missing header text")
	}
	if !strings.Contains(output, "init") || !strings.Contains(output, "release") {
		t.Error("missing command descriptions")
	}
}

func TestResolvePaths(t *testing.T) {
	p, err := resolvePaths()
	if err != nil {
		t.Fatalf("resolvePaths failed: %v", err)
	}
	if p.root == "" {
		t.Error("expected non-empty root")
	}
}

func TestResolvePathsError(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	_, err := resolvePaths()
	if err == nil {
		t.Fatal("expected error when not in a Go project")
	}
}

func TestEnsureChangesetsExist(t *testing.T) {
	p := setupProject(t, "v0.0.0")
	if err := ensureChangesetsExist(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureChangesetsExistMissing(t *testing.T) {
	p := newPaths(t.TempDir())
	if err := ensureChangesetsExist(p); err == nil {
		t.Fatal("expected error when .changesets missing")
	}
}

func TestCmdInitFresh(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	p := newPaths(dir)

	var err error
	output := captureStdout(func() {
		err = cmdInit(p, newScanner(""))
	})
	if err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}
	if !strings.Contains(output, "Initialized") {
		t.Error("expected 'Initialized' message")
	}
	for _, path := range []string{p.changes, p.config, p.readme, p.gitkeep} {
		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("expected %s to exist", path)
		}
	}
}

func TestCmdInitExistingYes(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	output := captureStdout(func() {
		err = cmdInit(p, newScanner("y\n"))
	})
	if err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}
	if !strings.Contains(output, "Initialized") {
		t.Error("expected 'Initialized' after recreate")
	}
}

func TestCmdInitExistingNo(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	output := captureStdout(func() {
		err = cmdInit(p, newScanner("n\n"))
	})
	if err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}
	if !strings.Contains(output, "Aborted") {
		t.Error("expected 'Aborted' message")
	}
}

func TestCmdInitExistingNoInput(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdInit(p, newScanner(""))
	})
	if err == nil {
		t.Fatal("expected error for no input")
	}
}

func TestCmdAddPatch(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	output := captureStdout(func() {
		err = cmdAdd(p, newScanner("1\nFixed a bug\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
	if !strings.Contains(output, "Created changeset") {
		t.Error("expected 'Created changeset' message")
	}
}

func TestCmdAddMinor(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("2\nNew feature\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMajor(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("3\nBreaking change\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddPatchText(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("patch\nFix\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMinorText(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("minor\nFeat\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMajorText(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("major\nBreaking\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddInvalidSelection(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("invalid\n"))
	})
	if err == nil {
		t.Fatal("expected error for invalid selection")
	}
}

func TestCmdAddEmptySummary(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\n\n"))
	})
	if err == nil {
		t.Fatal("expected error for empty summary")
	}
}

func TestCmdAddAbort(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	output := captureStdout(func() {
		err = cmdAdd(p, newScanner("1\nSome change\nn\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd should not error on abort: %v", err)
	}
	if !strings.Contains(output, "Aborted") {
		t.Error("expected 'Aborted' message")
	}
}

func TestCmdAddNoInputBump(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner(""))
	})
	if err == nil {
		t.Fatal("expected error for no input on bump")
	}
}

func TestCmdAddNoInputSummary(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\n"))
	})
	if err == nil {
		t.Fatal("expected error for no input on summary")
	}
}

func TestCmdAddNoInputConfirm(t *testing.T) {
	p := setupProject(t, "v0.0.0")

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\nSome change\n"))
	})
	if err == nil {
		t.Fatal("expected error for no input on confirmation")
	}
}

func TestCmdAddNoChangesetsDir(t *testing.T) {
	p := newPaths(t.TempDir())

	err := cmdAdd(p, newScanner("1\ntest\ny\n"))
	if err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdAddModuleNameError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("go 1.25.0\n"), 0644)
	p := newPaths(dir)
	os.MkdirAll(p.changesets, 0755)
	os.MkdirAll(p.changes, 0755)
	saveConfig(p.config, &config{Version: "v0.0.0"})

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\ntest\ny\n"))
	})
	if err == nil {
		t.Fatal("expected error when moduleName fails")
	}
}

func TestCmdAddWriteError(t *testing.T) {
	p := setupProject(t, "v0.0.0")
	os.Chmod(p.changes, 0555)
	defer os.Chmod(p.changes, 0755)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\ntest change\ny\n"))
	})
	if err == nil {
		t.Fatal("expected error when changes dir is read-only")
	}
}

func TestCmdNext(t *testing.T) {
	p := setupProject(t, "v1.0.0", "---\ntest: minor\n---\n\nAdded feature")

	var err error
	output := captureStdout(func() {
		err = cmdNext(p)
	})
	if err != nil {
		t.Fatalf("cmdNext failed: %v", err)
	}
	if strings.TrimSpace(output) != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %q", strings.TrimSpace(output))
	}
}

func TestCmdNextNoChangesets(t *testing.T) {
	p := setupProject(t, "v1.0.0")

	var err error
	output := captureStdout(func() {
		err = cmdNext(p)
	})
	if err != nil {
		t.Fatalf("cmdNext failed: %v", err)
	}
	if strings.TrimSpace(output) != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %q", strings.TrimSpace(output))
	}
}

func TestCmdNextNoDir(t *testing.T) {
	p := newPaths(t.TempDir())
	if err := cmdNext(p); err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdNextCalculateError(t *testing.T) {
	dir := t.TempDir()
	p := newPaths(dir)
	os.MkdirAll(p.changesets, 0755)
	os.MkdirAll(p.changes, 0755)

	var err error
	captureStdout(func() {
		err = cmdNext(p)
	})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

func TestCmdRelease(t *testing.T) {
	p := setupProject(t, "v1.0.0", "---\ntest: patch\n---\n\nFixed bug")

	var err error
	output := captureStdout(func() {
		err = cmdRelease(p)
	})
	if err != nil {
		t.Fatalf("cmdRelease failed: %v", err)
	}
	if strings.TrimSpace(output) != "v1.0.1" {
		t.Errorf("expected v1.0.1, got %q", strings.TrimSpace(output))
	}

	data, readErr := os.ReadFile(filepath.Join(p.root, "CHANGELOG.md"))
	if readErr != nil {
		t.Fatal("CHANGELOG.md not created")
	}
	if !strings.Contains(string(data), "v1.0.1") {
		t.Error("CHANGELOG.md missing version")
	}

	cfg, _ := loadConfig(p.config)
	if cfg.Version != "v1.0.1" {
		t.Errorf("expected config v1.0.1, got %s", cfg.Version)
	}

	entries, _ := os.ReadDir(p.changes)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			t.Errorf("changeset %s should be removed", e.Name())
		}
	}
}

func TestCmdReleaseNoChangesets(t *testing.T) {
	p := setupProject(t, "v1.0.0")

	var err error
	captureStdout(func() {
		err = cmdRelease(p)
	})
	if err == nil {
		t.Fatal("expected error when no changesets")
	}
}

func TestCmdReleaseNoDir(t *testing.T) {
	p := newPaths(t.TempDir())
	if err := cmdRelease(p); err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdReleaseCalculateError(t *testing.T) {
	dir := t.TempDir()
	p := newPaths(dir)
	os.MkdirAll(p.changesets, 0755)
	os.MkdirAll(p.changes, 0755)

	err := cmdRelease(p)
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

func TestCalculateNextVersionPatch(t *testing.T) {
	p := setupProject(t, "v1.0.0", "---\ntest: patch\n---\n\nFix")

	ver, changes, cfg, err := calculateNextVersion(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver != "v1.0.1" {
		t.Errorf("expected v1.0.1, got %s", ver)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 changeset, got %d", len(changes))
	}
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

func TestCalculateNextVersionMinor(t *testing.T) {
	p := setupProject(t, "v1.0.0", "---\ntest: minor\n---\n\nFeat")

	ver, _, _, err := calculateNextVersion(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %s", ver)
	}
}

func TestCalculateNextVersionMajor(t *testing.T) {
	p := setupProject(t, "v1.0.0", "---\ntest: major\n---\n\nBreaking")

	ver, _, _, err := calculateNextVersion(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver != "v2.0.0" {
		t.Errorf("expected v2.0.0, got %s", ver)
	}
}

func TestCalculateNextVersionNoChangesets(t *testing.T) {
	p := setupProject(t, "v1.0.0")

	ver, changes, cfg, err := calculateNextVersion(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", ver)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changesets, got %d", len(changes))
	}
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

func TestCalculateNextVersionInvalidVersion(t *testing.T) {
	p := setupProject(t, "not-a-version", "---\ntest: patch\n---\n\nFix")

	_, _, _, err := calculateNextVersion(p)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestCalculateNextVersionConfigMissing(t *testing.T) {
	dir := t.TempDir()
	p := newPaths(dir)
	os.MkdirAll(p.changes, 0755)

	_, _, _, err := calculateNextVersion(p)
	if err == nil {
		t.Fatal("expected error when config missing")
	}
}

func TestCalculateNextVersionListError(t *testing.T) {
	dir := t.TempDir()
	p := newPaths(dir)
	os.MkdirAll(p.changesets, 0755)
	saveConfig(p.config, &config{Version: "v1.0.0"})

	_, _, _, err := calculateNextVersion(p)
	if err == nil {
		t.Fatal("expected error when changes dir missing")
	}
}

func TestBuildChangelogSection(t *testing.T) {
	changes := []*changeset{
		{filepath: "test1.md", bump: major, summary: "Breaking change"},
		{filepath: "test2.md", bump: minor, summary: "New feature"},
		{filepath: "test3.md", bump: patch, summary: "Bug fix"},
	}

	result := buildChangelogSection("v2.0.0", changes)

	if !strings.Contains(result, "## v2.0.0") {
		t.Error("missing version header")
	}
	if !strings.Contains(result, "### Major Changes") {
		t.Error("missing Major Changes")
	}
	if !strings.Contains(result, "### Minor Changes") {
		t.Error("missing Minor Changes")
	}
	if !strings.Contains(result, "### Patch Changes") {
		t.Error("missing Patch Changes")
	}
	if !strings.Contains(result, "Breaking change") {
		t.Error("missing major summary")
	}
	if !strings.Contains(result, "New feature") {
		t.Error("missing minor summary")
	}
	if !strings.Contains(result, "Bug fix") {
		t.Error("missing patch summary")
	}
}

func TestBuildChangelogSectionEmptyGroups(t *testing.T) {
	changes := []*changeset{
		{filepath: "test.md", bump: patch, summary: "Fix"},
	}

	result := buildChangelogSection("v1.0.1", changes)

	if strings.Contains(result, "Major Changes") {
		t.Error("should not have Major Changes")
	}
	if strings.Contains(result, "Minor Changes") {
		t.Error("should not have Minor Changes")
	}
	if !strings.Contains(result, "Patch Changes") {
		t.Error("missing Patch Changes")
	}
}

func TestBuildChangelogSectionWithSHA(t *testing.T) {
	changes := []*changeset{
		{filepath: "go.mod", bump: patch, summary: "Updated deps"},
	}

	result := buildChangelogSection("v1.0.1", changes)

	if !strings.Contains(result, ": Updated deps") {
		t.Error("expected SHA-prefixed entry for git-tracked file")
	}
}

func TestBuildChangelogSectionWithoutSHA(t *testing.T) {
	changes := []*changeset{
		{filepath: "/nonexistent/file.md", bump: patch, summary: "Fix"},
	}

	result := buildChangelogSection("v1.0.1", changes)

	if !strings.Contains(result, "- Fix\n") {
		t.Error("expected plain entry without SHA for non-tracked file")
	}
}

func TestPrependChangelogNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")

	if err := prependChangelog(path, "## v1.0.0\n\n- Fix\n"); err != nil {
		t.Fatalf("failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "# Changelog") {
		t.Error("expected '# Changelog' header")
	}
	if !strings.Contains(string(data), "v1.0.0") {
		t.Error("missing version")
	}
}

func TestPrependChangelogExistingWithHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	os.WriteFile(path, []byte("# Changelog\n\n## v0.1.0\n\n- Old\n"), 0644)

	if err := prependChangelog(path, "## v1.0.0\n\n- New\n"); err != nil {
		t.Fatalf("failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.HasPrefix(content, "# Changelog\n") {
		t.Error("header should be preserved")
	}
	if strings.Index(content, "v1.0.0") >= strings.Index(content, "v0.1.0") {
		t.Error("new version should come before old")
	}
}

func TestPrependChangelogHeaderNoNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	os.WriteFile(path, []byte("# Changelog"), 0644)

	if err := prependChangelog(path, "## v1.0.0\n"); err != nil {
		t.Fatalf("failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "Changelog") || !strings.Contains(content, "v1.0.0") {
		t.Error("content missing expected parts")
	}
}

func TestPrependChangelogNoHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	os.WriteFile(path, []byte("existing content\n"), 0644)

	if err := prependChangelog(path, "## v1.0.0\n"); err != nil {
		t.Fatalf("failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.HasPrefix(content, "## v1.0.0") {
		t.Error("new section should be prepended")
	}
	if !strings.Contains(content, "existing content") {
		t.Error("existing content should be preserved")
	}
}

func TestPrependChangelogWriteError(t *testing.T) {
	err := prependChangelog("/nonexistent/nested/CHANGELOG.md", "## v1.0.0\n")
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}

func TestCleanupChanges(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "one.md"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "two.md"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0644)

	if err := cleanupChanges(dir); err != nil {
		t.Fatalf("failed: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			t.Errorf("expected .md files removed, found %s", e.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitkeep")); err != nil {
		t.Error(".gitkeep should remain")
	}
}

func TestCleanupChangesWithSubdir(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)
	os.WriteFile(filepath.Join(dir, "test.md"), []byte("x"), 0644)

	if err := cleanupChanges(dir); err != nil {
		t.Fatalf("failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "subdir")); err != nil {
		t.Error("subdirectory should remain")
	}
}

func TestCleanupChangesInvalidDir(t *testing.T) {
	err := cleanupChanges("/nonexistent/dir")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestCleanupChangesRemoveError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: permission-based test requires non-root user")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.md"), []byte("x"), 0644)
	os.Chmod(dir, 0555)
	defer os.Chmod(dir, 0755)

	err := cleanupChanges(dir)
	if err == nil {
		t.Fatal("expected error when file can't be removed")
	}
}

func TestCmdInitMkdirError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: permission-based test requires non-root user")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	p := newPaths(dir)
	os.Chmod(dir, 0555)
	defer os.Chmod(dir, 0755)

	var err error
	captureStdout(func() {
		err = cmdInit(p, newScanner(""))
	})
	if err == nil {
		t.Fatal("expected error when parent dir is read-only")
	}
}
