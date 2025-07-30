# Go + Charm.sh Conversion Plan for cc-buddy

## Project Overview

This document outlines the plan to convert `cc-buddy` from a bash script to a Go application using the Charm.sh ecosystem for creating an attractive terminal user interface (TUI).

## Current State Analysis

The existing `cc-buddy` bash script (591 lines) provides:
- Git worktree creation and management
- Container building and lifecycle management
- Environment state tracking via JSON files
- Support for Docker/Podman runtimes
- Commands: `init`, `create`, `list`, `delete`, `terminal`

## Target Architecture

### Core Framework Stack
- **Bubble Tea**: TUI framework based on Elm architecture (Model-View-Update pattern)
- **Bubbles**: Pre-built UI components (lists, inputs, tables, progress bars)
- **Lip Gloss**: Styling and layout library for terminal applications

### Project Structure
```
cc-buddy-go/
├── cmd/
│   └── cc-buddy/
│       └── main.go                 # Entry point and CLI parsing
├── internal/
│   ├── commands/
│   │   ├── create.go              # Environment creation logic
│   │   ├── list.go                # Environment listing
│   │   ├── delete.go              # Environment cleanup
│   │   ├── init.go                # Containerfile.dev generation
│   │   └── terminal.go            # Container shell access
│   ├── ui/
│   │   ├── models/
│   │   │   ├── main.go            # Root Bubble Tea model
│   │   │   ├── create.go          # Creation wizard model
│   │   │   ├── list.go            # Environment list model
│   │   │   └── confirm.go         # Confirmation dialog model
│   │   ├── components/
│   │   │   ├── table.go           # Environment table component
│   │   │   ├── progress.go        # Progress indicator component
│   │   │   └── input.go           # Input field components
│   │   └── styles.go              # Lip Gloss styling definitions
│   ├── config/
│   │   ├── config.go              # Configuration management
│   │   └── state.go               # JSON state persistence
│   ├── environment/
│   │   ├── manager.go             # Environment lifecycle management
│   │   ├── git.go                 # Git operations (worktree, branch)
│   │   └── validation.go          # Environment validation logic
│   ├── container/
│   │   ├── runtime.go             # Docker/Podman abstraction
│   │   ├── builder.go             # Container building logic
│   │   └── executor.go            # Container execution
│   └── utils/
│       ├── logger.go              # Structured logging
│       └── errors.go              # Error handling utilities
├── go.mod
├── go.sum
└── README.md
```

## Functional Requirements

### Core Commands (Feature Parity)
1. **init**: Generate Containerfile.dev with interactive options
2. **create**: Environment creation with enhanced validation and progress feedback
3. **list**: Interactive environment listing with real-time status
4. **delete**: Environment cleanup with confirmation dialogs
5. **terminal**: Container shell access with connection validation

### Enhanced User Experience
- **Interactive Selection**: Navigate environments with keyboard instead of CLI args
- **Real-time Updates**: Live status monitoring for containers and operations
- **Progress Visualization**: Animated progress bars for long-running operations
- **Input Validation**: Real-time validation with helpful error messages
- **Confirmation Dialogs**: Interactive prompts for destructive operations
- **Context-Sensitive Help**: Dynamic help system based on current screen

### TUI-Specific Features
- **Environment Dashboard**: Main screen showing all environments with status
- **Creation Wizard**: Step-by-step environment creation with form validation
- **Live Monitoring**: Real-time container resource usage and logs
- **Configuration UI**: Interactive settings management
- **Keyboard Navigation**: Full keyboard-driven interface with shortcuts

## Technical Requirements

### Data Models
```go
type Environment struct {
    Name          string    `json:"name"`
    Branch        string    `json:"branch"`
    WorktreePath  string    `json:"worktree_path"`
    ContainerID   string    `json:"container_id"`
    ContainerName string    `json:"container_name"`
    VolumeName    string    `json:"volume_name"`
    Created       time.Time `json:"created"`
    Status        string    `json:"status"`
}

type Config struct {
    WorktreeDir   string `json:"worktree_dir"`
    Runtime       string `json:"runtime"`
    Containerfile string `json:"containerfile"`
    ExposeAll     bool   `json:"expose_all"`
}
```

