package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jhjaggars/cc-buddy/internal/config"
	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// CreateWizardModel handles the environment creation wizard
type CreateWizardModel struct {
	envManager *environment.Manager
	
	// Wizard state
	step        int
	totalSteps  int
	
	// Form inputs
	branchInput     textinput.Model
	branchType      int // 0=new, 1=existing local, 2=remote
	remoteInput     textinput.Model
	worktreeInput   textinput.Model
	
	// UI state
	width   int
	height  int
	focused int
	err     error
	
	// Options
	options environment.CreateEnvironmentOptions
}

// CreateProgressMsg represents progress during environment creation
type CreateProgressMsg struct {
	Step        string
	Progress    float64
	Error       error
	Completed   bool
	Environment *config.Environment
}

// NewCreateWizardModel creates a new creation wizard
func NewCreateWizardModel() *CreateWizardModel {
	envManager, err := environment.NewManager()
	
	// Initialize text inputs
	branchInput := textinput.New()
	branchInput.Placeholder = "Enter branch name (e.g., feature-auth)"
	branchInput.Focus()
	branchInput.CharLimit = 100
	branchInput.Width = 50
	
	remoteInput := textinput.New()
	remoteInput.Placeholder = "origin"
	remoteInput.CharLimit = 50
	remoteInput.Width = 30
	
	worktreeInput := textinput.New()
	worktreeInput.Placeholder = "Leave empty for default"
	worktreeInput.CharLimit = 200
	worktreeInput.Width = 50
	
	return &CreateWizardModel{
		envManager:   envManager,
		step:         0,
		totalSteps:   3,
		branchInput:  branchInput,
		remoteInput:  remoteInput,
		worktreeInput: worktreeInput,
		err:          err,
	}
}

// Init implements tea.Model
func (m *CreateWizardModel) Init() tea.Cmd {
	if m.envManager == nil {
		return nil
	}
	return textinput.Blink
}

