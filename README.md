# cc-buddy

A Go-powered development environment manager that creates isolated environments using git worktrees and containers, featuring both CLI and interactive TUI modes.

## Overview

cc-buddy automates the creation of development environments for working on different branches or features. Each environment gets its own git worktree and containerized development environment with all your tools pre-configured.

Perfect for:
- Working on multiple branches simultaneously
- Testing changes in isolation
- Quick feature development
- Code reviews and testing

## Quick Start

```bash
# Interactive mode (recommended)
cc-buddy                               # Launch TUI interface

# CLI mode
cc-buddy init                          # Initialize Containerfile.dev
cc-buddy create feature-branch         # Create new environment
cc-buddy list                          # Interactive environment list
cc-buddy terminal myrepo-feature-branch # Open shell in environment
cc-buddy delete myrepo-feature-branch   # Clean up when done
```

## Features

- **Interactive TUI** - Full-screen interface with keyboard navigation built with Charm.sh
- **Isolated git worktrees** - Each environment has its own working directory
- **Containerized development** - Full development environment with your tools
- **Claude Code integration** - Your configuration and credentials automatically mounted
- **Dual interface modes** - Interactive TUI and traditional CLI for scripting
- **Real-time updates** - Live status monitoring and progress indicators
- **Fast operations** - Optimized for quick creation and deletion
- **Shell completions** - Tab completion for bash, zsh, and fish
- **Type-safe operations** - Go's type system ensures reliability

## Usage

```bash
cc-buddy <command> [options]

Commands:
  init                Create Containerfile.dev in current directory
  create <branch>     Create new development environment
  list               List all active environments  
  delete <env-name>  Delete development environment
  terminal <env-name> Open shell in running environment

Options:
  --worktree-dir <path>      Set custom worktree location
  --containerfile <path>     Specify custom containerfile
  --runtime <docker|podman>  Override container runtime
  --expose-all              Publish all container ports
  --terminal, -t            Launch terminal after creation
  --force                   Force overwrite existing files (init only)
```

## Requirements

- Git
- Docker or Podman
- Go 1.24+ (for building from source)

## Installation

### Building from Source
```bash
git clone https://github.com/jhjaggars/cc-buddy
cd cc-buddy

# Using Makefile (recommended)
make build              # Build binary
make install            # Install to GOPATH/bin
make dev                # Development workflow (format, vet, test, build)

# Or build directly with Go
go build -o cc-buddy ./cmd/cc-buddy
```

### Running
```bash
# After building
./cc-buddy --help       # Show help
./cc-buddy              # Launch interactive TUI

# Or run directly with Go (for development)
go run ./cmd/cc-buddy
```

## Container Environment

Each environment includes:
- Your git worktree mounted at `/workspace`
- Claude Code CLI with your configuration
- GitHub CLI (`gh`)
- Node.js and npm
- Standard development tools

The container automatically mounts:
- Your Claude Code configuration and credentials
- Your custom agents and commands
- The main git repository for worktree access

## Environment Naming

Environments are named using the pattern: `{repo-name}-{branch-name}`

Forward slashes in branch names are converted to hyphens for container compatibility.

## Interactive TUI

The interactive Terminal User Interface (TUI) provides:

- **Environment Dashboard**: Overview of all environments with real-time status
- **Keyboard Navigation**: Navigate with arrow keys, select with Enter
- **Built-in Actions**: Terminal access, deletion, refresh - all with single keystrokes
- **Progress Indicators**: Visual feedback during long-running operations
- **Confirmation Dialogs**: Safe deletion with detailed environment information
- **Context-Sensitive Help**: Press `?` for available keyboard shortcuts

### TUI Keyboard Shortcuts

- `↑↓` - Navigate environment list
- `Enter` - Open terminal in selected environment
- `d` - Delete selected environment (with confirmation)
- `r` - Refresh environment list
- `q` / `Ctrl+C` / `Esc` - Quit
- `?` / `h` - Toggle help

### Technology Stack

Built with the [Charm.sh](https://charm.sh) ecosystem:
- **Bubble Tea**: Elm architecture TUI framework
- **Bubbles**: Pre-built UI components (tables, inputs, progress bars)
- **Lip Gloss**: Terminal styling and layouts