package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	git("init")
	git("config", "user.email", "nesymno@gmail.com")
	git("config", "user.name", "nesymno")

	return dir
}

func TestGetFileCommitSHA(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("hello"), 0644)
	exec.Command("git", "-C", dir, "add", "tracked.txt").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add tracked file").Run()

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	sha, err := getFileCommitSHA("tracked.txt")
	if err != nil {
		t.Fatalf("getFileCommitSHA failed: %v", err)
	}

	if sha == "" {
		t.Error("expected non-empty SHA for tracked file")
	}
}

func TestGetFileCommitSHAUntracked(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "dummy.txt"), []byte("x"), 0644)
	exec.Command("git", "-C", dir, "add", "dummy.txt").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "initial").Run()

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	sha, err := getFileCommitSHA("nonexistent.txt")
	if err != nil {
		t.Fatalf("getFileCommitSHA failed: %v", err)
	}

	if sha != "" {
		t.Errorf("expected empty SHA for untracked file, got %q", sha)
	}
}

func TestGetFileCommitSHAGitNotFound(t *testing.T) {
	t.Setenv("PATH", "/nonexistent")

	_, err := getFileCommitSHA("go.mod")
	if err == nil {
		t.Fatal("expected error when git is not in PATH, got nil")
	}
}
