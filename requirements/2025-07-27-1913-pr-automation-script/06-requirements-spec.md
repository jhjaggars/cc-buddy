# Requirements Specification: PR Automation Script

## Problem Statement

Developers need an automated way to create isolated development environments for working on pull requests using Claude Code. The current manual process of creating git worktrees, building containers, and managing development environments is time-consuming and error-prone.

## Solution Overview

A bash script (`cc-buddy`) that automates:
- Git worktree creation for specified branches
- Development container building and startup 
- Environment state tracking and management
- Clean resource cleanup

## Functional Requirements

### Core Commands

#### `cc-buddy create <branch-name>`
- **Automatic branch creation**: Create new git branch from current HEAD if branch doesn't exist locally or remotely
- Create git worktree for specified branch (local, remote, or newly created)
- Build development container from Containerfile.dev or Dockerfile.dev
- Start container in detached mode with worktree mounted at /workspace
- Pass GitHub token via environment variable
- Store environment metadata in `.cc-buddy/` directory
- Generate automatic environment name: `{repo-name}-{branch-name}`
- **Duplicate prevention**: Prevent creation if environment already exists for branch

#### `cc-buddy list`
- Display all active environments with status information
- Show: environment name, branch, container status, worktree path

#### `cc-buddy delete <environment-name>`
- Stop and remove container
- Remove associated volumes
- Remove git worktree
- Clean up environment metadata

### Configuration Options

#### Worktree Location
- **Command line argument**: `--worktree-dir <path>`
- **Environment variable**: `GIT_WORKTREES_DIR`
- **Default**: `~/.worktrees`

#### Container File
- **Command line argument**: `--containerfile <path>`
- **Default search order**: `Containerfile.dev`, `Dockerfile.dev`
- **Error handling**: Exit with helpful message if no dev containerfile found

#### Container Runtime
- **Auto-detection**: Check for `podman` first, fall back to `docker`
- **Override**: `--runtime <docker|podman>`

#### Port Exposure
- **Command line argument**: `--expose-all`
- **Default**: No ports published to host
- **Behavior**: When flag provided, publish all EXPOSE directives from containerfile

## Technical Requirements

### Directory Structure
```
{repo-root}/
├── .cc-buddy/
│   ├── environments.json    # Environment state tracking
│   └── config.json         # Local configuration
└── Containerfile.dev       # Development container definition
```

### Environment State Tracking
Store in `.cc-buddy/environments.json`:
```json
{
  "environments": [
    {
      "name": "repo-feature-branch",
      "branch": "feature-branch", 
      "worktree_path": "/home/user/.worktrees/repo-feature-branch",
      "container_id": "abc123",
      "container_name": "cc-buddy-repo-feature-branch",
      "volume_name": "cc-buddy-repo-feature-branch-data",
      "created": "2025-07-27T19:13:00Z",
      "status": "running"
    }
  ]
}
```

### Container Configuration
- **Working directory**: `/workspace`
- **Volume mount**: `{worktree_path}:/workspace`
- **Environment variables**: 
  - `GITHUB_TOKEN` (from host environment)
  - Additional variables set manually as needed
- **Networking**: Publish ports only when `--expose-all` flag provided
- **Persistence**: Named volume for container data persistence

### Git Integration
- **Repository detection**: Auto-detect from `git remote get-url origin`
- **Branch handling**: Support local, remote, and automatic creation of new branches
- **Branch validation**: Check for branch existence and create if needed from current HEAD
- **Remote branch support**: Handle `origin/branch-name` format without attempting to create
- **Worktree cleanup**: Remove worktree when environment deleted

## Implementation Patterns

### Error Handling
- Validate git repository exists
- Check for required containerfile
- Verify container runtime availability
- Prevent duplicate environments for same branch
- **Branch validation**: Verify remote branches exist before attempting worktree creation
- **Partial failure recovery**: Automatically cleanup any created resources on failure
- Graceful failure with detailed, actionable error messages

### Resource Naming Convention
- **Environment**: `{repo-name}-{branch-name}`
- **Container**: `cc-buddy-{repo-name}-{branch-name}`
- **Volume**: `cc-buddy-{repo-name}-{branch-name}-data`
- **Worktree**: `{worktree-dir}/{repo-name}-{branch-name}`

### State Management
- Atomic operations for environment creation/deletion
- Orphan resource detection and cleanup
- Consistent state between filesystem and tracking data

## Acceptance Criteria

1. **Environment Creation**
   - ✅ Creates worktree in configurable location
   - ✅ Builds container from dev containerfile
   - ✅ Starts container with proper mounts and environment
   - ✅ Records environment state
   - ✅ **Automatic branch creation**: Creates new git branch from current HEAD when specified branch doesn't exist

2. **Environment Management**
   - ✅ Lists all active environments with status
   - ✅ Cleanly removes all resources on delete
   - ✅ Handles multiple repositories without conflicts

3. **Configuration Flexibility**
   - ✅ Supports both Docker and Podman
   - ✅ Configurable worktree location via args/env
   - ✅ Custom containerfile path support

4. **Error Handling**
   - ✅ Clear error messages for missing requirements
   - ✅ Graceful handling of missing containerfiles
   - ✅ Prevention of duplicate environments for same branch
   - ✅ Automatic cleanup of partial resources on creation failure
   - ✅ Detailed error reporting with failure context
   - ✅ **Branch validation**: Proper error handling for invalid remote branch references

## Assumptions

- Users have Git, Docker/Podman, and bash available
- GitHub token is available in environment when needed
- Development containerfiles define necessary ports and dependencies
- Users manage their own branch synchronization
- Container attachment handled via standard Docker/Podman commands