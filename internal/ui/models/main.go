package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jhjaggars/cc-buddy/internal/utils"
)

// ViewState represents the current TUI view
type ViewState int

const (
	MainView ViewState = iota
	CreateView
	DeleteView
	ProgressView
	ConfirmationView
	InterruptionView
)

// MainModel is the root Bubble Tea model
type MainModel struct {
	currentView ViewState
	width       int
	height      int
	
	// Sub-models for different views
	listModel           *EnvironmentListModel
	createModel         *CreateWizardModel
	deleteModel         *DeleteModel
	progressModel       *ProgressModel
	confirmationModel   *ConfirmationModel
	interruptionDialog  *InterruptionDialog
	helpModel           *HelpModel
	
	// Operation management
	operationManager    *utils.OperationManager
	signalHandler       *utils.SignalHandler
	
	// Terminal launch state
	terminalEnvName     string
}

// NewMainModel creates a new main model
func NewMainModel() *MainModel {
	operationManager := utils.NewOperationManager()
	
	m := &MainModel{
		currentView:      MainView,
		listModel:        NewEnvironmentListModel(),
		createModel:      NewCreateWizardModel(),
		deleteModel:      NewDeleteModel(),
		helpModel:        NewHelpModel(),
		operationManager: operationManager,
	}
	
	return m
}

// SetProgram sets the Tea program for signal handling
func (m *MainModel) SetProgram(program *tea.Program) {
	m.signalHandler = utils.NewSignalHandler(program, m.operationManager)
	m.signalHandler.Start()
}

// Init implements tea.Model
func (m *MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.listModel.Init(),
		m.createModel.Init(),
		m.deleteModel.Init(),
	)
}

// Update implements tea.Model
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update sub-models with new size
		m.listModel.SetSize(msg.Width, msg.Height)
		m.createModel.SetSize(msg.Width, msg.Height)
		m.deleteModel.SetSize(msg.Width, msg.Height)
		if m.progressModel != nil {
			m.progressModel.SetSize(msg.Width, msg.Height)
		}
		if m.confirmationModel != nil {
			m.confirmationModel.SetSize(msg.Width, msg.Height)
		}
		m.helpModel.SetSize(msg.Width, msg.Height)
		
	case utils.InterruptionMsg:
		// Handle signal interruption
		m.showInterruptionDialog(msg)
		return m, nil
		
	case ConfirmationResult:
		// Handle confirmation dialog result
		if msg.Confirmed {
			return m.handleConfirmationResult(msg)
		} else {
			m.currentView = MainView
			m.confirmationModel = nil
		}
		return m, nil
		
	case CreateProgressMsg:
		// Handle creation progress
		if msg.Error != nil {
			// Show error and return to main view
			m.currentView = MainView
			m.progressModel = nil
		} else if msg.Completed {
			// Creation completed, refresh list and return to main
			m.currentView = MainView
			m.progressModel = nil
			return m, func() tea.Msg { return RefreshEnvironmentsMsg{} }
		}
		return m, nil

	case OpenTerminalMsg:
		// Store environment name and quit to launch terminal
		m.terminalEnvName = msg.Environment
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Let signal handler manage this
			return m, nil
			
		case "q":
			if m.currentView == MainView {
				return m, tea.Quit
			}
			// In other views, return to main
			m.currentView = MainView
			m.progressModel = nil
			m.confirmationModel = nil
			return m, nil
			
		case "n":
			if m.currentView == MainView {
				m.currentView = CreateView
				m.helpModel.SetContext(CreateHelpContext)
				return m, nil
			}
			
		case "?", "h":
			// Toggle help
			m.helpModel.Update(msg)
			return m, nil
		}
	}

	// Update help model (always active for overlay)
	m.helpModel, cmd = m.helpModel.Update(msg)
	cmds = append(cmds, cmd)

	// Route updates to appropriate sub-model
	switch m.currentView {
	case MainView:
		m.helpModel.SetContext(ListHelpContext)
		m.listModel, cmd = m.listModel.Update(msg)
		cmds = append(cmds, cmd)
		
	case CreateView:
		m.helpModel.SetContext(CreateHelpContext)
		m.createModel, cmd = m.createModel.Update(msg)
		cmds = append(cmds, cmd)
		
	case DeleteView:
		m.deleteModel, cmd = m.deleteModel.Update(msg)
		cmds = append(cmds, cmd)
		
	case ProgressView:
		m.helpModel.SetContext(ProgressHelpContext)
		if m.progressModel != nil {
			m.progressModel, cmd = m.progressModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		
	case ConfirmationView:
		m.helpModel.SetContext(ConfirmationHelpContext)
		if m.confirmationModel != nil {
			m.confirmationModel, cmd = m.confirmationModel.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *MainModel) View() string {
	var baseView string
	
	switch m.currentView {
	case MainView:
		baseView = m.renderMainView()
	case CreateView:
		baseView = m.createModel.View()
	case DeleteView:
		baseView = m.deleteModel.View()
	case ProgressView:
		if m.progressModel != nil {
			baseView = m.progressModel.View()
		} else {
			baseView = "Error: progress model not initialized"
		}
	case ConfirmationView:
		if m.confirmationModel != nil {
			baseView = m.confirmationModel.View()
		} else {
			baseView = "Error: confirmation model not initialized"
		}
	case InterruptionView:
		if m.interruptionDialog != nil {
			baseView = m.interruptionDialog.View()
		} else {
			baseView = "Error: interruption dialog not initialized"
		}
	default:
		baseView = "Unknown view state"
	}
	
	// Overlay help if visible
	helpView := m.helpModel.View()
	if helpView != "" {
		// Create overlay effect
		return lipgloss.NewStyle().Render(baseView + "\n" + helpView)
	}
	
	return baseView
}

func (m *MainModel) renderMainView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("cc-buddy")
		
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[q] quit  [n] new environment  [?] help")
		
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(m.width-len(title)-len(help)).Render(""),
		help,
	)
	
	content := m.listModel.View()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
	)
}



