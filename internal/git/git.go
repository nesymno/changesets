package git

import (
	"os/exec"
	"strings"
)

// GetFileCommitSHA returns the short SHA of the commit that added the given file.
// It shells out to: git log --diff-filter=A --format=%h -- <filepath>
// Returns an empty string if the file is not yet tracked by git.
func GetFileCommitSHA(filePath string) string {
	cmd := exec.Command("git", "log", "--diff-filter=A", "--format=%h", "--", filePath)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	sha := strings.TrimSpace(string(out))
	// git log may return multiple lines if the file was added multiple times
	// (e.g., deleted and re-added). Take the last one (oldest add).
	lines := strings.Split(sha, "\n")
	if len(lines) == 0 {
		return ""
	}

	// The last line is the original commit that added the file.
	result := strings.TrimSpace(lines[len(lines)-1])
	return result
}
