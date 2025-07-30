package environment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jhjaggars/cc-buddy/internal/config"
	"github.com/jhjaggars/cc-buddy/internal/container"
	"github.com/jhjaggars/cc-buddy/internal/system"
)

// Manager orchestrates environment creation, management, and cleanup
type Manager struct {
	configMgr     *config.Manager
	containerMgr  *container.Manager
	gitOps        *GitOperations
}

// NewManager creates a new environment manager
func NewManager() (*Manager, error) {
	configMgr, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}
	
	// Load existing configuration
	if err := configMgr.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	if err := configMgr.LoadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	
	// Initialize container manager based on config
	var containerMgr *container.Manager
	cfg := configMgr.GetConfig()
	if cfg.Runtime == "auto" {
		containerMgr, err = container.NewManager()
	} else {
		containerMgr, err = container.NewManagerWithRuntime(cfg.Runtime)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create container manager: %w", err)
	}
	
	// Initialize git operations
	gitOps, err := NewGitOperations()
	if err != nil {
		return nil, fmt.Errorf("failed to create git operations: %w", err)
	}
	
	return &Manager{
		configMgr:    configMgr,
		containerMgr: containerMgr,
		gitOps:       gitOps,
	}, nil
}

// CreateEnvironmentOptions holds options for environment creation
type CreateEnvironmentOptions struct {
	BranchName      string
	IsRemoteBranch  bool
	RemoteName      string
	WorktreeDir     string
	Containerfile   string
	ExposeAllPorts  bool
	StartupCommand  []string
}