type DeleteModel struct {
	width  int
	height int
}

func NewDeleteModel() *DeleteModel {
	return &DeleteModel{}
}

func (m *DeleteModel) Init() tea.Cmd {
	return nil
}

func (m *DeleteModel) Update(msg tea.Msg) (*DeleteModel, tea.Cmd) {
	return m, nil
}

func (m *DeleteModel) View() string {
	return "Delete environment confirmation coming soon..."
}

func (m *DeleteModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// showInterruptionDialog displays the interruption dialog
func (m *MainModel) showInterruptionDialog(msg utils.InterruptionMsg) {
	// TODO: Implement interruption dialog
	m.currentView = InterruptionView
}

// handleConfirmationResult processes confirmation dialog results
func (m *MainModel) handleConfirmationResult(result ConfirmationResult) (tea.Model, tea.Cmd) {
	m.currentView = MainView
	m.confirmationModel = nil
	
	// TODO: Process the confirmation result based on context
	return m, func() tea.Msg { return RefreshEnvironmentsMsg{} }
}

// ShowProgress displays a progress dialog
func (m *MainModel) ShowProgress(title string, steps []string) {
	m.progressModel = NewProgressModel(title, steps)
	m.currentView = ProgressView
}

// ShowConfirmation displays a confirmation dialog
func (m *MainModel) ShowConfirmation(title, message string, details []string) {
	m.confirmationModel = NewConfirmationModel(title, message, details)
	m.currentView = ConfirmationView
}

// GetTerminalEnvironment returns the environment name for terminal launch
func (m *MainModel) GetTerminalEnvironment() string {
	return m.terminalEnvName
}

// Cleanup performs cleanup when the model is destroyed
func (m *MainModel) Cleanup() {
	if m.signalHandler != nil {
		m.signalHandler.Stop()
	}
}

type InterruptionDialog struct{}

func (d *InterruptionDialog) View() string {
	return "Interruption dialog coming soon..."
}