### State Management
- **JSON Persistence**: Maintain compatibility with existing `.cc-buddy/environments.json`
- **Atomic Operations**: Ensure state consistency during environment lifecycle
- **Error Recovery**: Graceful handling of partial state and orphaned resources
- **Configuration**: User preferences with defaults and validation

### Container Runtime Abstraction
```go
type ContainerRuntime interface {
    Detect() (string, error)
    Build(context, containerfile, tags, args) error
    Run(image, name, options) (string, error)
    Stop(container) error
    Remove(container) error
    Exec(container, command) error
    Status(container) (Status, error)
}
```

### Git Operations
- **Branch Management**: Creation, validation, and checkout
- **Worktree Operations**: Creation, removal, and path management
- **Repository Detection**: Remote URL parsing and local repo validation
- **Remote Handling**: Support for origin/branch-name format

### Error Handling
- **Structured Errors**: Type-safe error handling with context
- **User-Friendly Messages**: Clear error descriptions with suggested actions
- **Recovery Options**: Automatic cleanup and retry mechanisms
- **Logging**: Structured logging with configurable levels

## UI Components & User Experience

### Main Dashboard
```
┌─ cc-buddy ─────────────────────────────────────────────────────┐
│                                                                │
│  Environments (3)                                    [q] quit  │
│                                                                │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ NAME                 BRANCH       STATUS    CREATED        │ │
│  │ repo-feature-1      feature-1    🟢 running  2h ago       │ │
│  │ repo-bugfix         bugfix       🟡 stopped  1d ago       │ │
│  │ repo-main           main         🟢 running  3h ago       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                │
│  [↑↓] navigate  [enter] terminal  [d] delete  [n] new  [r] refresh │
└────────────────────────────────────────────────────────────────┘
```

### Creation Wizard
```
┌─ Create New Environment ───────────────────────────────────────┐
│                                                                │
│  Step 1 of 3: Branch Configuration                            │
│                                                                │
│  Branch name: feature-auth-system_                            │
│                                                                │
│  ○ Create new branch from HEAD                                 │
│  ○ Use existing local branch                                   │
│  ○ Use remote branch (origin/...)                             │
│                                                                │
│  Worktree location: /home/user/.worktrees/repo-feature-auth-system │
│                                                                │
│  [tab] next field  [enter] continue  [esc] cancel             │
└────────────────────────────────────────────────────────────────┘
```

### Progress Display
```
┌─ Creating Environment: repo-feature-auth ─────────────────────┐
│                                                                │
│  ✓ Branch validation                                          │
│  ✓ Worktree creation                                          │
│  ⟳ Building container...                    [████████░░] 80%  │
│  ○ Starting container                                          │
│  ○ Environment setup                                           │
│                                                                │
│  Building image layers... (2m 15s remaining)                  │
│                                                                │
│  [ctrl+c] cancel                                               │
└────────────────────────────────────────────────────────────────┘
```

## Technical Benefits

### User Experience Improvements
- **Interactive Selection**: Navigate environments with arrow keys instead of remembering names
- **Real-time Feedback**: Live progress bars and status updates during operations
- **Better Error Handling**: Graceful error display with suggested actions and retry options
- **Context-Sensitive Help**: Show relevant keyboard shortcuts based on current screen
- **Visual Status**: Color-coded indicators for environment health and container status

### Developer Experience Improvements
- **Type Safety**: Compile-time error checking vs runtime bash errors
- **Better Testing**: Unit tests for all business logic components
- **Code Organization**: Clear separation of concerns vs monolithic bash script
- **Error Handling**: Structured error types with context vs simple exit codes
- **Performance**: Faster startup and operations, especially for file system operations

### Operational Improvements
- **Cross-platform**: Better Windows and macOS support than bash
- **Resource Usage**: Lower memory footprint than bash + external tools
- **Logging**: Structured logging with levels and context
- **Configuration**: Typed configuration with validation
- **Maintainability**: Easier to extend and modify than 800+ line bash script

## Testing Strategy
- Unit tests for all business logic components
- Integration tests with real git repositories and containers
- TUI interaction testing with automated key sequences
- Performance benchmarking vs bash implementation
- Cross-platform testing (Linux, macOS, Windows)

## Success Metrics

### Functional Metrics
- 100% feature parity with bash version
- All existing workflows continue to work
- No breaking changes to JSON state format
- Performance equal or better than bash version

