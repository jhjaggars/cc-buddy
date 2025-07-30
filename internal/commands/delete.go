package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// DeleteCommand handles environment deletion
type DeleteCommand struct {
	envManager *environment.Manager
}

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(envManager *environment.Manager) *DeleteCommand {
	return &DeleteCommand{envManager: envManager}
}

// Execute runs the delete command
func (c *DeleteCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cc-buddy delete <environment-name>")
	}

	envName := args[0]

	// Check if environment exists
	env, err := c.envManager.GetConfig().GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	// Show what will be deleted
	fmt.Printf("Environment Details:\n")
	fmt.Printf("  Name: %s\n", env.Name)
	fmt.Printf("  Branch: %s\n", env.Branch)
	fmt.Printf("  Worktree: %s\n", env.WorktreePath)
	fmt.Printf("  Container: %s\n", env.ContainerName)
	fmt.Printf("  Volume: %s\n", env.VolumeName)
	fmt.Printf("  Status: %s\n", env.Status)
	fmt.Println()

	// Confirmation prompt
	fmt.Printf("⚠️  This will permanently delete the environment and all associated resources.\n")
	fmt.Printf("Are you sure you want to delete '%s'? [y/N]: ", envName)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Perform deletion
	fmt.Printf("Deleting environment '%s'...\n", envName)
	
	if err := c.envManager.DeleteEnvironment(ctx, envName); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	fmt.Printf("✅ Environment '%s' deleted successfully!\n", envName)
	return nil
}