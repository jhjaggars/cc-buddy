package environment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GitOperations handles git repository operations
type GitOperations struct {
	repoRoot string
}

// NewGitOperations creates a new git operations instance
func NewGitOperations() (*GitOperations, error) {
	repoRoot, err := findGitRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}
	
	return &GitOperations{repoRoot: repoRoot}, nil
}

// findGitRoot finds the root of the git repository
func findGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoName returns the repository name
func (g *GitOperations) GetRepoName() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = g.repoRoot
	out, err := cmd.Output()
	if err != nil {
		// No origin remote, use directory name
		return filepath.Base(g.repoRoot), nil
	}
	
	remoteURL := strings.TrimSpace(string(out))
	return extractRepoName(remoteURL), nil
}

// extractRepoName extracts repository name from various URL formats
func extractRepoName(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Handle different URL formats
	if strings.HasPrefix(url, "git@") {
		// SSH format: git@github.com:user/repo
		parts := strings.Split(url, ":")
		if len(parts) >= 2 {
			path := parts[len(parts)-1]
			return filepath.Base(path)
		}
	} else if strings.Contains(url, "://") {
		// HTTPS format: https://github.com/user/repo
		return filepath.Base(url)
	}
	
	// Fallback: use the last part after /
	return filepath.Base(url)
}

// BranchExists checks if a branch exists locally
func (g *GitOperations) BranchExists(ctx context.Context, branch string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = g.repoRoot
	err := cmd.Run()
	if err != nil {
		// Check if it's just that the branch doesn't exist
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check branch existence: %w", err)
	}
	return true, nil
}

// RemoteBranchExists checks if a branch exists on remote
func (g *GitOperations) RemoteBranchExists(ctx context.Context, remote, branch string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/remotes/"+remote+"/"+branch)
	cmd.Dir = g.repoRoot
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check remote branch existence: %w", err)
	}
	return true, nil
}

// CreateBranch creates a new branch from the current HEAD
func (g *GitOperations) CreateBranch(ctx context.Context, branchName string) error {
	// Validate branch name
	if err := validateBranchName(branchName); err != nil {
		return err
	}
	
	// Check if branch already exists
	exists, err := g.BranchExists(ctx, branchName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("branch %s already exists", branchName)
	}
	
	// Create the branch without checking it out
	cmd := exec.CommandContext(ctx, "git", "branch", branchName)
	cmd.Dir = g.repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}
	
	return nil
}

// CreateWorktree creates a git worktree for the specified branch
func (g *GitOperations) CreateWorktree(ctx context.Context, worktreePath, branchName, remoteBranch string) error {
	// Pre-flight checks
	if err := g.validateWorktreeCreation(ctx, worktreePath, branchName, remoteBranch); err != nil {
		return err
	}
	
	// Ensure the parent directory exists
	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	
	args := []string{"worktree", "add"}
	
	if remoteBranch != "" {
		// Create worktree from remote branch
		args = append(args, worktreePath, remoteBranch)
	} else {
		// Create worktree from existing local branch
		args = append(args, worktreePath, branchName)
	}
	
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.repoRoot
	
	// Capture both stdout and stderr for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		gitOutput := strings.TrimSpace(string(output))
		commandStr := fmt.Sprintf("git %s", strings.Join(args, " "))
		
		// Provide specific error messages based on common git errors
		errorMsg := g.interpretWorktreeError(gitOutput, branchName, remoteBranch, worktreePath)
		
		return fmt.Errorf("failed to create worktree\nCommand: %s\nGit output: %s\nError: %s", 
			commandStr, gitOutput, errorMsg)
	}
	
	return nil
}

// RemoveWorktree removes a git worktree
func (g *GitOperations) RemoveWorktree(ctx context.Context, worktreePath string) error {
	// First remove the worktree directory if it exists
	if _, err := os.Stat(worktreePath); err == nil {
		cmd := exec.CommandContext(ctx, "git", "worktree", "remove", worktreePath)
		cmd.Dir = g.repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
	} else {
		// Worktree directory doesn't exist, try to prune it
		cmd := exec.CommandContext(ctx, "git", "worktree", "prune")
		cmd.Dir = g.repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to prune worktrees: %w", err)
		}
	}
	
	return nil
}

// ListWorktrees returns a list of all worktrees
func (g *GitOperations) ListWorktrees(ctx context.Context) ([]WorktreeInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = g.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	
	return parseWorktreeList(string(out)), nil
}

// WorktreeInfo represents information about a git worktree
type WorktreeInfo struct {
	Path   string
	Branch string
	Commit string
}

// parseWorktreeList parses the output of 'git worktree list --porcelain'
func parseWorktreeList(output string) []WorktreeInfo {
	var worktrees []WorktreeInfo
	var current WorktreeInfo
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}
		
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		}
	}
	
	// Add the last worktree if there's no trailing empty line
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	
	return worktrees
}

