package utils

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
)

// CleanupTask represents a cleanup task
type CleanupTask struct {
	Description string
	Priority    int // Higher priority runs first
	Fn          func() error
}

// CleanupManager manages cleanup tasks
type CleanupManager struct {
	tasks  []CleanupTask
	logger *slog.Logger
	mu     sync.Mutex
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager() *CleanupManager {
	return &CleanupManager{
		tasks:  make([]CleanupTask, 0),
		logger: slog.Default(),
	}
}

// Register adds a cleanup task
func (cm *CleanupManager) Register(task CleanupTask) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.tasks = append(cm.tasks, task)
	
	// Sort by priority (highest first)
	sort.Slice(cm.tasks, func(i, j int) bool {
		return cm.tasks[i].Priority > cm.tasks[j].Priority
	})
}

// ExecuteAll executes all cleanup tasks
func (cm *CleanupManager) ExecuteAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
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

// ForceCleanup executes all cleanup tasks without logging errors
func (cm *CleanupManager) ForceCleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	for _, task := range cm.tasks {
		// Best effort cleanup, ignore errors
		task.Fn()
	}
}

// Clear removes all cleanup tasks
func (cm *CleanupManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.tasks = cm.tasks[:0]
}

// Count returns the number of registered cleanup tasks
func (cm *CleanupManager) Count() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	return len(cm.tasks)
}