// CreateEnvironment creates a new development environment
func (m *Manager) CreateEnvironment(ctx context.Context, opts CreateEnvironmentOptions) (retEnv *config.Environment, retErr error) {
	// Generate environment name
	envName, err := m.gitOps.GenerateEnvironmentName(opts.BranchName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate environment name: %w", err)
	}
	
	// Check if environment already exists
	if _, err := m.configMgr.GetEnvironment(envName); err == nil {
		return nil, fmt.Errorf("environment %s already exists", envName)
	}
	
	// Set up default options
	if opts.WorktreeDir == "" {
		opts.WorktreeDir = m.configMgr.GetConfig().WorktreeDir
	}
	if opts.Containerfile == "" {
		opts.Containerfile = m.configMgr.GetConfig().Containerfile
	}
	
	// Create worktree path
	worktreePath := filepath.Join(opts.WorktreeDir, envName)
	
	// Track resources for cleanup
	type cleanupState struct {
		environmentInState bool
		branchCreated     bool
		worktreeCreated   bool
		imageBuilt        bool
		volumeCreated     bool
		containerStarted  bool
		imageName         string
	}
	
	cleanup := &cleanupState{}
	
	// Create the environment step by step
	env := &config.Environment{
		Name:          envName,
		Branch:        opts.BranchName,
		WorktreePath:  worktreePath,
		ContainerName: fmt.Sprintf("cc-buddy-%s", envName),
		VolumeName:    fmt.Sprintf("cc-buddy-%s-data", envName),
		Created:       time.Now(),
		Status:        "creating",
	}
	
	// Enhanced cleanup on failure - preserves original error
	defer func() {
		if retErr != nil {
			// Perform granular cleanup in reverse order of creation
			if cleanup.containerStarted && env.ContainerID != "" {
				if stopErr := m.containerMgr.GetRuntime().Stop(ctx, env.ContainerID); stopErr != nil {
					// Log but don't override original error
					fmt.Printf("Warning: Failed to stop container during cleanup: %v\n", stopErr)
				}
				if removeErr := m.containerMgr.GetRuntime().Remove(ctx, env.ContainerID); removeErr != nil {
					fmt.Printf("Warning: Failed to remove container during cleanup: %v\n", removeErr)
				}
			}
			
			if cleanup.volumeCreated {
				if removeErr := m.containerMgr.GetRuntime().RemoveVolume(ctx, env.VolumeName); removeErr != nil {
					fmt.Printf("Warning: Failed to remove volume during cleanup: %v\n", removeErr)
				}
			}
			
			if cleanup.imageBuilt && cleanup.imageName != "" {
				if removeErr := m.containerMgr.GetRuntime().RemoveImage(ctx, cleanup.imageName); removeErr != nil {
					// Image removal might fail if container still exists, that's okay
					fmt.Printf("Warning: Failed to remove image during cleanup: %v\n", removeErr)
				}
			}
			
			if cleanup.worktreeCreated {
				if removeErr := m.gitOps.RemoveWorktree(ctx, worktreePath); removeErr != nil {
					fmt.Printf("Warning: Failed to remove worktree during cleanup: %v\n", removeErr)
				}
			}
			
			if cleanup.branchCreated {
				// Only remove branch if we created it (not if it already existed)
				if deleteErr := m.gitOps.DeleteBranch(ctx, opts.BranchName); deleteErr != nil {
					fmt.Printf("Warning: Failed to remove created branch during cleanup: %v\n", deleteErr)
				}
			}
			
			if cleanup.environmentInState {
				if removeErr := m.configMgr.RemoveEnvironment(envName); removeErr != nil {
					fmt.Printf("Warning: Failed to remove environment from state during cleanup: %v\n", removeErr)
				}
			}
		}
	}()
	
	// Step 1: Handle branch creation/validation
	if opts.IsRemoteBranch {
		// Fetch remote updates first
		if err := m.gitOps.FetchRemote(ctx, opts.RemoteName); err != nil {
			return nil, fmt.Errorf("failed to fetch remote %s: %w", opts.RemoteName, err)
		}
		
		// Check if remote branch exists
		exists, err := m.gitOps.RemoteBranchExists(ctx, opts.RemoteName, opts.BranchName)
		if err != nil {
			return nil, fmt.Errorf("failed to check remote branch: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("remote branch %s/%s does not exist", opts.RemoteName, opts.BranchName)
		}
	} else {
		// Check if local branch exists
		exists, err := m.gitOps.BranchExists(ctx, opts.BranchName)
		if err != nil {
			return nil, fmt.Errorf("failed to check local branch: %w", err)
		}
		if !exists {
			// Create new branch
			if err := m.gitOps.CreateBranch(ctx, opts.BranchName); err != nil {
				return nil, fmt.Errorf("failed to create branch: %w", err)
			}
			cleanup.branchCreated = true
		}
	}
	
	// Step 2: Create git worktree
	var remoteBranch string
	if opts.IsRemoteBranch {
		remoteBranch = fmt.Sprintf("%s/%s", opts.RemoteName, opts.BranchName)
	}
	
	if err := m.gitOps.CreateWorktree(ctx, worktreePath, opts.BranchName, remoteBranch); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}
	cleanup.worktreeCreated = true
	
	// Step 3: Check for containerfile
	containerfilePath := filepath.Join(worktreePath, opts.Containerfile)
	if _, err := os.Stat(containerfilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("containerfile not found: %s", containerfilePath)
	}
	
	// Step 4: Build container image with user sync
	imageTag := fmt.Sprintf("cc-buddy-%s:latest", envName)
	
	// Get host user information for user ID synchronization
	userInfo := system.GetUserInfoWithFallback()
	
	buildOpts := container.BuildOptions{
		Context:    worktreePath,
		Dockerfile: opts.Containerfile,
		Tags:       []string{imageTag},
		BuildArgs: map[string]string{
			"USER_UID": strconv.Itoa(userInfo.UID),
			"USER_GID": strconv.Itoa(userInfo.GID),
		},
	}
	
	if err := m.containerMgr.GetRuntime().Build(ctx, buildOpts); err != nil {
		return nil, fmt.Errorf("failed to build container image: %w", err)
	}
	cleanup.imageBuilt = true
	cleanup.imageName = imageTag
	
	// Step 5: Create named volume
	if err := m.containerMgr.GetRuntime().CreateVolume(ctx, env.VolumeName); err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	cleanup.volumeCreated = true
	
	// Step 6: Start container
	mounts := []container.Mount{
		{
			Type:   "bind",
			Source: worktreePath,
			Target: "/workspace",
			Options: []string{"Z"}, // SELinux relabel for exclusive access
		},
		{
			Type:   "volume",
			Source: env.VolumeName,
			Target: "/data",
		},
	}
	
	envVars := map[string]string{
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
	}
	
	// Set startup command - let entrypoint handle the default case
	startupCommand := opts.StartupCommand
	if len(startupCommand) == 0 {
		// Use empty command to let Dockerfile CMD and ENTRYPOINT work together
		startupCommand = nil
	}

	runOpts := container.RunOptions{
		Name:       env.ContainerName,
		Image:      imageTag,
		WorkingDir: "/workspace",
		Detach:     true,
		Mounts:     mounts,
		EnvVars:    envVars,
		Command:    startupCommand,
	}
	
	// Add port mappings if requested
	if opts.ExposeAllPorts {
		runOpts.Ports = []container.PortMapping{
			{Host: 0, Container: 0, Protocol: "tcp"}, // Expose all ports
		}
	}
	
	containerID, err := m.containerMgr.GetRuntime().Run(ctx, runOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	cleanup.containerStarted = true
	
	// Step 7: Update environment with container info and mark as running
	env.ContainerID = containerID
	env.Status = "running"
	
	// Add environment to state only after all resources are successfully created
	if err := m.configMgr.AddEnvironment(*env); err != nil {
		return nil, fmt.Errorf("failed to add environment to state: %w", err)
	}
	cleanup.environmentInState = true
	
	return env, nil
}

// ListEnvironments returns all environments with their current status
func (m *Manager) ListEnvironments(ctx context.Context) ([]config.Environment, error) {
	environments := m.configMgr.GetState().Environments
	
	// Update status for each environment
	for i := range environments {
		if environments[i].ContainerID != "" {
			status, err := m.containerMgr.GetRuntime().Status(ctx, environments[i].ContainerID)
			if err == nil && status.Running {
				environments[i].Status = "running"
			} else {
				environments[i].Status = "stopped"
			}
		}
	}
	
	return environments, nil
}

// DeleteEnvironment removes an environment and cleans up all resources
func (m *Manager) DeleteEnvironment(ctx context.Context, envName string) error {
	_, err := m.configMgr.GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}
	
	return m.CleanupEnvironment(ctx, envName)
}

