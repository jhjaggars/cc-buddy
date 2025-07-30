package utils

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// OperationType represents the type of operation
type OperationType int

const (
	ContainerBuild OperationType = iota
	EnvironmentCreate
	EnvironmentDelete
	GitWorktree
	ContainerStart
)

// String returns the string representation of the operation type
func (ot OperationType) String() string {
	switch ot {
	case ContainerBuild:
		return "Container Build"
	case EnvironmentCreate:
		return "Environment Create"
	case EnvironmentDelete:
		return "Environment Delete"
	case GitWorktree:
		return "Git Worktree"
	case ContainerStart:
		return "Container Start"
	default:
		return "Unknown"
	}
}

// Operation represents a long-running operation
type Operation struct {
	ID          string
	Type        OperationType
	Environment string
	StartTime   time.Time
	Context     context.Context
	Cancel      context.CancelFunc
	Cleanup     []CleanupFunc
	Progress    float64
	Status      string
	Error       error
	mu          sync.RWMutex
}

// CleanupFunc is a function that performs cleanup
type CleanupFunc func() error

// OperationManager manages long-running operations
type OperationManager struct {
	mu         sync.RWMutex
	operations map[string]*Operation
	logger     *slog.Logger
	idCounter  int
}

// NewOperationManager creates a new operation manager
func NewOperationManager() *OperationManager {
	return &OperationManager{
		operations: make(map[string]*Operation),
		logger:     slog.Default(),
	}
}

// StartOperation starts a new operation
func (om *OperationManager) StartOperation(opType OperationType, env string) (*Operation, error) {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	om.idCounter++
	id := fmt.Sprintf("op-%d", om.idCounter)
	
	op := &Operation{
		ID:          id,
		Type:        opType,
		Environment: env,
		StartTime:   time.Now(),
		Context:     ctx,
		Cancel:      cancel,
		Cleanup:     make([]CleanupFunc, 0),
		Status:      "starting",
	}
	
	om.operations[id] = op
	om.logger.Info("Started operation", "id", id, "type", opType.String(), "environment", env)
	
	return op, nil
}

// CompleteOperation marks an operation as completed
func (om *OperationManager) CompleteOperation(id string) error {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	op, exists := om.operations[id]
	if !exists {
		return fmt.Errorf("operation %s not found", id)
	}
	
	op.mu.Lock()
	op.Status = "completed"
	op.Progress = 1.0
	op.mu.Unlock()
	
	delete(om.operations, id)
	om.logger.Info("Completed operation", "id", id, "duration", time.Since(op.StartTime))
	
	return nil
}

// FailOperation marks an operation as failed
func (om *OperationManager) FailOperation(id string, err error) error {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	op, exists := om.operations[id]
	if !exists {
		return fmt.Errorf("operation %s not found", id)
	}
	
	op.mu.Lock()
	op.Status = "failed"
	op.Error = err
	op.mu.Unlock()
	
	// Execute cleanup functions
	for _, cleanup := range op.Cleanup {
		if cleanupErr := cleanup(); cleanupErr != nil {
			om.logger.Error("Cleanup failed", "operation", id, "error", cleanupErr)
		}
	}
	
	delete(om.operations, id)
	om.logger.Error("Failed operation", "id", id, "error", err, "duration", time.Since(op.StartTime))
	
	return nil
}

// UpdateProgress updates the progress of an operation
func (om *OperationManager) UpdateProgress(id string, progress float64, status string) error {
	om.mu.RLock()
	op, exists := om.operations[id]
	om.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("operation %s not found", id)
	}
	
	op.mu.Lock()
	op.Progress = progress
	if status != "" {
		op.Status = status
	}
	op.mu.Unlock()
	
	return nil
}

// GetActiveOperations returns all currently active operations
func (om *OperationManager) GetActiveOperations() []Operation {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	operations := make([]Operation, 0, len(om.operations))
	for _, op := range om.operations {
		op.mu.RLock()
		operations = append(operations, *op)
		op.mu.RUnlock()
	}
	
	return operations
}

// CancelAll cancels all active operations
func (om *OperationManager) CancelAll(ctx context.Context) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	for _, op := range om.operations {
		om.logger.Info("Cancelling operation", "id", op.ID, "type", op.Type.String())
		op.Cancel()
	}
}

// ForceCancel immediately cancels all operations without cleanup
func (om *OperationManager) ForceCancel() {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	for id, op := range om.operations {
		op.Cancel()
		delete(om.operations, id)
	}
}

// WaitForCompletion waits for all operations to complete or timeout
func (om *OperationManager) WaitForCompletion(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			om.mu.RLock()
			count := len(om.operations)
			om.mu.RUnlock()
			
			if count == 0 {
				return nil
			}
		}
	}
}

// RegisterCleanup adds a cleanup function to an operation
func (om *OperationManager) RegisterCleanup(id string, cleanup CleanupFunc) error {
	om.mu.RLock()
	op, exists := om.operations[id]
	om.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("operation %s not found", id)
	}
	
	op.mu.Lock()
	op.Cleanup = append(op.Cleanup, cleanup)
	op.mu.Unlock()
	
	return nil
}

// GetOperation returns an operation by ID
func (om *OperationManager) GetOperation(id string) (*Operation, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	op, exists := om.operations[id]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", id)
	}
	
	return op, nil
}