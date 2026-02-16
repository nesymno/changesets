package main

import (
	"testing"
)

func TestGetFileCommitSHA(t *testing.T) {
	sha, err := getFileCommitSHA("go.mod")
	if err != nil {
		t.Fatalf("getFileCommitSHA failed: %v", err)
	}

	if sha == "" {
		t.Error("expected non-empty SHA for tracked file go.mod")
	}
}

func TestGetFileCommitSHAUntracked(t *testing.T) {
	// Use a filename that doesn't exist in git history but is inside the repo.
	sha, err := getFileCommitSHA("nonexistent_file_that_was_never_committed.txt")
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
