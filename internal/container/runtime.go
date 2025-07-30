package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Status represents container status
type Status struct {
	Running bool
	Health  string
	Uptime  string
}

// RunOptions holds container run configuration
type RunOptions struct {
	Name        string
	Image       string
	WorkingDir  string
	Mounts      []Mount
	Ports       []PortMapping
	EnvVars     map[string]string
	Detach      bool
	Remove      bool
	Interactive bool
	TTY         bool
	Command     []string
}

// Mount represents a volume mount
type Mount struct {
	Source  string
	Target  string
	Type    string   // "bind", "volume", etc.
	Options []string // Mount options like "Z", "ro", etc.
}

// PortMapping represents port forwarding
type PortMapping struct {
	Host      int
	Container int
	Protocol  string // "tcp", "udp"
}

// BuildOptions holds container build configuration
type BuildOptions struct {
	Context       string
	Dockerfile    string
	Tags          []string
	BuildArgs     map[string]string
	Target        string
	NoCache       bool
	Progress      string // "auto", "plain", "tty"
}

// Runtime defines the interface for container operations
type Runtime interface {
	// Detect returns the runtime name if available
	Detect(ctx context.Context) (string, error)
	
	// Build builds a container image
	Build(ctx context.Context, opts BuildOptions) error
	
	// Run starts a new container
	Run(ctx context.Context, opts RunOptions) (string, error)
	
	// Stop stops a running container
	Stop(ctx context.Context, containerID string) error
	
	// Remove removes a container
	Remove(ctx context.Context, containerID string) error
	
	// Exec executes a command in a running container (interactive mode)
	Exec(ctx context.Context, containerID string, command []string) error
	
	// ExecNonInteractive executes a command in a running container (non-interactive mode)
	ExecNonInteractive(ctx context.Context, containerID string, command []string) error
	
	// Status returns the status of a container
	Status(ctx context.Context, containerID string) (Status, error)
	
	// Logs returns container logs
	Logs(ctx context.Context, containerID string, follow bool) ([]string, error)
	
	// CreateVolume creates a named volume
	CreateVolume(ctx context.Context, name string) error
	
	// RemoveVolume removes a named volume
	RemoveVolume(ctx context.Context, name string) error
	
	// RemoveImage removes a container image
	RemoveImage(ctx context.Context, imageID string) error
}

// Manager manages container runtime detection and operations
type Manager struct {
	runtime Runtime
}

// NewManager creates a new container manager with auto-detected runtime
func NewManager() (*Manager, error) {
	ctx := context.Background()
	
	// Try Podman first (preferred)
	podman := &PodmanRuntime{}
	if isRuntimeAvailable(ctx, podman) {
		return &Manager{runtime: podman}, nil
	}
	
	// Fall back to Docker
	docker := &DockerRuntime{}
	if isRuntimeAvailable(ctx, docker) {
		return &Manager{runtime: docker}, nil
	}
	
	return nil, fmt.Errorf("no container runtime found (tried podman, docker)")
}

// NewManagerWithRuntime creates a manager with a specific runtime
func NewManagerWithRuntime(runtimeName string) (*Manager, error) {
	ctx := context.Background()
	
	var runtime Runtime
	switch strings.ToLower(runtimeName) {
	case "podman":
		runtime = &PodmanRuntime{}
	case "docker":
		runtime = &DockerRuntime{}
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtimeName)
	}
	
	if !isRuntimeAvailable(ctx, runtime) {
		return nil, fmt.Errorf("runtime %s is not available", runtimeName)
	}
	
	return &Manager{runtime: runtime}, nil
}

// GetRuntime returns the underlying runtime interface
func (m *Manager) GetRuntime() Runtime {
	return m.runtime
}

// isRuntimeAvailable checks if a runtime is available on the system
func isRuntimeAvailable(ctx context.Context, runtime Runtime) bool {
	_, err := runtime.Detect(ctx)
	return err == nil
}

// Base implementation for common runtime operations
type baseRuntime struct {
	command string
}

func (r *baseRuntime) execCommand(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.command, args...)
	return cmd.Output()
}

func (r *baseRuntime) execCommandStreaming(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.command, args...)
	cmd.Stdout = nil // TODO: wire up to progress reporting
	cmd.Stderr = nil // TODO: wire up to error reporting
	return cmd.Run()
}

