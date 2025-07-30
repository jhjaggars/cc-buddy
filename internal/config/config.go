package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	StateDir          = ".cc-buddy"
	EnvironmentsFile  = "environments.json"
	ConfigFile        = "config.json"
)

// Manager handles configuration and state persistence
type Manager struct {
	stateDir string
	config   *Config
	state    *State
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	stateDir := StateDir
	
	// Ensure state directory exists
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}
	
	return &Manager{
		stateDir: stateDir,
		config:   DefaultConfig(),
		state:    &State{Environments: []Environment{}},
	}, nil
}

// LoadConfig loads configuration from disk or creates default if not found
func (m *Manager) LoadConfig() error {
	configPath := filepath.Join(m.stateDir, ConfigFile)
	
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Config doesn't exist, use defaults and save
		return m.SaveConfig()
	}
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// SaveConfig saves current configuration to disk
func (m *Manager) SaveConfig() error {
	configPath := filepath.Join(m.stateDir, ConfigFile)
	
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// LoadState loads environment state from disk
func (m *Manager) LoadState() error {
	statePath := filepath.Join(m.stateDir, EnvironmentsFile)
	
	data, err := os.ReadFile(statePath)
	if os.IsNotExist(err) {
		// State file doesn't exist, use empty state
		m.state = &State{Environments: []Environment{}}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}
	
	if err := json.Unmarshal(data, m.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}
	
	return nil
}

// SaveState saves current environment state to disk
func (m *Manager) SaveState() error {
	statePath := filepath.Join(m.stateDir, EnvironmentsFile)
	
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	
	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetState returns the current state
func (m *Manager) GetState() *State {
	return m.state
}

// AddEnvironment adds a new environment to the state
func (m *Manager) AddEnvironment(env Environment) error {
	// Check for duplicate names
	for _, existing := range m.state.Environments {
		if existing.Name == env.Name {
			return fmt.Errorf("environment with name %s already exists", env.Name)
		}
	}
	
	m.state.Environments = append(m.state.Environments, env)
	return m.SaveState()
}

// RemoveEnvironment removes an environment from the state
func (m *Manager) RemoveEnvironment(name string) error {
	for i, env := range m.state.Environments {
		if env.Name == name {
			m.state.Environments = append(m.state.Environments[:i], m.state.Environments[i+1:]...)
			return m.SaveState()
		}
	}
	return fmt.Errorf("environment %s not found", name)
}

// UpdateEnvironment updates an existing environment in the state
func (m *Manager) UpdateEnvironment(name string, updater func(*Environment)) error {
	for i, env := range m.state.Environments {
		if env.Name == name {
			updater(&m.state.Environments[i])
			return m.SaveState()
		}
	}
	return fmt.Errorf("environment %s not found", name)
}

// GetEnvironment returns an environment by name
func (m *Manager) GetEnvironment(name string) (Environment, error) {
	for _, env := range m.state.Environments {
		if env.Name == name {
			return env, nil
		}
	}
	return Environment{}, fmt.Errorf("environment %s not found", name)
}