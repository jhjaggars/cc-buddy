package commands

import (
	"context"
	"fmt"

	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// ExecCommand handles executing commands in running environments
type ExecCommand struct {
	envManager *environment.Manager
}

// NewExecCommand creates a new exec command
func NewExecCommand(envManager *environment.Manager) *ExecCommand {
	return &ExecCommand{envManager: envManager}
}

// Execute runs the exec command
func (c *ExecCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cc-buddy exec <environment-name> -- <command> [args...]")
	}

	// Find the separator "--"
	separatorIndex := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIndex = i
			break
		}
	}

	if separatorIndex == -1 {
		return fmt.Errorf("usage: cc-buddy exec <environment-name> -- <command> [args...]\nThe '--' separator is required to separate environment name from command")
	}

	if separatorIndex == 0 {
		return fmt.Errorf("environment name is required before '--'")
	}

	if separatorIndex == len(args)-1 {
		return fmt.Errorf("command is required after '--'")
	}

	// Parse environment name and command
	envName := args[0]
	if separatorIndex > 1 {
		return fmt.Errorf("only one environment name is allowed before '--'")
	}

	command := args[separatorIndex+1:]

	// Execute the command
	if err := c.envManager.ExecuteCommand(ctx, envName, command, true); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// ExecuteNonInteractive executes a command without TTY/interactive mode
func (c *ExecCommand) ExecuteNonInteractive(ctx context.Context, envName string, command []string) error {
	return c.envManager.ExecuteCommand(ctx, envName, command, false)
}