func (r *baseRuntime) execCommandInteractive(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PodmanRuntime implements Runtime for Podman
type PodmanRuntime struct {
	baseRuntime
}

func (r *PodmanRuntime) Detect(ctx context.Context) (string, error) {
	r.command = "podman"
	out, err := r.execCommand(ctx, "--version")
	if err != nil {
		return "", fmt.Errorf("podman not available: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *PodmanRuntime) Build(ctx context.Context, opts BuildOptions) error {
	args := []string{"build"}
	
	if opts.NoCache {
		args = append(args, "--no-cache")
	}
	
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}
	
	if opts.Dockerfile != "" {
		args = append(args, "-f", opts.Dockerfile)
	}
	
	// Add build arguments in a consistent order
	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}
	
	for _, tag := range opts.Tags {
		args = append(args, "-t", tag)
	}
	
	args = append(args, opts.Context)
	
	return r.execCommandStreaming(ctx, args...)
}

func (r *PodmanRuntime) Run(ctx context.Context, opts RunOptions) (string, error) {
	args := []string{"run"}
	
	if opts.Detach {
		args = append(args, "-d")
	}
	
	if opts.Remove {
		args = append(args, "--rm")
	}
	
	if opts.Interactive {
		args = append(args, "-i")
	}
	
	if opts.TTY {
		args = append(args, "-t")
	}
	
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	
	if opts.WorkingDir != "" {
		args = append(args, "-w", opts.WorkingDir)
	}
	
	for _, mount := range opts.Mounts {
		mountStr := fmt.Sprintf("type=%s,source=%s,target=%s", mount.Type, mount.Source, mount.Target)
		if len(mount.Options) > 0 {
			for _, option := range mount.Options {
				mountStr += "," + option
			}
		}
		args = append(args, "--mount", mountStr)
	}
	
	for _, port := range opts.Ports {
		portStr := fmt.Sprintf("%d:%d/%s", port.Host, port.Container, port.Protocol)
		args = append(args, "-p", portStr)
	}
	
	for key, value := range opts.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	
	args = append(args, opts.Image)
	
	// Add custom command if specified
	if len(opts.Command) > 0 {
		args = append(args, opts.Command...)
	}
	
	out, err := r.execCommand(ctx, args...)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(out)), nil
}

func (r *PodmanRuntime) Stop(ctx context.Context, containerID string) error {
	return r.execCommandStreaming(ctx, "stop", containerID)
}

func (r *PodmanRuntime) Remove(ctx context.Context, containerID string) error {
	return r.execCommandStreaming(ctx, "rm", "-f", containerID)
}

func (r *PodmanRuntime) Exec(ctx context.Context, containerID string, command []string) error {
	args := append([]string{"exec", "-it", containerID}, command...)
	return r.execCommandInteractive(ctx, args...)
}

func (r *PodmanRuntime) ExecNonInteractive(ctx context.Context, containerID string, command []string) error {
	args := append([]string{"exec", containerID}, command...)
	return r.execCommandStreaming(ctx, args...)
}

func (r *PodmanRuntime) Status(ctx context.Context, containerID string) (Status, error) {
	out, err := r.execCommand(ctx, "inspect", "--format", "{{.State.Status}}", containerID)
	if err != nil {
		return Status{Running: false}, fmt.Errorf("failed to get container status: %w", err)
	}
	
	statusStr := strings.TrimSpace(string(out))
	running := statusStr == "running"
	
	// Get uptime if running
	var uptime string
	if running {
		uptimeOut, err := r.execCommand(ctx, "inspect", "--format", "{{.State.StartedAt}}", containerID)
		if err == nil {
			uptime = strings.TrimSpace(string(uptimeOut))
		}
	}
	
	return Status{
		Running: running,
		Health:  statusStr,
		Uptime:  uptime,
	}, nil
}

func (r *PodmanRuntime) Logs(ctx context.Context, containerID string, follow bool) ([]string, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerID)
	
	out, err := r.execCommand(ctx, args...)
	if err != nil {
		return nil, err
	}
	
	return strings.Split(string(out), "\n"), nil
}

func (r *PodmanRuntime) CreateVolume(ctx context.Context, name string) error {
	return r.execCommandStreaming(ctx, "volume", "create", name)
}

func (r *PodmanRuntime) RemoveVolume(ctx context.Context, name string) error {
	return r.execCommandStreaming(ctx, "volume", "rm", name)
}

func (r *PodmanRuntime) RemoveImage(ctx context.Context, imageID string) error {
	return r.execCommandStreaming(ctx, "rmi", imageID)
}

// DockerRuntime implements Runtime for Docker
type DockerRuntime struct {
	baseRuntime
}