### User Experience Metrics
- Reduced time to create environments (interactive vs CLI)
- Fewer user errors (validation and confirmation dialogs)
- Improved discoverability (help system and visual cues)

### Technical Metrics
- Code coverage >90%
- Cross-platform compatibility
- Memory usage <50MB during normal operation
- Startup time <500ms

## Signal Handling and Graceful Shutdown

### Signal Handling Strategy

The TUI application must handle system signals gracefully to ensure resource cleanup and state consistency, especially during long-running operations like container builds.

#### Supported Signals
- **SIGINT (Ctrl+C)**: Graceful shutdown with user confirmation
- **SIGTERM**: Immediate graceful shutdown for process management
- **SIGHUP**: Configuration reload (if applicable)
- **SIGQUIT**: Force quit with emergency cleanup

#### Implementation Architecture

```go
type SignalHandler struct {
    shutdownCh   chan os.Signal
    program      *tea.Program
    operations   *OperationManager
    cleanup      *CleanupManager
    logger       *slog.Logger
}

func NewSignalHandler(program *tea.Program, ops *OperationManager) *SignalHandler {
    sh := &SignalHandler{
        shutdownCh: make(chan os.Signal, 1),
        program:    program,
        operations: ops,
        cleanup:    NewCleanupManager(),
        logger:     slog.Default(),
    }
    
    signal.Notify(sh.shutdownCh, 
        syscall.SIGINT, 
        syscall.SIGTERM, 
        syscall.SIGQUIT,
        syscall.SIGHUP,
    )
    
    return sh
}

func (sh *SignalHandler) Start() {
    go sh.handleSignals()
}

func (sh *SignalHandler) handleSignals() {
    for sig := range sh.shutdownCh {
        switch sig {
        case syscall.SIGINT:
            sh.handleSIGINT()
        case syscall.SIGTERM:
            sh.handleSIGTERM()
        case syscall.SIGQUIT:
            sh.handleSIGQUIIT()
        case syscall.SIGHUP:
            sh.handleSIGHUP()
        }
    }
}
```

#### SIGINT (Ctrl+C) Handling

SIGINT requires user-friendly handling since it's the most common way users interrupt operations:

```go
func (sh *SignalHandler) handleSIGINT() {
    activeOps := sh.operations.GetActiveOperations()
    
    if len(activeOps) == 0 {
        // No active operations, quit immediately
        sh.program.Quit()
        return
    }
    
    // Show interruption dialog in TUI
    sh.program.Send(InterruptionMsg{
        Signal: "SIGINT",
        ActiveOperations: activeOps,
        Options: []string{
            "Cancel current operations and quit",
            "Wait for operations to complete",
            "Force quit (may leave orphaned resources)",
            "Continue running",
        },
    })
}
```

#### SIGTERM Handling

SIGTERM indicates process management (systemd, Docker, etc.) and requires immediate graceful shutdown:

```go
func (sh *SignalHandler) handleSIGTERM() {
    sh.logger.Info("Received SIGTERM, initiating graceful shutdown")
    
    // Set shutdown timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Cancel all active operations
    sh.operations.CancelAll(ctx)
    
    // Wait for operations to finish or timeout
    if err := sh.operations.WaitForCompletion(ctx); err != nil {
        sh.logger.Warn("Operations did not complete within timeout", "error", err)
        sh.cleanup.ForceCleanup()
    }
    
    // Quit TUI
    sh.program.Quit()
}
```

### Operation Cancellation Framework

Long-running operations must be cancellable and cleanupable:

```go
type Operation struct {
    ID          string
    Type        OperationType
    Environment string
    StartTime   time.Time
    Context     context.Context
    Cancel      context.CancelFunc
    Progress    *Progress
    Cleanup     []CleanupFunc
}

type OperationManager struct {
    mu         sync.RWMutex
    operations map[string]*Operation
    logger     *slog.Logger
}

func (om *OperationManager) StartOperation(opType OperationType, env string) (*Operation, error) {
    om.mu.Lock()
    defer om.mu.Unlock()
    
    ctx, cancel := context.WithCancel(context.Background())
    
    op := &Operation{
        ID:          generateID(),
        Type:        opType,
        Environment: env,
        StartTime:   time.Now(),
        Context:     ctx,
        Cancel:      cancel,
        Progress:    NewProgress(),
        Cleanup:     make([]CleanupFunc, 0),
    }
    
    om.operations[op.ID] = op
    return op, nil
}

func (om *OperationManager) CancelAll(ctx context.Context) {
    om.mu.RLock()
    defer om.mu.RUnlock()
    
    for _, op := range om.operations {
        om.logger.Info("Cancelling operation", "id", op.ID, "type", op.Type)
        op.Cancel()
    }
}
```

