package commands

import (
	"context"
	"fmt"

	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// CreateCommand handles environment creation
type CreateCommand struct {
	envManager *environment.Manager
}

// NewCreateCommand creates a new create command
func NewCreateCommand(envManager *environment.Manager) *CreateCommand {
	return &CreateCommand{envManager: envManager}
}

// Execute runs the create command
func (c *CreateCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cc-buddy create <branch-name>")
	}

	branchName := args[0]
	
	// Parse branch reference (handle origin/branch-name format)
	gitOps := c.envManager.GetGitOperations()
	remote, branch, isRemote := gitOps.ParseBranchReference(branchName)
	
	if isRemote {
		fmt.Printf("Creating environment for remote branch %s/%s...\n", remote, branch)
	} else {
		fmt.Printf("Creating environment for branch %s...\n", branch)
	}

	opts := environment.CreateEnvironmentOptions{
		BranchName:     branch,
		IsRemoteBranch: isRemote,
		RemoteName:     remote,
	}

	// Create the environment
	env, err := c.envManager.CreateEnvironment(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	fmt.Printf("✅ Environment '%s' created successfully!\n", env.Name)
	fmt.Printf("   Branch: %s\n", env.Branch)
	fmt.Printf("   Worktree: %s\n", env.WorktreePath)
	fmt.Printf("   Container: %s\n", env.ContainerName)
	fmt.Printf("   Status: %s\n", env.Status)
	fmt.Printf("\nTo access the environment:\n")
	fmt.Printf("   cc-buddy terminal %s\n", env.Name)

	return nil
}