package workspace

import (
	"fmt"
	"os/exec"
	"strings"
)

// detectVCS detects the version control system for a directory
func detectVCS(dirPath string) (VCSType, error) {
	// Check for Git
	if isGitRepository(dirPath) {
		return VCSTypeGit, nil
	}

	// Check for Mercurial
	if isMercurialRepository(dirPath) {
		return VCSTypeMercurial, nil
	}

	return VCSTypeNone, nil
}

// isGitRepository checks if a directory is a Git repository
func isGitRepository(dirPath string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dirPath
	return cmd.Run() == nil
}

// isMercurialRepository checks if a directory is a Mercurial repository
func isMercurialRepository(dirPath string) bool {
	cmd := exec.Command("hg", "root")
	cmd.Dir = dirPath
	return cmd.Run() == nil
}

// getGitCommitHash gets the latest commit hash for a Git repository
func getGitCommitHash(dirPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dirPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitRemoteURLs gets all remote URLs for a Git repository
func getGitRemoteURLs(dirPath string) ([]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = dirPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote URLs: %w", err)
	}

	// Parse output and extract unique URLs
	lines := strings.Split(string(output), "\n")
	urlSet := make(map[string]bool)
	var urls []string

	for _, line := range lines {
		// Only process fetch URLs to avoid duplicates
		if !strings.Contains(line, "(fetch)") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			url := parts[1]
			if !urlSet[url] {
				urlSet[url] = true
				urls = append(urls, url)
			}
		}
	}

	return urls, nil
}

// getGitBranch gets the current branch name for a Git repository
func getGitBranch(dirPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dirPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