func (r *DockerRuntime) Detect(ctx context.Context) (string, error) {
	r.command = "docker"
	out, err := r.execCommand(ctx, "--version")
	if err != nil {
		return "", fmt.Errorf("docker not available: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Docker implementation methods (similar to Podman but with docker command)
func (r *DockerRuntime) Build(ctx context.Context, opts BuildOptions) error {
	args := []string{"build"}
	
	if opts.NoCache {
		args = append(args, "--no-cache")
	}
	
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}
	
	if opts.Dockerfile != "" {
		args = append(args, "-f", opts.Dockerfile)
	}
	
	// Add build arguments in a consistent order
	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}
	
	for _, tag := range opts.Tags {
		args = append(args, "-t", tag)
	}
	
	args = append(args, opts.Context)
	
	return r.execCommandStreaming(ctx, args...)
}

func (r *DockerRuntime) Run(ctx context.Context, opts RunOptions) (string, error) {
	args := []string{"run"}
	
	if opts.Detach {
		args = append(args, "-d")
	}
	
	if opts.Remove {
		args = append(args, "--rm")
	}
	
	if opts.Interactive {
		args = append(args, "-i")
	}
	
	if opts.TTY {
		args = append(args, "-t")
	}
	
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	
	if opts.WorkingDir != "" {
		args = append(args, "-w", opts.WorkingDir)
	}
	
	for _, mount := range opts.Mounts {
		mountStr := fmt.Sprintf("type=%s,source=%s,target=%s", mount.Type, mount.Source, mount.Target)
		if len(mount.Options) > 0 {
			for _, option := range mount.Options {
				mountStr += "," + option
			}
		}
		args = append(args, "--mount", mountStr)
	}
	
	for _, port := range opts.Ports {
		portStr := fmt.Sprintf("%d:%d/%s", port.Host, port.Container, port.Protocol)
		args = append(args, "-p", portStr)
	}
	
	for key, value := range opts.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	
	args = append(args, opts.Image)
	
	// Add custom command if specified
	if len(opts.Command) > 0 {
		args = append(args, opts.Command...)
	}
	
	out, err := r.execCommand(ctx, args...)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(out)), nil
}

func (r *DockerRuntime) Stop(ctx context.Context, containerID string) error {
	return r.execCommandStreaming(ctx, "stop", containerID)
}

func (r *DockerRuntime) Remove(ctx context.Context, containerID string) error {
	return r.execCommandStreaming(ctx, "rm", "-f", containerID)
}

func (r *DockerRuntime) Exec(ctx context.Context, containerID string, command []string) error {
	args := append([]string{"exec", "-it", containerID}, command...)
	return r.execCommandInteractive(ctx, args...)
}

func (r *DockerRuntime) ExecNonInteractive(ctx context.Context, containerID string, command []string) error {
	args := append([]string{"exec", containerID}, command...)
	return r.execCommandStreaming(ctx, args...)
}

func (r *DockerRuntime) Status(ctx context.Context, containerID string) (Status, error) {
	out, err := r.execCommand(ctx, "inspect", "--format", "{{.State.Status}}", containerID)
	if err != nil {
		return Status{Running: false}, fmt.Errorf("failed to get container status: %w", err)
	}
	
	statusStr := strings.TrimSpace(string(out))
	running := statusStr == "running"
	
	// Get uptime if running
	var uptime string
	if running {
		uptimeOut, err := r.execCommand(ctx, "inspect", "--format", "{{.State.StartedAt}}", containerID)
		if err == nil {
			uptime = strings.TrimSpace(string(uptimeOut))
		}
	}
	
	return Status{
		Running: running,
		Health:  statusStr,
		Uptime:  uptime,
	}, nil
}

func (r *DockerRuntime) Logs(ctx context.Context, containerID string, follow bool) ([]string, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerID)
	
	out, err := r.execCommand(ctx, args...)
	if err != nil {
		return nil, err
	}
	
	return strings.Split(string(out), "\n"), nil
}

func (r *DockerRuntime) CreateVolume(ctx context.Context, name string) error {
	return r.execCommandStreaming(ctx, "volume", "create", name)
}

func (r *DockerRuntime) RemoveVolume(ctx context.Context, name string) error {
	return r.execCommandStreaming(ctx, "volume", "rm", name)
}

func (r *DockerRuntime) RemoveImage(ctx context.Context, imageID string) error {
	return r.execCommandStreaming(ctx, "rmi", imageID)
}