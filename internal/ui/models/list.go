package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jhjaggars/cc-buddy/internal/config"
	"github.com/jhjaggars/cc-buddy/internal/environment"
)

// EnvironmentListModel handles the environment list view
type EnvironmentListModel struct {
	table       table.Model
	envManager  *environment.Manager
	environments []config.Environment
	width       int
	height      int
	loading     bool
	err         error
}

// RefreshEnvironmentsMsg is sent when environments should be refreshed
type RefreshEnvironmentsMsg struct{}

// EnvironmentsLoadedMsg is sent when environments are loaded
type EnvironmentsLoadedMsg struct {
	Environments []config.Environment
	Error        error
}

// NewEnvironmentListModel creates a new environment list model
func NewEnvironmentListModel() *EnvironmentListModel {
	// Initialize environment manager
	envManager, err := environment.NewManager()
	
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Branch", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Created", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &EnvironmentListModel{
		table:      t,
		envManager: envManager,
		loading:    true,
		err:        err,
	}
}

// Init implements tea.Model
func (m *EnvironmentListModel) Init() tea.Cmd {
	return tea.Batch(
		m.refreshEnvironments(),
		m.startPeriodicRefresh(),
	)
}

// startPeriodicRefresh starts periodic status updates
func (m *EnvironmentListModel) startPeriodicRefresh() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return RefreshEnvironmentsMsg{}
	})
}

// Update implements tea.Model
func (m *EnvironmentListModel) Update(msg tea.Msg) (*EnvironmentListModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
		
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh environments
			m.loading = true
			return m, m.refreshEnvironments()
			
		case "enter":
			// Open terminal for selected environment
			if m.table.SelectedRow() != nil {
				envName := m.table.SelectedRow()[0]
				return m, m.openTerminal(envName)
			}
			
		case "d":
			// Delete selected environment
			if m.table.SelectedRow() != nil {
				envName := m.table.SelectedRow()[0]
				// TODO: Show confirmation dialog
				return m, m.deleteEnvironment(envName)
			}
		}

	case RefreshEnvironmentsMsg:
		m.loading = true
		return m, m.refreshEnvironments()

	case EnvironmentsLoadedMsg:
		m.loading = false
		m.err = msg.Error
		if msg.Error == nil {
			m.environments = msg.Environments
			m.updateTableRows()
		}
		// Continue periodic refresh
		return m, m.startPeriodicRefresh()
	}

	// Update table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m *EnvironmentListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading environments: %v\n\nPress 'r' to retry", m.err)
	}

	if m.loading {
		return "Loading environments..."
	}

	if len(m.environments) == 0 {
		return lipgloss.NewStyle().
			Margin(2, 0).
			Render("No environments found.\n\nPress 'n' to create your first environment.")
	}

	// Build the view
	var b strings.Builder
	
	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n\n")
	
	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[↑↓] navigate  [enter] terminal  [d] delete  [n] new  [r] refresh")
	
	b.WriteString(help)
	
	return b.String()
}

// SetSize updates the model size
func (m *EnvironmentListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateTableSize()
}

// updateTableSize adjusts table dimensions based on available space
func (m *EnvironmentListModel) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		// Leave space for header, help text, and margins
		tableHeight := m.height - 8
		if tableHeight < 3 {
			tableHeight = 3
		}
		
		m.table.SetHeight(tableHeight)
		
		// Adjust column widths based on available width
		totalWidth := m.width - 4 // Account for borders and padding
		if totalWidth > 0 {
			nameWidth := totalWidth * 35 / 100
			branchWidth := totalWidth * 30 / 100
			statusWidth := totalWidth * 20 / 100 
			createdWidth := totalWidth - nameWidth - branchWidth - statusWidth
			
			if nameWidth < 15 {
				nameWidth = 15
			}
			if branchWidth < 10 {
				branchWidth = 10
			}
			if statusWidth < 8 {
				statusWidth = 8
			}
			if createdWidth < 8 {
				createdWidth = 8
			}
			
			columns := []table.Column{
				{Title: "Name", Width: nameWidth},
				{Title: "Branch", Width: branchWidth},
				{Title: "Status", Width: statusWidth},
				{Title: "Created", Width: createdWidth},
			}
			m.table.SetColumns(columns)
		}
	}
}

// refreshEnvironments loads environments from the manager
func (m *EnvironmentListModel) refreshEnvironments() tea.Cmd {
	if m.envManager == nil {
		return func() tea.Msg {
			return EnvironmentsLoadedMsg{Error: fmt.Errorf("environment manager not initialized")}
		}
	}
	
	return func() tea.Msg {
		ctx := context.Background()
		environments, err := m.envManager.ListEnvironments(ctx)
		return EnvironmentsLoadedMsg{
			Environments: environments,
			Error:        err,
		}
	}
}

// updateTableRows updates the table with current environment data
func (m *EnvironmentListModel) updateTableRows() {
	var rows []table.Row
	
	for _, env := range m.environments {
		status := getStatusDisplay(env.Status)
		created := formatTimeAgo(env.Created)
		
		rows = append(rows, table.Row{
			env.Name,
			env.Branch,
			status,
			created,
		})
	}
	
	m.table.SetRows(rows)
}

// openTerminal opens a terminal for the specified environment
func (m *EnvironmentListModel) openTerminal(envName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.envManager.OpenTerminal(ctx, envName); err != nil {
			// TODO: Show error message
			return nil
		}
		return nil
	}
}

// deleteEnvironment deletes the specified environment
func (m *EnvironmentListModel) deleteEnvironment(envName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.envManager.DeleteEnvironment(ctx, envName); err != nil {
			// TODO: Show error message
			return nil
		}
		// Refresh environments after deletion
		return RefreshEnvironmentsMsg{}
	}
}

// getStatusDisplay returns a user-friendly status display with emoji
func getStatusDisplay(status string) string {
	switch status {
	case "running":
		return "🟢 running"
	case "stopped":
		return "🟡 stopped"
	case "creating":
		return "🔄 creating"
	case "error":
		return "🔴 error"
	default:
		return status
	}
}

// formatTimeAgo formats a time as "2h ago", "1d ago", etc.
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "now"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm", minutes)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh", hours)
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	
	// For older dates, show the actual date
	return t.Format("Jan 2")
}