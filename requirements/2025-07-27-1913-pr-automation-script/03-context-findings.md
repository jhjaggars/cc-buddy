# Context Findings

## Git Worktree Automation Patterns

### Existing Solutions
- **git-worktree-relative**: Bash script enabling git-worktree to use relative paths
- CLI tools exist that "automate git worktree and Docker Compose development workflows"
- Best practice: Use relative paths for worktrees to avoid path dependencies

### Git Worktree Management
- Worktrees should be created in a dedicated directory structure
- Automatic cleanup is essential to prevent disk space issues
- Branch-based naming conventions provide clear organization

## Container Technology Integration

### Docker vs Podman Compatibility
- Podman is a "drop-in replacement for Docker" - same commands work
- Both use OCI compliant images - same Dockerfiles work for both
- Podman has native CLI on Mac/Windows with embedded Linux guest system
- podman-compose available for Docker Compose compatibility

### Development Container Best Practices

#### Volume Mount Strategies
- **Bind mounts**: Best for development workflows, direct file synchronization
- **Named volumes**: Better for persistent data, managed by container runtime
- Mount source code as volume for live editing while building/running in container

#### Container Configuration
- Use containerfile.dev or dockerfile.dev for development-specific builds
- Leverage EXPOSE directives in Dockerfile for port configuration
- Use lifecycle scripts (postCreateCommand, postStartCommand) for setup automation

#### State Management
- Use named volumes for persistent storage between container restarts
- Container data should persist until explicit deletion
- Volume cleanup should be part of environment destruction

## Environment Tracking Requirements

### State Tracking Needs
- Track active worktrees and their associated containers
- Store environment metadata (branch, container ID, volume names)
- Enable listing all active environments
- Support clean deletion of environments and associated resources

### Naming Conventions
- Automatic naming based on branch names or PR numbers
- Consistent naming scheme for containers, volumes, and worktrees
- Avoid naming conflicts across different repositories

## CLI Command Structure

### Core Commands Needed
- `create <branch>`: Create worktree and start development environment
- `list`: Show all active environments with status
- `delete <env-name>`: Clean up environment (container, volume, worktree)

### Environment Variables
- GitHub token passed via environment to container
- Working directory set to /workspace inside container
- Project-specific configurations available through mounted worktree

## Technical Constraints

### File Detection
- Look for containerfile.dev or dockerfile.dev in repository root
- Fall back to standard Dockerfile if dev-specific file not found
- Error handling for missing container configuration

### Resource Management
- Clean up Docker/Podman volumes on environment deletion
- Remove git worktrees when destroying environments
- Handle orphaned resources from interrupted operations

### Integration Points
- Auto-detect GitHub repository URL from git remote
- Support both Docker and Podman container runtimes
- Maintain compatibility with existing development workflows