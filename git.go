package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// getFileCommitSHA returns the short SHA of the commit that added the given file.
// It shells out to: git log --diff-filter=A --format=%h -- <filepath>
// Returns an empty string and nil error if the file is not yet tracked by git.
// Returns an error if the git command fails for other reasons.
func getFileCommitSHA(filePath string) (string, error) {
	cmd := exec.Command("git", "log", "--diff-filter=A", "--format=%h", "--", filePath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed for %s: %w", filePath, err)
	}

	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", nil
	}

	// git log may return multiple lines if the file was added multiple times
	// (e.g., deleted and re-added). Take the last one (oldest add).
	lines := strings.Split(sha, "\n")

	// The last line is the original commit that added the file.
	result := strings.TrimSpace(lines[len(lines)-1])
	return result, nil
}
