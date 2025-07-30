# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `cc-buddy`, a Go application for PR automation that creates isolated development environments using git worktrees and containers. The application features both CLI mode for backward compatibility and an interactive Terminal User Interface (TUI) built with Charm.sh libraries.

## Architecture

### Main Components

- **CLI Entry Point**: `cmd/cc-buddy/main.go` - Main application entry with both CLI and TUI modes
- **Command Layer**: `internal/commands/` - Individual command implementations
  - `create.go` - Environment creation logic
  - `list.go` - Environment listing with TUI and plain text modes
  - `delete.go` - Environment cleanup
  - `init.go` - Containerfile.dev generation
  - `terminal.go` - Container shell access
- **UI Layer**: `internal/ui/` - Bubble Tea TUI components
  - `models/` - TUI models (main, list, create, confirm, help, progress)
  - `components/` - Reusable UI components
- **Core Logic**: 
  - `internal/environment/` - Environment lifecycle management and git operations
  - `internal/container/` - Docker/Podman runtime abstraction
  - `internal/config/` - Configuration and state management
- **Metadata Management**: `.cc-buddy/` directory for state tracking
  - `environments.json` - Tracks active development environments
  - `config.json` - Local configuration storage
- **Requirements Documentation**: `requirements/` directory with detailed specifications

### Key Functionality

The application provides six main commands:
- `init` - Generates Containerfile.dev interactively
- `create <branch-name> [-e "command"]` - Creates isolated development environment with git worktree and container
- `list` - Shows all active environments with status (interactive TUI by default, `--plain` for text output)  
- `delete <environment-name>` - Cleans up environment resources
- `terminal <environment-name>` - Opens shell in running environment
- `exec <environment-name> -- <command>` - Executes arbitrary commands in running environments

### Operational Modes

- **Interactive TUI Mode**: Default when run without arguments, provides full-screen interface with keyboard navigation
- **CLI Mode**: Traditional command-line interface for backward compatibility and scripting

### Environment Naming Convention

All resources follow the pattern: `{repo-name}-{branch-name}` where forward slashes in branch names are converted to hyphens.

### Container Integration

- Supports both Docker and Podman (auto-detects, prefers Podman)
- Uses `Containerfile.dev` or `Dockerfile.dev` for development containers
- Mounts worktree at `/workspace` in container
- Creates named volumes for data persistence
- Passes `GITHUB_TOKEN` environment variable to containers

## Common Commands

### Interactive TUI Mode (Default)
```bash
cc-buddy                                   # Launch interactive interface
```

### CLI Mode Commands
```bash
cc-buddy init                              # Initialize Containerfile.dev
cc-buddy create feature-branch             # Create environment for branch
cc-buddy create feature-branch -e "npm run dev"  # Create with custom startup command
cc-buddy create origin/feature-branch      # Create from remote branch
cc-buddy list                              # Interactive environment list
cc-buddy list --plain                      # Plain text output for scripts
cc-buddy terminal myrepo-feature-branch    # Open shell in environment
cc-buddy exec myrepo-feature-branch -- npm test    # Execute command in environment
cc-buddy exec myrepo-feature-branch -- bash -c "cd /workspace && make build"  # Complex command
cc-buddy delete myrepo-feature-branch      # Delete environment
```

### Command Options
```bash
--worktree-dir <path>          # Set custom worktree location
--containerfile <path>         # Specify custom containerfile
--runtime <docker|podman>      # Override container runtime
--expose-all                   # Publish all container ports
--terminal, -t                 # Launch terminal after creation (create command)
--force                        # Force overwrite existing files (init command)
```

### Development Workflow
```bash
# Interactive mode (recommended)
cc-buddy                                  # Launch TUI, navigate with arrow keys

# CLI mode
cc-buddy create my-feature                # Create environment
cc-buddy create my-feature -e "python -m http.server 8000"  # Create with custom startup
cc-buddy terminal myrepo-my-feature       # Open terminal in environment
cc-buddy exec myrepo-my-feature -- npm test  # Execute commands in environment
cc-buddy delete myrepo-my-feature         # Clean up when done

# Alternative: direct container access
podman exec -it myrepo-my-feature bash    # Direct container access
docker exec -it myrepo-my-feature bash    # Docker alternative
```

## Dependencies

Required system dependencies:
- `git` - Git version control
- `docker` or `podman` - Container runtime
- Go 1.24+ (for building from source)

Optional dependencies:
- `gh` - GitHub CLI (included in containers)

## Technology Stack

- **Language**: Go 1.24+
- **TUI Framework**: Bubble Tea (Elm architecture)
- **UI Components**: Bubbles (tables, inputs, progress bars)
- **Styling**: Lip Gloss (terminal styling and layouts)
- **Configuration**: JSON-based state management

## Error Handling

The application includes comprehensive error handling:
- Type-safe error handling with structured error types
- Validates git repository existence and branch references
- Checks for required containerfiles and dependencies
- Prevents duplicate environments with validation
- Automatic cleanup on partial failures with rollback
- Graceful signal handling (SIGINT, SIGTERM) with user confirmation
- Recovery of orphaned resources on startup

## State Management

Environment state is tracked in `.cc-buddy/environments.json` with the following structure:
```json
{
  "environments": [
    {
      "name": "repo-feature-branch",
      "branch": "feature-branch",
      "worktree_path": "/path/to/worktree",
      "container_id": "abc123",
      "container_name": "cc-buddy-repo-feature-branch",
      "volume_name": "cc-buddy-repo-feature-branch-data",
      "created": "2025-07-27T19:13:00Z",
      "status": "running"
    }
  ]
}
```