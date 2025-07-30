package commands

import (
	"context"
	"fmt"
	"strings"

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
		return fmt.Errorf("usage: cc-buddy create <branch-name> [-e \"command\"]")
	}

	// Parse arguments
	var branchName string
	var startupCommand []string
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		if arg == "-e" {
			// Next argument should be the command
			if i+1 >= len(args) {
				return fmt.Errorf("-e flag requires a command argument")
			}
			i++
			commandStr := args[i]
			// Parse command string into arguments using shell-like splitting
			startupCommand = parseCommand(commandStr)
		} else if branchName == "" {
			branchName = arg
		} else {
			return fmt.Errorf("unexpected argument: %s", arg)
		}
		i++
	}
	
	if branchName == "" {
		return fmt.Errorf("branch name is required")
	}
	
	// Parse branch reference (handle origin/branch-name format)
	gitOps := c.envManager.GetGitOperations()
	remote, branch, isRemote := gitOps.ParseBranchReference(branchName)
	
	if isRemote {
		fmt.Printf("Creating environment for remote branch %s/%s...\n", remote, branch)
	} else {
		fmt.Printf("Creating environment for branch %s...\n", branch)
	}
	
	if len(startupCommand) > 0 {
		fmt.Printf("Custom startup command: %s\n", strings.Join(startupCommand, " "))
	}

	opts := environment.CreateEnvironmentOptions{
		BranchName:     branch,
		IsRemoteBranch: isRemote,
		RemoteName:     remote,
		StartupCommand: startupCommand,
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

// parseCommand parses a command string into arguments
// Simple implementation that splits on spaces, respecting quoted strings
func parseCommand(commandStr string) []string {
	if commandStr == "" {
		return nil
	}
	
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	
	for i := 0; i < len(commandStr); i++ {
		char := commandStr[i]
		
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteByte(char)
			}
		case ' ', '\t':
			if inQuotes {
				current.WriteByte(char)
			} else {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteByte(char)
		}
	}
	
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	
	return args
}