// Update implements tea.Model
func (m *CreateWizardModel) Update(msg tea.Msg) (*CreateWizardModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Cancel creation
			return m, tea.Quit
			
		case "tab", "shift+tab", "up", "down":
			// Navigate between inputs within the current step
			if m.step == 0 {
				// Step 0: Branch configuration
				if msg.String() == "tab" || msg.String() == "down" {
					m.focused = (m.focused + 1) % 4 // 3 radio buttons + 1 input
				} else {
					m.focused = (m.focused - 1 + 4) % 4
				}
				m.updateFocus()
			}
			
		case "enter":
			if m.step < m.totalSteps-1 {
				// Move to next step
				if m.validateCurrentStep() {
					m.step++
					m.focused = 0
					m.updateFocus()
				}
			} else {
				// Final step - start creation
				if m.validateCurrentStep() {
					return m, m.startCreation()
				}
			}
			
		case " ": // Space for radio buttons
			if m.step == 0 && m.focused < 3 {
				m.branchType = m.focused
				m.updateFocus()
			}
		}

	case CreateProgressMsg:
		// Handle creation progress updates
		if msg.Error != nil {
			m.err = msg.Error
		} else if msg.Completed {
			// Creation completed successfully
			// TODO: Return to main view or show success message
		}
		return m, nil
	}

	// Update text inputs
	switch m.step {
	case 0:
		if m.focused == 3 { // Branch name input is focused
			m.branchInput, cmd = m.branchInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case 1:
		if m.focused == 0 {
			m.remoteInput, cmd = m.remoteInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case 2:
		if m.focused == 0 {
			m.worktreeInput, cmd = m.worktreeInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *CreateWizardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Esc to cancel", m.err)
	}

	var b strings.Builder
	
	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Create New Environment")
		
	progress := fmt.Sprintf("Step %d of %d", m.step+1, m.totalSteps)
	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(progress)
		
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(m.width-len(title)-len(progress)).Render(""),
		progressStyle,
	)
	
	b.WriteString(header)
	b.WriteString("\n\n")
	
	// Step content
	switch m.step {
	case 0:
		b.WriteString(m.renderBranchStep())
	case 1:
		b.WriteString(m.renderRemoteStep())
	case 2:
		b.WriteString(m.renderConfigStep())
	}
	
	// Footer
	b.WriteString("\n\n")
	if m.step < m.totalSteps-1 {
		b.WriteString("[tab] next field  [enter] continue  [esc] cancel")
	} else {
		b.WriteString("[enter] create environment  [esc] cancel")
	}
	
	return b.String()
}

// SetSize updates the model size
func (m *CreateWizardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// renderBranchStep renders the branch configuration step
func (m *CreateWizardModel) renderBranchStep() string {
	var b strings.Builder
	
	b.WriteString("Branch Configuration\n\n")
	
	// Branch type selection
	branchTypes := []string{
		"Create new branch from HEAD",
		"Use existing local branch", 
		"Use remote branch (origin/...)",
	}
	
	for i, option := range branchTypes {
		var style lipgloss.Style
		if i == m.branchType {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		}
		
		marker := "○"
		if i == m.branchType {
			marker = "●"
		}
		
		focused := ""
		if m.focused == i {
			focused = " <"
		}
		
		b.WriteString(fmt.Sprintf("  %s %s%s\n", 
			style.Render(marker), 
			style.Render(option),
			focused))
	}
	
	b.WriteString("\n")
	
	// Branch name input
	inputLabel := "Branch name:"
	if m.branchType == 2 {
		inputLabel = "Remote branch (without origin/):"
	}
	
	b.WriteString(inputLabel + "\n")
	b.WriteString(m.branchInput.View())
	
	return b.String()
}

// renderRemoteStep renders the remote configuration step
func (m *CreateWizardModel) renderRemoteStep() string {
	if m.branchType != 2 {
		// Skip this step for non-remote branches
		return "Remote configuration not needed for local branches."
	}
	
	var b strings.Builder
	
	b.WriteString("Remote Configuration\n\n")
	b.WriteString("Remote name:\n")
	b.WriteString(m.remoteInput.View())
	b.WriteString("\n\n")
	b.WriteString("Full branch reference: ")
	
	remote := m.remoteInput.Value()
	if remote == "" {
		remote = "origin"
	}
	
	branch := m.branchInput.Value()
	fullRef := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(fmt.Sprintf("%s/%s", remote, branch))
		
	b.WriteString(fullRef)
	
	return b.String()
}

// renderConfigStep renders the final configuration step
func (m *CreateWizardModel) renderConfigStep() string {
	var b strings.Builder
	
	b.WriteString("Final Configuration\n\n")
	
	// Show summary
	b.WriteString("Environment Summary:\n")
	
	branchName := m.branchInput.Value()
	if m.branchType == 2 {
		remote := m.remoteInput.Value()
		if remote == "" {
			remote = "origin"
		}
		b.WriteString(fmt.Sprintf("  Branch: %s/%s (remote)\n", remote, branchName))
	} else {
		typeStr := "new"
		if m.branchType == 1 {
			typeStr = "existing local"
		}
		b.WriteString(fmt.Sprintf("  Branch: %s (%s)\n", branchName, typeStr))
	}
	
	// Generate environment name
	if m.envManager != nil {
		gitOps := m.envManager.GetGitOperations()
		if envName, err := gitOps.GenerateEnvironmentName(branchName); err == nil {
			b.WriteString(fmt.Sprintf("  Environment Name: %s\n", envName))
		}
	}
	
	b.WriteString("\n")
	
	// Worktree directory input
	b.WriteString("Worktree Directory (optional):\n")
	b.WriteString(m.worktreeInput.View())
	
	return b.String()
}

// updateFocus updates which input is focused
func (m *CreateWizardModel) updateFocus() {
	// Reset all focus states
	m.branchInput.Blur()
	m.remoteInput.Blur()
	m.worktreeInput.Blur()
	
	// Set focus based on current step and focused element
	switch m.step {
	case 0:
		if m.focused == 3 { // Branch input
			m.branchInput.Focus()
		}
	case 1:
		if m.focused == 0 { // Remote input
			m.remoteInput.Focus()
		}
	case 2:
		if m.focused == 0 { // Worktree input
			m.worktreeInput.Focus()
		}
	}
}

// validateCurrentStep validates the current step's input
func (m *CreateWizardModel) validateCurrentStep() bool {
	switch m.step {
	case 0:
		// Validate branch name
		branchName := strings.TrimSpace(m.branchInput.Value())
		if branchName == "" {
			m.err = fmt.Errorf("branch name cannot be empty")
			return false
		}
		m.err = nil
		return true
		
	case 1:
		// Validate remote (if applicable)
		if m.branchType == 2 {
			remote := strings.TrimSpace(m.remoteInput.Value())
			if remote == "" {
				m.remoteInput.SetValue("origin") // Set default
			}
		}
		m.err = nil
		return true
		
	case 2:
		// Final validation
		m.err = nil
		return true
		
	default:
		return true
	}
}

// startCreation begins the environment creation process
func (m *CreateWizardModel) startCreation() tea.Cmd {
	// Build options from form data
	branchName := strings.TrimSpace(m.branchInput.Value())
	
	opts := environment.CreateEnvironmentOptions{
		BranchName:     branchName,
		IsRemoteBranch: m.branchType == 2,
	}
	
	if m.branchType == 2 {
		opts.RemoteName = strings.TrimSpace(m.remoteInput.Value())
		if opts.RemoteName == "" {
			opts.RemoteName = "origin"
		}
	}
	
	if worktree := strings.TrimSpace(m.worktreeInput.Value()); worktree != "" {
		opts.WorktreeDir = worktree
	}
	
	return func() tea.Msg {
		ctx := context.Background()
		env, err := m.envManager.CreateEnvironment(ctx, opts)
		return CreateProgressMsg{
			Completed:   err == nil,
			Error:       err,
			Environment: env,
		}
	}
}