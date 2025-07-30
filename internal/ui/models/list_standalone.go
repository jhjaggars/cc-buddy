package models

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// StandaloneListModel provides a focused, standalone environment list interface
type StandaloneListModel struct {
	listModel       *EnvironmentListModel
	helpModel       *HelpModel
	confirmModel    *ConfirmationModel
	envManager      *environment.Manager
	
	// UI state
	width           int
	height          int
	showConfirm     bool
	selectedEnvName string
	message         string
	messageStyle    lipgloss.Style
	quitting        bool
}

// NewStandaloneListModel creates a new standalone list model
func NewStandaloneListModel() (*StandaloneListModel, error) {
	envManager, err := environment.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize environment manager: %w", err)
	}

	listModel := NewEnvironmentListModel()
	helpModel := NewHelpModel()
	helpModel.SetContext(ListHelpContext)

	return &StandaloneListModel{
		listModel:    listModel,
		helpModel:    helpModel,
		envManager:   envManager,
		messageStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
	}, nil
}

// Init implements tea.Model
func (m *StandaloneListModel) Init() tea.Cmd {
	return m.listModel.Init()
}

// Update implements tea.Model
func (m *StandaloneListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.listModel.SetSize(msg.Width, msg.Height-4) // Leave space for header/footer
		m.helpModel.SetSize(msg.Width, msg.Height)
		if m.confirmModel != nil {
			m.confirmModel.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.showConfirm {
				// Cancel confirmation
				m.showConfirm = false
				m.confirmModel = nil
				m.selectedEnvName = ""
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "?", "h":
			// Toggle help
			m.helpModel.Update(msg)
			return m, nil

		case "enter":
			if m.showConfirm {
				// Let confirmation model handle this
				break
			}
			// Open terminal for selected environment
			return m.handleTerminalAction()

		case "d":
			if m.showConfirm {
				// Let confirmation model handle this
				break
			}
			// Delete selected environment
			return m.handleDeleteAction()

		case "r":
			if !m.showConfirm {
				// Refresh environments
				m.message = "Refreshing environments..."
				m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
				return m, func() tea.Msg { return RefreshEnvironmentsMsg{} }
			}
		}

	case ConfirmationResult:
		// Handle confirmation dialog result
		m.showConfirm = false
		if msg.Confirmed && m.selectedEnvName != "" {
			return m.executeDelete()
		}
		m.confirmModel = nil
		m.selectedEnvName = ""
		return m, nil

	case RefreshEnvironmentsMsg, EnvironmentsLoadedMsg:
		// Clear any status messages when environments are loaded
		if _, ok := msg.(EnvironmentsLoadedMsg); ok {
			m.message = ""
		}

	case TerminalErrorMsg:
		m.message = fmt.Sprintf("Failed to open terminal for %s: %v", msg.Environment, msg.Error)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return m, nil

	case TerminalSuccessMsg:
		m.message = fmt.Sprintf("Opened terminal for %s", msg.Environment)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		return m, nil

	case DeleteErrorMsg:
		m.message = fmt.Sprintf("Failed to delete %s: %v", msg.Environment, msg.Error)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return m, nil

	case DeleteSuccessMsg:
		m.message = fmt.Sprintf("Successfully deleted %s", msg.Environment)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		// Refresh the environment list
		return m, func() tea.Msg { return RefreshEnvironmentsMsg{} }
	}

	// Update help model
	m.helpModel, cmd = m.helpModel.Update(msg)
	cmds = append(cmds, cmd)

	// Route to appropriate sub-model
	if m.showConfirm && m.confirmModel != nil {
		m.confirmModel, cmd = m.confirmModel.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.listModel, cmd = m.listModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *StandaloneListModel) View() string {
	if m.quitting {
		return ""
	}

	// Build the main view
	var view string

	if m.showConfirm && m.confirmModel != nil {
		// Show confirmation dialog overlay
		view = m.confirmModel.View()
	} else {
		// Show main list interface
		view = m.renderMainView()
	}

	// Overlay help if visible
	helpView := m.helpModel.View()
	if helpView != "" {
		view = lipgloss.NewStyle().Render(view + "\n" + helpView)
	}

	return view
}

// renderMainView renders the main list interface
func (m *StandaloneListModel) renderMainView() string {
	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("cc-buddy - Environment List")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[↑↓] navigate  [enter] terminal  [d] delete  [r] refresh  [q] quit  [?] help")

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(m.width-len(title)-len(help)).Render(""),
		help,
	)

	// List content
	content := m.listModel.View()

	// Footer with status message
	var footer string
	if m.message != "" {
		footer = "\n" + m.messageStyle.Render(m.message)
	}

	return header + "\n\n" + content + footer
}

// handleTerminalAction opens terminal for the selected environment
func (m *StandaloneListModel) handleTerminalAction() (tea.Model, tea.Cmd) {
	selectedRow := m.listModel.table.SelectedRow()
	if selectedRow == nil {
		m.message = "No environment selected"
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		return m, nil
	}

	envName := selectedRow[0]
	
	return m, func() tea.Msg {
		ctx := context.Background()
		if err := m.envManager.OpenTerminal(ctx, envName); err != nil {
			return TerminalErrorMsg{
				Environment: envName,
				Error:       err,
			}
		}
		return TerminalSuccessMsg{Environment: envName}
	}
}

// handleDeleteAction shows confirmation dialog for deletion
func (m *StandaloneListModel) handleDeleteAction() (tea.Model, tea.Cmd) {
	selectedRow := m.listModel.table.SelectedRow()
	if selectedRow == nil {
		m.message = "No environment selected"
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		return m, nil
	}

	envName := selectedRow[0]
	branch := selectedRow[1]
	
	// Get environment details for confirmation
	env, err := m.envManager.GetConfig().GetEnvironment(envName)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return m, nil
	}

	details := []string{
		fmt.Sprintf("Branch: %s", branch),
		fmt.Sprintf("Worktree: %s", env.WorktreePath),
		fmt.Sprintf("Container: %s", env.ContainerName),
		fmt.Sprintf("Volume: %s", env.VolumeName),
	}

	m.confirmModel = NewDeleteConfirmationModel(envName, "Environment", details)
	m.confirmModel.SetSize(m.width, m.height)
	m.selectedEnvName = envName
	m.showConfirm = true

	return m, nil
}

// executeDelete performs the actual deletion
func (m *StandaloneListModel) executeDelete() (tea.Model, tea.Cmd) {
	envName := m.selectedEnvName
	m.selectedEnvName = ""
	m.confirmModel = nil

	return m, func() tea.Msg {
		ctx := context.Background()
		if err := m.envManager.DeleteEnvironment(ctx, envName); err != nil {
			return DeleteErrorMsg{
				Environment: envName,
				Error:       err,
			}
		}
		return DeleteSuccessMsg{Environment: envName}
	}
}

// Message types for async operations
type TerminalErrorMsg struct {
	Environment string
	Error       error
}

type TerminalSuccessMsg struct {
	Environment string
}

type DeleteErrorMsg struct {
	Environment string
	Error       error
}

type DeleteSuccessMsg struct {
	Environment string
}