# cc-buddy

A PR automation tool that creates isolated development environments using git worktrees and containers.

## Overview

cc-buddy automates the creation of development environments for working on different branches or features. Each environment gets its own git worktree and containerized development environment with all your tools pre-configured.

Perfect for:
- Working on multiple branches simultaneously
- Testing changes in isolation
- Quick feature development
- Code reviews and testing

## Quick Start

```bash
# Initialize a new project (creates Containerfile.dev)
./cc-buddy init

# Create a new development environment
./cc-buddy create feature-branch

# List all environments
./cc-buddy list

# Open shell in environment
./cc-buddy terminal cc-buddy-feature-branch

# Clean up when done
./cc-buddy delete cc-buddy-feature-branch
```

## Features

- **Isolated git worktrees** - Each environment has its own working directory
- **Containerized development** - Full development environment with your tools
- **Claude Code integration** - Your configuration and credentials automatically mounted
- **Fast operations** - Optimized for quick creation and deletion
- **Shell completions** - Tab completion for bash, zsh, and fish

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
- jq (for JSON processing)

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