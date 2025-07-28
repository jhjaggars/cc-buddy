# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `cc-buddy`, a bash script for PR automation that creates isolated development environments using git worktrees and containers. The script automates the creation of git worktrees, building development containers, and managing environment state.

## Architecture

### Main Components

- **Core Script**: `cc-buddy` - Single bash script (591 lines) containing all functionality
- **Metadata Management**: `.cc-buddy/` directory for state tracking
  - `environments.json` - Tracks active development environments
  - `config.json` - Local configuration storage
- **Requirements Documentation**: `requirements/` directory with detailed specifications

### Key Functionality

The script provides three main commands:
- `create <branch-name>` - Creates isolated development environment with git worktree and container
- `list` - Shows all active environments with status
- `delete <environment-name>` - Cleans up environment resources

### Environment Naming Convention

All resources follow the pattern: `{repo-name}-{branch-name}` where forward slashes in branch names are converted to hyphens.

### Container Integration

- Supports both Docker and Podman (auto-detects, prefers Podman)
- Uses `Containerfile.dev` or `Dockerfile.dev` for development containers
- Mounts worktree at `/workspace` in container
- Creates named volumes for data persistence
- Passes `GITHUB_TOKEN` environment variable to containers

## Common Commands

### Running the Script
```bash
./cc-buddy create feature-branch           # Create environment for branch
./cc-buddy create origin/feature-branch    # Create from remote branch
./cc-buddy list                            # List all environments
./cc-buddy delete repo-feature-branch      # Delete environment
```

### Script Options
```bash
--worktree-dir <path>          # Set custom worktree location
--containerfile <path>         # Specify custom containerfile
--runtime <docker|podman>      # Override container runtime
--expose-all                   # Publish all container ports
```

### Development Workflow
```bash
# Create environment
./cc-buddy create my-feature

# Attach to container for development
podman exec -it cc-buddy-repo-my-feature bash
# or
docker exec -it cc-buddy-repo-my-feature bash

# Clean up when done
./cc-buddy delete repo-my-feature
```

## Dependencies

Required system dependencies:
- `git` - Git version control
- `jq` - JSON processing
- `docker` or `podman` - Container runtime

## Error Handling

The script includes comprehensive error handling:
- Validates git repository existence
- Checks for required containerfiles
- Prevents duplicate environments
- Automatic cleanup on partial failures
- Branch validation for remote references

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