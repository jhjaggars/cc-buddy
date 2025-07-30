package commands

import (
	"context"
	"fmt"

	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// TerminalCommand handles opening terminal sessions
type TerminalCommand struct {
	envManager *environment.Manager
}

// NewTerminalCommand creates a new terminal command
func NewTerminalCommand(envManager *environment.Manager) *TerminalCommand {
	return &TerminalCommand{envManager: envManager}
}

// Execute runs the terminal command
func (c *TerminalCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cc-buddy terminal <environment-name>")
	}

	envName := args[0]

	// Check if environment exists
	env, err := c.envManager.GetConfig().GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	fmt.Printf("Opening terminal for environment '%s'...\n", envName)
	fmt.Printf("Container: %s\n", env.ContainerName)
	fmt.Printf("Working directory: /workspace\n")
	fmt.Println()

	// Open terminal
	if err := c.envManager.OpenTerminal(ctx, envName); err != nil {
		return fmt.Errorf("failed to open terminal: %w", err)
	}

	return nil
}