### Resource Cleanup Management

Ensure proper cleanup of partially created resources:

```go
type CleanupManager struct {
    tasks  []CleanupTask
    logger *slog.Logger
}

type CleanupTask struct {
    Description string
    Priority    int // Higher priority runs first
    Fn          func() error
}

func (cm *CleanupManager) Register(task CleanupTask) {
    cm.tasks = append(cm.tasks, task)
    // Sort by priority
    sort.Slice(cm.tasks, func(i, j int) bool {
        return cm.tasks[i].Priority > cm.tasks[j].Priority
    })
}

func (cm *CleanupManager) ExecuteAll() error {
    var errors []error
    
    for _, task := range cm.tasks {
        cm.logger.Info("Executing cleanup task", "description", task.Description)
        if err := task.Fn(); err != nil {
            cm.logger.Error("Cleanup task failed", "description", task.Description, "error", err)
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("cleanup failed with %d errors: %v", len(errors), errors)
    }
    
    return nil
}
```

### TUI Integration for Signal Handling

Update Bubble Tea models to handle signal-related messages:

```go
type InterruptionMsg struct {
    Signal           string
    ActiveOperations []Operation
    Options          []string
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case InterruptionMsg:
        // Switch to interruption dialog view
        return m.showInterruptionDialog(msg), nil
    
    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            // Handle Ctrl+C within TUI context
            return m.handleInterruption(), nil
        }
    }
    
    return m, nil
}

func (m MainModel) showInterruptionDialog(msg InterruptionMsg) MainModel {
    dialog := NewInterruptionDialog(msg.ActiveOperations, msg.Options)
    m.currentView = InterruptionView
    m.interruptionDialog = dialog
    return m
}
```

### Emergency State Recovery

Handle scenarios where the application was forcefully terminated:

```go
func (app *App) RecoverOrphanedResources() error {
    stateFile := filepath.Join(app.config.StateDir, "environments.json")
    
    // Read existing state
    environments, err := app.state.LoadEnvironments()
    if err != nil {
        return fmt.Errorf("failed to load environments: %w", err)
    }
    
    // Check each environment for orphaned resources
    for _, env := range environments {
        if err := app.validateEnvironmentState(env); err != nil {
            app.logger.Warn("Found orphaned resources", 
                "environment", env.Name, 
                "error", err)
            
            // Attempt cleanup
            if err := app.cleanupOrphanedEnvironment(env); err != nil {
                app.logger.Error("Failed to cleanup orphaned environment",
                    "environment", env.Name,
                    "error", err)
            }
        }
    }
    
    return nil
}
```

### Signal Handling Testing

Test signal handling scenarios:

```go
func TestSignalHandling(t *testing.T) {
    tests := []struct {
        name           string
        signal         os.Signal
        activeOps      []Operation
        expectedAction string
    }{
        {
            name:           "SIGINT with no active operations",
            signal:         syscall.SIGINT,
            activeOps:      []Operation{},
            expectedAction: "immediate_quit",
        },
        {
            name:   "SIGINT with active container build",
            signal: syscall.SIGINT,
            activeOps: []Operation{
                {Type: ContainerBuild, Environment: "test-env"},
            },
            expectedAction: "show_interruption_dialog",
        },
        {
            name:           "SIGTERM forces graceful shutdown",
            signal:         syscall.SIGTERM,
            activeOps:      []Operation{},
            expectedAction: "graceful_shutdown",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

This signal handling strategy ensures that:
1. User interruptions are handled gracefully with clear options
2. Process management signals trigger appropriate cleanup
3. Long-running operations can be cancelled cleanly
4. Orphaned resources are detected and cleaned up on restart
5. The TUI remains responsive during shutdown procedures