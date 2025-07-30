package utils

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SignalHandler manages system signals and graceful shutdown
type SignalHandler struct {
	shutdownCh chan os.Signal
	program    *tea.Program
	operations *OperationManager
	cleanup    *CleanupManager
	logger     *slog.Logger
	mu         sync.RWMutex
	shutdown   bool
}

// InterruptionMsg is sent when the application is interrupted
type InterruptionMsg struct {
	Signal           string
	ActiveOperations []Operation
	Options          []string
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(program *tea.Program, operations *OperationManager) *SignalHandler {
	return &SignalHandler{
		shutdownCh: make(chan os.Signal, 1),
		program:    program,
		operations: operations,
		cleanup:    NewCleanupManager(),
		logger:     slog.Default(),
	}
}

// Start begins signal monitoring
func (sh *SignalHandler) Start() {
	signal.Notify(sh.shutdownCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	
	go sh.handleSignals()
}

// Stop stops signal monitoring
func (sh *SignalHandler) Stop() {
	signal.Stop(sh.shutdownCh)
	close(sh.shutdownCh)
}

// handleSignals processes incoming signals
func (sh *SignalHandler) handleSignals() {
	for sig := range sh.shutdownCh {
		sh.mu.Lock()
		if sh.shutdown {
			sh.mu.Unlock()
			continue
		}
		sh.mu.Unlock()

		switch sig {
		case syscall.SIGINT:
			sh.handleSIGINT()
		case syscall.SIGTERM:
			sh.handleSIGTERM()
		case syscall.SIGQUIT:
			sh.handleSIGQUIT()
		case syscall.SIGHUP:
			sh.handleSIGHUP()
		}
	}
}

// handleSIGINT handles Ctrl+C with user-friendly options
func (sh *SignalHandler) handleSIGINT() {
	activeOps := sh.operations.GetActiveOperations()
	
	if len(activeOps) == 0 {
		// No active operations, quit immediately
		sh.gracefulShutdown()
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

// handleSIGTERM handles process management signals
func (sh *SignalHandler) handleSIGTERM() {
	sh.logger.Info("Received SIGTERM, initiating graceful shutdown")
	sh.gracefulShutdown()
}

// handleSIGQUIT handles force quit signals
func (sh *SignalHandler) handleSIGQUIT() {
	sh.logger.Warn("Received SIGQUIT, forcing immediate shutdown")
	sh.forceShutdown()
}

// handleSIGHUP handles configuration reload
func (sh *SignalHandler) handleSIGHUP() {
	sh.logger.Info("Received SIGHUP, configuration reload not implemented")
}

// gracefulShutdown performs graceful shutdown with timeout
func (sh *SignalHandler) gracefulShutdown() {
	sh.mu.Lock()
	if sh.shutdown {
		sh.mu.Unlock()
		return
	}
	sh.shutdown = true
	sh.mu.Unlock()

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

// forceShutdown performs immediate shutdown
func (sh *SignalHandler) forceShutdown() {
	sh.mu.Lock()
	sh.shutdown = true
	sh.mu.Unlock()

	sh.operations.ForceCancel()
	sh.cleanup.ForceCleanup()
	sh.program.Quit()
}

// RegisterCleanup adds a cleanup task
func (sh *SignalHandler) RegisterCleanup(task CleanupTask) {
	sh.cleanup.Register(task)
}

// IsShutdown returns true if shutdown has been initiated
func (sh *SignalHandler) IsShutdown() bool {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.shutdown
}