// CleanupEnvironment performs cleanup of environment resources
func (m *Manager) CleanupEnvironment(ctx context.Context, envName string) error {
	env, err := m.configMgr.GetEnvironment(envName)
	if err != nil {
		// Environment not in state, but try to clean up anyway
		env = config.Environment{Name: envName}
	}
	
	var cleanupErrors []error
	
	// Stop and remove container
	if env.ContainerID != "" {
		if err := m.containerMgr.GetRuntime().Stop(ctx, env.ContainerID); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to stop container: %w", err))
		}
		
		if err := m.containerMgr.GetRuntime().Remove(ctx, env.ContainerID); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove container: %w", err))
		}
	} else if env.ContainerName != "" {
		// Try with container name
		if err := m.containerMgr.GetRuntime().Stop(ctx, env.ContainerName); err != nil {
			// Might already be stopped, continue
		}
		if err := m.containerMgr.GetRuntime().Remove(ctx, env.ContainerName); err != nil {
			// Might already be removed, continue
		}
	}
	
	// Remove container image
	imageTag := fmt.Sprintf("cc-buddy-%s:latest", envName)
	if err := m.containerMgr.GetRuntime().RemoveImage(ctx, imageTag); err != nil {
		// Image removal might fail if other containers are using it, that's okay
		// Don't add to cleanupErrors as this is not critical
	}
	
	// Remove volume
	if env.VolumeName != "" {
		if err := m.containerMgr.GetRuntime().RemoveVolume(ctx, env.VolumeName); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove volume: %w", err))
		}
	}
	
	// Remove worktree
	if env.WorktreePath != "" {
		if err := m.gitOps.RemoveWorktree(ctx, env.WorktreePath); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove worktree: %w", err))
		}
	}
	
	// Remove from state
	if err := m.configMgr.RemoveEnvironment(envName); err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove from state: %w", err))
	}
	
	if len(cleanupErrors) > 0 {
		return fmt.Errorf("cleanup errors: %v", cleanupErrors)
	}
	
	return nil
}

// OpenTerminal opens a terminal session in the environment's container
func (m *Manager) OpenTerminal(ctx context.Context, envName string) error {
	env, err := m.configMgr.GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}
	
	if env.ContainerID == "" {
		return fmt.Errorf("environment %s has no running container", envName)
	}
	
	// Check container status
	status, err := m.containerMgr.GetRuntime().Status(ctx, env.ContainerID)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	
	if !status.Running {
		return fmt.Errorf("container for environment %s is not running", envName)
	}
	
	// Open terminal
	return m.containerMgr.GetRuntime().Exec(ctx, env.ContainerID, []string{"/bin/bash"})
}

// ExecuteCommand executes a command in the environment's container
func (m *Manager) ExecuteCommand(ctx context.Context, envName string, command []string, interactive bool) error {
	env, err := m.configMgr.GetEnvironment(envName)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}
	
	if env.ContainerID == "" {
		return fmt.Errorf("environment %s has no running container", envName)
	}
	
	// Check container status
	status, err := m.containerMgr.GetRuntime().Status(ctx, env.ContainerID)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	
	if !status.Running {
		return fmt.Errorf("container for environment %s is not running", envName)
	}
	
	// Execute command with runtime-specific implementation
	if interactive {
		return m.containerMgr.GetRuntime().Exec(ctx, env.ContainerID, command)
	} else {
		return m.containerMgr.GetRuntime().ExecNonInteractive(ctx, env.ContainerID, command)
	}
}

// GetConfig returns the configuration manager
func (m *Manager) GetConfig() *config.Manager {
	return m.configMgr
}

// GetContainerManager returns the container manager
func (m *Manager) GetContainerManager() *container.Manager {
	return m.containerMgr
}

// GetGitOperations returns the git operations
func (m *Manager) GetGitOperations() *GitOperations {
	return m.gitOps
}