// ParseBranchReference parses branch references like "origin/branch-name"
func (g *GitOperations) ParseBranchReference(branchRef string) (remote, branch string, isRemote bool) {
	if strings.Contains(branchRef, "/") {
		parts := strings.SplitN(branchRef, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], true
		}
	}
	return "", branchRef, false
}

// FetchRemote fetches updates from a remote repository
func (g *GitOperations) FetchRemote(ctx context.Context, remote string) error {
	cmd := exec.CommandContext(ctx, "git", "fetch", remote)
	cmd.Dir = g.repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from %s: %w", remote, err)
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch
func (g *GitOperations) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = g.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// validateBranchName validates that a branch name is valid according to git rules
func validateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	
	// Git branch name restrictions
	invalidChars := regexp.MustCompile(`[~^: \t\n\r\f\v\[\]\\?*]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("branch name contains invalid characters")
	}
	
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot start with '-' or end with '.'")
	}
	
	if strings.Contains(name, "..") || strings.Contains(name, "@{") {
		return fmt.Errorf("branch name cannot contain '..' or '@{'")
	}
	
	return nil
}

// GenerateEnvironmentName creates a standardized environment name
func (g *GitOperations) GenerateEnvironmentName(branchName string) (string, error) {
	repoName, err := g.GetRepoName()
	if err != nil {
		return "", err
	}
	
	// Convert forward slashes to hyphens for branch names like "feature/auth"
	safeBranch := strings.ReplaceAll(branchName, "/", "-")
	
	return fmt.Sprintf("%s-%s", repoName, safeBranch), nil
}

// validateWorktreeCreation performs pre-flight checks before creating a worktree
func (g *GitOperations) validateWorktreeCreation(ctx context.Context, worktreePath, branchName, remoteBranch string) error {
	// Check if worktree directory already exists and is not empty
	if stat, err := os.Stat(worktreePath); err == nil {
		if stat.IsDir() {
			entries, err := os.ReadDir(worktreePath)
			if err != nil {
				return fmt.Errorf("cannot check worktree directory: %w", err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("worktree directory already exists and is not empty: %s", worktreePath)
			}
		} else {
			return fmt.Errorf("worktree path exists but is not a directory: %s", worktreePath)
		}
	}
	
	// Check if branch is already checked out in another worktree
	worktrees, err := g.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing worktrees: %w", err)
	}
	
	targetBranch := branchName
	if remoteBranch != "" {
		// For remote branches, the worktree will create a local tracking branch
		targetBranch = branchName
	}
	
	for _, wt := range worktrees {
		if wt.Branch == targetBranch {
			return fmt.Errorf("branch '%s' is already checked out in worktree: %s", targetBranch, wt.Path)
		}
	}
	
	// Check if the branch exists (for local branches)
	if remoteBranch == "" {
		exists, err := g.BranchExists(ctx, branchName)
		if err != nil {
			return fmt.Errorf("failed to check if branch exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("local branch '%s' does not exist\nTip: Use 'origin/%s' to create from remote branch, or create the local branch first", branchName, branchName)
		}
	}
	
	return nil
}

// interpretWorktreeError provides human-readable error messages for common git worktree failures
func (g *GitOperations) interpretWorktreeError(gitOutput, branchName, remoteBranch, worktreePath string) string {
	output := strings.ToLower(gitOutput)
	
	// Common git worktree error patterns
	switch {
	case strings.Contains(output, "already exists"):
		return fmt.Sprintf("Directory already exists: %s", worktreePath)
		
	case strings.Contains(output, "is already checked out"):
		return fmt.Sprintf("Branch '%s' is already checked out in another worktree", branchName)
		
	case strings.Contains(output, "not a valid object name"):
		if remoteBranch != "" {
			return fmt.Sprintf("Remote branch '%s' does not exist", remoteBranch)
		}
		return fmt.Sprintf("Branch '%s' does not exist", branchName)
		
	case strings.Contains(output, "invalid reference"):
		return fmt.Sprintf("Invalid branch name: '%s'", branchName)
		
	case strings.Contains(output, "permission denied"):
		return fmt.Sprintf("Permission denied when creating worktree at: %s", worktreePath)
		
	case strings.Contains(output, "not a git repository"):
		return "Current directory is not a git repository"
		
	case strings.Contains(output, "no such file or directory"):
		return fmt.Sprintf("Cannot create worktree directory: %s", worktreePath)
		
	default:
		// Generic error message with troubleshooting tips
		suggestions := []string{
			fmt.Sprintf("• Check if branch '%s' exists: git branch -a", branchName),
			"• Ensure the branch is not already checked out elsewhere",
			fmt.Sprintf("• Verify write permissions for directory: %s", filepath.Dir(worktreePath)),
		}
		
		if remoteBranch == "" && !strings.Contains(branchName, "/") {
			suggestions = append(suggestions, fmt.Sprintf("• Try using remote branch: cc-buddy create origin/%s", branchName))
		}
		
		return fmt.Sprintf("Git worktree creation failed. Troubleshooting:\n%s", strings.Join(suggestions, "\n"))
	}
}