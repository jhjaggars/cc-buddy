package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// InitCommand handles Containerfile.dev generation
type InitCommand struct {
	envManager *environment.Manager
}

// NewInitCommand creates a new init command
func NewInitCommand(envManager *environment.Manager) *InitCommand {
	return &InitCommand{envManager: envManager}
}

// Execute runs the init command
func (c *InitCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println("🐋 cc-buddy Containerfile.dev Generator")
	fmt.Println("=====================================")
	fmt.Println()

	// Check if Containerfile.dev already exists
	containerfilePath := "Containerfile.dev"
	if _, err := os.Stat(containerfilePath); err == nil {
		fmt.Printf("⚠️  %s already exists.\n", containerfilePath)
		if !c.confirmOverwrite() {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	// Interactive prompts
	baseImage := c.promptForBaseImage()
	packages := c.promptForPackages()
	ports := c.promptForPorts()
	volumes := c.promptForVolumes()
	envVars := c.promptForEnvVars()
	commands := c.promptForCommands()

	// Generate Containerfile content
	content := c.generateContainerfile(baseImage, packages, ports, volumes, envVars, commands)

	// Write to file
	if err := os.WriteFile(containerfilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", containerfilePath, err)
	}

	fmt.Printf("✅ %s created successfully!\n", containerfilePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review and customize %s\n", containerfilePath)
	fmt.Println("  2. Create your first environment:")
	fmt.Println("     cc-buddy create <branch-name>")

	return nil
}

func (c *InitCommand) confirmOverwrite() bool {
	fmt.Print("Do you want to overwrite it? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func (c *InitCommand) promptForBaseImage() string {
	fmt.Println("1. Base Image Selection")
	fmt.Println("   Choose a base image for your development environment:")
	fmt.Println("   1) ubuntu:22.04 (recommended)")
	fmt.Println("   2) node:18")
	fmt.Println("   3) python:3.11")
	fmt.Println("   4) golang:1.21")
	fmt.Println("   5) rust:1.70")
	fmt.Println("   6) Custom")
	fmt.Print("   Enter choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	switch response {
	case "", "1":
		return "ubuntu:22.04"
	case "2":
		return "node:18"
	case "3":
		return "python:3.11"
	case "4":
		return "golang:1.21"
	case "5":
		return "rust:1.70"
	case "6":
		fmt.Print("   Enter custom base image: ")
		custom, _ := reader.ReadString('\n')
		return strings.TrimSpace(custom)
	default:
		return "ubuntu:22.04"
	}
}

func (c *InitCommand) promptForPackages() []string {
	fmt.Println()
	fmt.Println("2. System Packages")
	fmt.Print("   Enter additional packages to install (space-separated) [git curl wget]: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	if response == "" {
		response = "git curl wget"
	}
	
	return strings.Fields(response)
}

func (c *InitCommand) promptForPorts() []string {
	fmt.Println()
	fmt.Println("3. Port Exposure")
	fmt.Print("   Enter ports to expose (space-separated) []: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	if response == "" {
		return []string{}
	}
	
	return strings.Fields(response)
}

func (c *InitCommand) promptForVolumes() []string {
	fmt.Println()
	fmt.Println("4. Volume Mounts")
	fmt.Print("   Enter additional volume mount points (space-separated) []: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	if response == "" {
		return []string{}
	}
	
	return strings.Fields(response)
}

func (c *InitCommand) promptForEnvVars() []string {
	fmt.Println()
	fmt.Println("5. Environment Variables")
	fmt.Print("   Enter environment variables (KEY=value, space-separated) []: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	if response == "" {
		return []string{}
	}
	
	return strings.Fields(response)
}

func (c *InitCommand) promptForCommands() []string {
	fmt.Println()
	fmt.Println("6. Startup Commands")
	fmt.Print("   Enter commands to run on container start (one per line, empty line to finish):\n")
	
	var commands []string
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("   > ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		
		if line == "" {
			break
		}
		
		commands = append(commands, line)
	}
	
	return commands
}

func (c *InitCommand) generateContainerfile(baseImage string, packages, ports, volumes, envVars, commands []string) string {
	var content strings.Builder
	
	content.WriteString("# Development Container for cc-buddy\n")
	content.WriteString("# Generated automatically - feel free to customize!\n\n")
	
	// Base image
	content.WriteString(fmt.Sprintf("FROM %s\n\n", baseImage))
	
	// System packages
	if len(packages) > 0 {
		content.WriteString("# Install system packages\n")
		content.WriteString("RUN apt-get update && apt-get install -y \\\n")
		for i, pkg := range packages {
			if i == len(packages)-1 {
				content.WriteString(fmt.Sprintf("    %s \\\n", pkg))
			} else {
				content.WriteString(fmt.Sprintf("    %s \\\n", pkg))
			}
		}
		content.WriteString("    && rm -rf /var/lib/apt/lists/*\n\n")
	}
	
	// Environment variables
	if len(envVars) > 0 {
		content.WriteString("# Environment variables\n")
		for _, env := range envVars {
			content.WriteString(fmt.Sprintf("ENV %s\n", env))
		}
		content.WriteString("\n")
	}
	
	// Expose ports
	if len(ports) > 0 {
		content.WriteString("# Expose ports\n")
		for _, port := range ports {
			content.WriteString(fmt.Sprintf("EXPOSE %s\n", port))
		}
		content.WriteString("\n")
	}
	
	// Volume mount points
	if len(volumes) > 0 {
		content.WriteString("# Volume mount points\n")
		for _, volume := range volumes {
			content.WriteString(fmt.Sprintf("VOLUME %s\n", volume))
		}
		content.WriteString("\n")
	}
	
	// Working directory
	content.WriteString("# Set working directory\n")
	content.WriteString("WORKDIR /workspace\n\n")
	
	// Startup commands
	if len(commands) > 0 {
		content.WriteString("# Startup commands\n")
		for _, cmd := range commands {
			content.WriteString(fmt.Sprintf("RUN %s\n", cmd))
		}
		content.WriteString("\n")
	}
	
	// Default command
	content.WriteString("# Keep container running\n")
	content.WriteString("CMD [\"tail\", \"-f\", \"/dev/null\"]\n")
	
	return content.String()
}