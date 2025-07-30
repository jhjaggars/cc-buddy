package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jhjaggars/cc-buddy/internal/commands"
	"github.com/jhjaggars/cc-buddy/internal/environment"
	"github.com/jhjaggars/cc-buddy/internal/ui/models"
)

func main() {
	if len(os.Args) > 1 {
		// CLI mode for backward compatibility
		if err := handleCLIMode(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// TUI mode
	for {
		mainModel := models.NewMainModel()
		p := tea.NewProgram(mainModel, tea.WithAltScreen())
		
		// Set up signal handling
		mainModel.SetProgram(p)
		
		// Run the TUI
		model, err := p.Run()
		if err != nil {
			log.Fatalf("Error running program: %v", err)
		}
		
		// Check if we need to launch a terminal
		finalModel := model.(*models.MainModel)
		terminalEnv := finalModel.GetTerminalEnvironment()
		finalModel.Cleanup()
		
		if terminalEnv != "" {
			// Launch terminal and restart TUI when done
			if err := launchTerminal(terminalEnv); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening terminal: %v\n", err)
				fmt.Println("Press Enter to continue...")
				fmt.Scanln()
			}
			// Continue the loop to restart TUI
		} else {
			// Normal quit, exit the loop
			break
		}
	}
}

func handleCLIMode(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: cc-buddy [command] [args...]")
		fmt.Println("Commands: init, create, list, delete, terminal, exec")
		fmt.Println("Run without arguments for interactive mode")
		return nil
	}

	ctx := context.Background()
	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "init":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		initCmd := commands.NewInitCommand(envManager)
		return initCmd.Execute(ctx, commandArgs)

	case "create":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		createCmd := commands.NewCreateCommand(envManager)
		return createCmd.Execute(ctx, commandArgs)

	case "list":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		listCmd := commands.NewListCommand(envManager)
		return listCmd.Execute(ctx, commandArgs)

	case "delete":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		deleteCmd := commands.NewDeleteCommand(envManager)
		return deleteCmd.Execute(ctx, commandArgs)

	case "terminal":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		terminalCmd := commands.NewTerminalCommand(envManager)
		return terminalCmd.Execute(ctx, commandArgs)

	case "exec":
		envManager, err := environment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		execCmd := commands.NewExecCommand(envManager)
		return execCmd.Execute(ctx, commandArgs)

	case "help", "-h", "--help":
		printHelp()
		return nil

	default:
		return fmt.Errorf("unknown command: %s\nRun 'cc-buddy help' for usage information", command)
	}
}

// launchTerminal opens a terminal for the specified environment
func launchTerminal(envName string) error {
	ctx := context.Background()
	envManager, err := environment.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	// Check if environment exists
	env, err := envManager.GetConfig().GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	fmt.Printf("Opening terminal for environment '%s'...\n", envName)
	fmt.Printf("Container: %s\n", env.ContainerName)
	fmt.Printf("Working directory: /workspace\n")
	fmt.Println()

	// Open terminal
	if err := envManager.OpenTerminal(ctx, envName); err != nil {
		return fmt.Errorf("failed to open terminal: %w", err)
	}

	return nil
}

func printHelp() {
	fmt.Println("cc-buddy - Development Environment Manager")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    cc-buddy [command] [args...]")
	fmt.Println("    cc-buddy                    # Interactive TUI mode")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("    init                        Generate Containerfile.dev interactively")
	fmt.Println("    create <branch-name> [-e \"cmd\"] Create new development environment")
	fmt.Println("    list [--plain]              Interactive environment list (--plain for text)")
	fmt.Println("    delete <env-name>           Delete an environment")
	fmt.Println("    terminal <env-name>         Open terminal in environment")
	fmt.Println("    exec <env-name> -- <command> Execute command in environment")
	fmt.Println("    help                        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    cc-buddy init")
	fmt.Println("    cc-buddy create feature-auth")
	fmt.Println("    cc-buddy create feature-auth -e \"npm run dev\"")
	fmt.Println("    cc-buddy create origin/main")
	fmt.Println("    cc-buddy list                      # Interactive list with navigation")
	fmt.Println("    cc-buddy list --plain              # Plain text output for scripts") 
	fmt.Println("    cc-buddy terminal myrepo-feature-auth")
	fmt.Println("    cc-buddy exec myrepo-feature-auth -- npm test")
	fmt.Println("    cc-buddy exec myrepo-feature-auth -- bash -c \"cd /workspace && make build\"")
	fmt.Println("    cc-buddy delete myrepo-feature-auth")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/jhjaggars/cc-buddy")
}