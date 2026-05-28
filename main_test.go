package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestProject creates a temporary project directory with .changesets structure.
func setupTestProject(t *testing.T, version string, changesetContents ...string) config {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.25.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	p := newConfig(dir)
	if err := os.MkdirAll(p.changes, 0755); err != nil {
		t.Fatal(err)
	}
	if err := saveConfig(p.config, &Config{Version: version}); err != nil {
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
	if !strings.Contains(output, "init") {
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
	p := setupTestProject(t, startVersion)
	if err := ensureChangesetsExist(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureChangesetsExistMissing(t *testing.T) {
	p := newConfig(t.TempDir())
	if err := ensureChangesetsExist(p); err == nil {
		t.Fatal("expected error when .changesets missing")
	}
}

func TestCmdInitFresh(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	p := newConfig(dir)

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
	p := setupTestProject(t, startVersion)

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
	p := setupTestProject(t, startVersion)

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
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdInit(p, newScanner(""))
	})
	if err == nil {
		t.Fatal("expected error for no input")
	}
}

func TestCmdAddPatch(t *testing.T) {
	p := setupTestProject(t, startVersion)

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
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("2\nNew feature\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMajor(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("3\nBreaking change\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddPatchText(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("patch\nFix\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMinorText(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("minor\nFeat\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddMajorText(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("major\nBreaking\ny\n"))
	})
	if err != nil {
		t.Fatalf("cmdAdd failed: %v", err)
	}
}

func TestCmdAddInvalidSelection(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("invalid\n"))
	})
	if err == nil {
		t.Fatal("expected error for invalid selection")
	}
}

func TestCmdAddEmptySummary(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\n\n"))
	})
	if err == nil {
		t.Fatal("expected error for empty summary")
	}
}

func TestCmdAddAbort(t *testing.T) {
	p := setupTestProject(t, startVersion)

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
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner(""))
	})
	if err == nil {
		t.Fatal("expected error for no input on bump")
	}
}

func TestCmdAddNoInputSummary(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\n"))
	})
	if err == nil {
		t.Fatal("expected error for no input on summary")
	}
}

func TestCmdAddNoInputConfirm(t *testing.T) {
	p := setupTestProject(t, startVersion)

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\nSome change\n"))
	})
	if err == nil {
		t.Fatal("expected error for no input on confirmation")
	}
}

func TestCmdAddNoChangesetsDir(t *testing.T) {
	p := newConfig(t.TempDir())

	err := cmdAdd(p, newScanner("1\ntest\ny\n"))
	if err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdAddModuleNameError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("go 1.25.0\n"), 0644)
	p := newConfig(dir)
	os.MkdirAll(p.changesets, 0755)
	os.MkdirAll(p.changes, 0755)
	saveConfig(p.config, &Config{Version: startVersion})

	var err error
	captureStdout(func() {
		err = cmdAdd(p, newScanner("1\ntest\ny\n"))
	})
	if err == nil {
		t.Fatal("expected error when moduleName fails")
	}
}

func TestCmdAddWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: permission-based test requires non-root user")
	}
	p := setupTestProject(t, startVersion)
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
	p := setupTestProject(t, "v1.0.0", "---\ntest: minor\n---\n\nAdded feature")

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
	p := setupTestProject(t, "v1.0.0")

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
	p := newConfig(t.TempDir())
	if err := cmdNext(p); err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdNextCalculateError(t *testing.T) {
	dir := t.TempDir()
	p := newConfig(dir)
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
	p := setupTestProject(t, "v1.0.0",
		"---\ntest: minor\n---\n\nNew feature",
		"---\ntest: patch\n---\n\nBug fix",
	)

	var err error
	output := captureStdout(func() {
		err = cmdRelease(p)
	})
	if err != nil {
		t.Fatalf("cmdRelease failed: %v", err)
	}
	if !strings.Contains(output, "v1.1.0") {
		t.Errorf("expected 'v1.1.0', got %q", output)
	}

	data, readErr := os.ReadFile(p.changelog)
	if readErr != nil {
		t.Fatalf("CHANGELOG.md not created: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "## v1.1.0") {
		t.Error("CHANGELOG missing version header")
	}
	if !strings.Contains(content, "New feature") {
		t.Error("CHANGELOG missing minor summary")
	}
	if !strings.Contains(content, "Bug fix") {
		t.Error("CHANGELOG missing patch summary")
	}

	cfg, cfgErr := loadConfig(p.config)
	if cfgErr != nil {
		t.Fatalf("loadConfig: %v", cfgErr)
	}
	if cfg.Version != "v1.1.0" {
		t.Errorf("expected config version v1.1.0, got %s", cfg.Version)
	}

	remaining, listErr := listChangesets(p.changes)
	if listErr != nil {
		t.Fatalf("listChangesets: %v", listErr)
	}
	if len(remaining) != 0 {
		t.Errorf("expected changesets cleaned up, got %d remaining", len(remaining))
	}
}

func TestCmdReleaseNoChangesets(t *testing.T) {
	p := setupTestProject(t, "v1.0.0")

	captureStdout(func() {})
	err := cmdRelease(p)
	if err == nil {
		t.Fatal("expected error when no changesets")
	}
}

func TestCmdReleaseNoDir(t *testing.T) {
	p := newConfig(t.TempDir())
	if err := cmdRelease(p); err == nil {
		t.Fatal("expected error when .changesets doesn't exist")
	}
}

func TestCmdReleasePrependsToExistingChangelog(t *testing.T) {
	p := setupTestProject(t, "v1.0.0", "---\ntest: patch\n---\n\nFix")
	existing := "# Changelog\n\n## v1.0.0 - 2026-01-01\n\n### Patch Changes\n\n- Old fix\n"
	os.WriteFile(p.changelog, []byte(existing), 0644)

	captureStdout(func() { cmdRelease(p) })

	data, _ := os.ReadFile(p.changelog)
	content := string(data)
	if strings.Index(content, "v1.0.1") >= strings.Index(content, "v1.0.0") {
		t.Error("new version should appear before old version")
	}
}

func TestBuildNextPatch(t *testing.T) {
	p := setupTestProject(t, "v1.0.0", "---\ntest: patch\n---\n\nFix")

	ver, err := buildNext(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver.version != "v1.0.1" {
		t.Errorf("expected v1.0.1, got %s", ver.version)
	}
	if len(ver.changes) != 1 {
		t.Errorf("expected 1 changeset, got %d", len(ver.changes))
	}
	if ver.config == nil {
		t.Error("expected non-nil config")
	}
}

func TestBuildNextMinor(t *testing.T) {
	p := setupTestProject(t, "v1.0.0", "---\ntest: minor\n---\n\nFeat")

	ver, err := buildNext(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver.version != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %s", ver.version)
	}
}

func TestBuildNextMajor(t *testing.T) {
	p := setupTestProject(t, "v1.0.0", "---\ntest: major\n---\n\nBreaking")

	ver, err := buildNext(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver.version != "v2.0.0" {
		t.Errorf("expected v2.0.0, got %s", ver.version)
	}
}

func TestBuildNextNoChangesets(t *testing.T) {
	p := setupTestProject(t, "v1.0.0")

	ver, err := buildNext(p)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if ver.version != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", ver.version)
	}
	if len(ver.changes) != 0 {
		t.Errorf("expected 0 changesets, got %d", len(ver.changes))
	}
	if ver.config == nil {
		t.Error("expected non-nil config")
	}
}

func TestBuildNextInvalidVersion(t *testing.T) {
	p := setupTestProject(t, "not-a-version", "---\ntest: patch\n---\n\nFix")

	_, err := buildNext(p)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestBuildNextConfigMissing(t *testing.T) {
	dir := t.TempDir()
	p := newConfig(dir)
	os.MkdirAll(p.changes, 0755)

	_, err := buildNext(p)
	if err == nil {
		t.Fatal("expected error when config missing")
	}
}

func TestBuildNextListError(t *testing.T) {
	dir := t.TempDir()
	p := newConfig(dir)
	os.MkdirAll(p.changesets, 0755)
	saveConfig(p.config, &Config{Version: "v1.0.0"})

	_, err := buildNext(p)
	if err == nil {
		t.Fatal("expected error when changes dir missing")
	}
}

func TestCmdAddPatchFileContent(t *testing.T) {
	p := setupTestProject(t, startVersion)
	captureStdout(func() {
		if err := cmdAdd(p, newScanner("1\nFix bug\ny\n")); err != nil {
			t.Fatalf("cmdAdd failed: %v", err)
		}
	})
	changes, err := listChangesets(p.changes)
	if err != nil {
		t.Fatalf("listChangesets: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 changeset, got %d", len(changes))
	}
	if changes[0].bump != patch {
		t.Errorf("expected patch, got %q", changes[0].bump)
	}
	if changes[0].summary != "Fix bug" {
		t.Errorf("expected 'Fix bug', got %q", changes[0].summary)
	}
}

func TestCmdAddMinorFileContent(t *testing.T) {
	p := setupTestProject(t, startVersion)
	captureStdout(func() {
		if err := cmdAdd(p, newScanner("2\nNew feature\ny\n")); err != nil {
			t.Fatalf("cmdAdd failed: %v", err)
		}
	})
	changes, err := listChangesets(p.changes)
	if err != nil {
		t.Fatalf("listChangesets: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 changeset, got %d", len(changes))
	}
	if changes[0].bump != minor {
		t.Errorf("expected minor, got %q", changes[0].bump)
	}
	if changes[0].summary != "New feature" {
		t.Errorf("expected 'New feature', got %q", changes[0].summary)
	}
}

func TestCmdAddMajorFileContent(t *testing.T) {
	p := setupTestProject(t, startVersion)
	captureStdout(func() {
		if err := cmdAdd(p, newScanner("3\nBreaking change\ny\n")); err != nil {
			t.Fatalf("cmdAdd failed: %v", err)
		}
	})
	changes, err := listChangesets(p.changes)
	if err != nil {
		t.Fatalf("listChangesets: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 changeset, got %d", len(changes))
	}
	if changes[0].bump != major {
		t.Errorf("expected major, got %q", changes[0].bump)
	}
	if changes[0].summary != "Breaking change" {
		t.Errorf("expected 'Breaking change', got %q", changes[0].summary)
	}
}

func TestBuildChangelogSectionMultiplePatch(t *testing.T) {
	changes := []*changeset{
		{filePath: "/nope.md", bump: patch, summary: "Fix one"},
		{filePath: "/nope2.md", bump: patch, summary: "Fix two"},
	}
	result := buildChangelogSection("v1.0.2", changes)
	if !strings.Contains(result, "Fix one") || !strings.Contains(result, "Fix two") {
		t.Error("missing patch summaries")
	}
	if strings.Contains(result, "Major Changes") || strings.Contains(result, "Minor Changes") {
		t.Error("unexpected sections present")
	}
}

func TestBuildChangelogSectionMultipleMinor(t *testing.T) {
	changes := []*changeset{
		{filePath: "/nope.md", bump: minor, summary: "Feat one"},
		{filePath: "/nope2.md", bump: minor, summary: "Feat two"},
	}
	result := buildChangelogSection("v1.1.0", changes)
	if !strings.Contains(result, "Feat one") || !strings.Contains(result, "Feat two") {
		t.Error("missing minor summaries")
	}
	if strings.Contains(result, "Major Changes") || strings.Contains(result, "Patch Changes") {
		t.Error("unexpected sections present")
	}
}

func TestBuildChangelogSectionMixedMultiple(t *testing.T) {
	changes := []*changeset{
		{filePath: "/a.md", bump: major, summary: "Break A"},
		{filePath: "/b.md", bump: major, summary: "Break B"},
		{filePath: "/c.md", bump: minor, summary: "Feat C"},
		{filePath: "/d.md", bump: patch, summary: "Fix D"},
		{filePath: "/e.md", bump: patch, summary: "Fix E"},
	}
	result := buildChangelogSection("v2.0.0", changes)

	majorIdx := strings.Index(result, "### Major Changes")
	minorIdx := strings.Index(result, "### Minor Changes")
	patchIdx := strings.Index(result, "### Patch Changes")
	if majorIdx < 0 || minorIdx < 0 || patchIdx < 0 {
		t.Fatal("missing one or more change sections")
	}
	if !(majorIdx < minorIdx && minorIdx < patchIdx) {
		t.Error("sections must appear in order: major, minor, patch")
	}
	for _, s := range []string{"Break A", "Break B", "Feat C", "Fix D", "Fix E"} {
		if !strings.Contains(result, s) {
			t.Errorf("missing summary %q", s)
		}
	}
}

func TestBuildChangelogSection(t *testing.T) {
	changes := []*changeset{
		{filePath: "test1.md", bump: major, summary: "Breaking change"},
		{filePath: "test2.md", bump: minor, summary: "New feature"},
		{filePath: "test3.md", bump: patch, summary: "Bug fix"},
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
		{filePath: "test.md", bump: patch, summary: "Fix"},
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
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "change.md"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "change.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add change").Run()

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	changes := []*changeset{
		{filePath: "change.md", bump: patch, summary: "Updated deps"},
	}

	result := buildChangelogSection("v1.0.1", changes)

	if !strings.Contains(result, ": Updated deps") {
		t.Error("expected SHA-prefixed entry for git-tracked file")
	}
}

func TestBuildChangelogSectionWithoutSHA(t *testing.T) {
	changes := []*changeset{
		{filePath: "/nonexistent/file.md", bump: patch, summary: "Fix"},
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
	p := newConfig(dir)
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
