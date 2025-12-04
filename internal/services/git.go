package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInfo holds git repository information
type GitInfo struct {
	IsGitRepo      bool     `json:"isGitRepo"`
	CurrentBranch  string   `json:"currentBranch"`
	Branches       []string `json:"branches"`
	HasUncommitted bool     `json:"hasUncommitted"`
}

// IsGitRepository checks if a directory is a git repository
func IsGitRepository(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetCurrentBranch returns the current git branch
func GetCurrentBranch(dir string) (string, error) {
	if !IsGitRepository(dir) {
		return "", fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// GetBranches returns all local branches
func GetBranches(dir string) ([]string, error) {
	if !IsGitRepository(dir) {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	branches := []string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// GetRemoteBranches returns all remote branches
func GetRemoteBranches(dir string) ([]string, error) {
	if !IsGitRepository(dir) {
		return nil, fmt.Errorf("not a git repository")
	}

	// Fetch latest from remote
	fetchCmd := exec.Command("git", "fetch", "--all")
	fetchCmd.Dir = dir
	_ = fetchCmd.Run() // Ignore errors, we'll try to get branches anyway

	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote branches: %w", err)
	}

	branches := []string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "HEAD") {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func HasUncommittedChanges(dir string) (bool, error) {
	if !IsGitRepository(dir) {
		return false, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// SwitchBranch switches to a different branch
func SwitchBranch(dir, branch string) error {
	if !IsGitRepository(dir) {
		return fmt.Errorf("not a git repository")
	}

	// Check for uncommitted changes
	hasChanges, err := HasUncommittedChanges(dir)
	if err != nil {
		return err
	}

	if hasChanges {
		return fmt.Errorf("cannot switch branches: you have uncommitted changes. Please commit or stash them first")
	}

	// Check if it's a remote branch that doesn't exist locally
	if strings.HasPrefix(branch, "origin/") {
		localBranch := strings.TrimPrefix(branch, "origin/")

		// Check if local branch exists
		branches, err := GetBranches(dir)
		if err != nil {
			return err
		}

		branchExists := false
		for _, b := range branches {
			if b == localBranch {
				branchExists = true
				break
			}
		}

		if !branchExists {
			// Create and track the remote branch
			cmd := exec.Command("git", "checkout", "-b", localBranch, branch)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to checkout remote branch: %s", string(output))
			}
			return nil
		}

		branch = localBranch
	}

	// Switch to the branch
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch branch: %s", string(output))
	}

	return nil
}

// GetGitInfo returns comprehensive git information for a directory
func GetGitInfo(dir string) (*GitInfo, error) {
	info := &GitInfo{
		IsGitRepo: IsGitRepository(dir),
	}

	if !info.IsGitRepo {
		return info, nil
	}

	// Get current branch
	currentBranch, err := GetCurrentBranch(dir)
	if err == nil {
		info.CurrentBranch = currentBranch
	}

	// Get all branches (local + remote)
	localBranches, _ := GetBranches(dir)
	remoteBranches, _ := GetRemoteBranches(dir)

	// Combine and deduplicate
	branchMap := make(map[string]bool)
	for _, b := range localBranches {
		branchMap[b] = true
	}
	for _, b := range remoteBranches {
		branchMap[b] = true
	}

	branches := []string{}
	for b := range branchMap {
		branches = append(branches, b)
	}
	info.Branches = branches

	// Check for uncommitted changes
	hasChanges, err := HasUncommittedChanges(dir)
	if err == nil {
		info.HasUncommitted = hasChanges
	}

	return info, nil
}
