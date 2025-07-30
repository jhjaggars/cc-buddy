package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jhjaggars/cc-buddy/internal/environment"
	"github.com/jhjaggars/cc-buddy/internal/ui/models"
)

// ListCommand handles environment listing
type ListCommand struct {
	envManager *environment.Manager
}

// NewListCommand creates a new list command
func NewListCommand(envManager *environment.Manager) *ListCommand {
	return &ListCommand{envManager: envManager}
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context, args []string) error {
	// Check for --plain flag for backward compatibility
	usePlainOutput := false
	for _, arg := range args {
		if arg == "--plain" {
			usePlainOutput = true
			break
		}
	}

	if usePlainOutput {
		return c.executePlainList(ctx)
	}

	// Launch interactive TUI list
	return c.executeInteractiveList()
}

// executeInteractiveList launches the interactive Bubble Tea list interface
func (c *ListCommand) executeInteractiveList() error {
	listModel, err := models.NewStandaloneListModel()
	if err != nil {
		return fmt.Errorf("failed to initialize list interface: %w", err)
	}

	p := tea.NewProgram(listModel, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("failed to run list interface: %w", err)
	}

	return nil
}

// executePlainList provides the original plain text output for scripts
func (c *ListCommand) executePlainList(ctx context.Context) error {
	environments, err := c.envManager.ListEnvironments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	if len(environments) == 0 {
		fmt.Println("No environments found.")
		fmt.Println("\nCreate your first environment with:")
		fmt.Println("  cc-buddy create <branch-name>")
		return nil
	}

	fmt.Printf("Environments (%d):\n\n", len(environments))

	// Print header
	fmt.Printf("%-25s %-20s %-10s %-15s\n", "NAME", "BRANCH", "STATUS", "CREATED")
	fmt.Printf("%s\n", strings.Repeat("-", 70))

	// Print environments
	for _, env := range environments {
		status := getStatusDisplay(env.Status)
		created := formatTimeAgo(env.Created)
		
		fmt.Printf("%-25s %-20s %-10s %-15s\n", 
			env.Name, 
			env.Branch, 
			status, 
			created)
	}

	fmt.Printf("\nCommands:\n")
	fmt.Printf("  cc-buddy terminal <name>  - Open terminal in environment\n")
	fmt.Printf("  cc-buddy delete <name>    - Delete environment\n")

	return nil
}

// getStatusDisplay returns a user-friendly status display
func getStatusDisplay(status string) string {
	switch status {
	case "running":
		return "🟢 running"
	case "stopped":
		return "🟡 stopped"
	case "creating":
		return "🔄 creating"
	case "error":
		return "🔴 error"
	default:
		return status
	}
}

// formatTimeAgo formats a time as "2h ago", "1d ago", etc.
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
	
	// For older dates, show the actual date
	return t.Format("Jan 2")
}