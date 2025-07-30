package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationModel displays confirmation dialogs for destructive operations
type ConfirmationModel struct {
	title       string
	message     string
	details     []string
	confirmText string
	cancelText  string
	selected    int // 0 = cancel, 1 = confirm
	width       int
	height      int
	confirmed   bool
	cancelled   bool
}

// ConfirmationResult represents the result of a confirmation dialog
type ConfirmationResult struct {
	Confirmed bool
	Data      interface{} // Optional data to pass along
}

// NewConfirmationModel creates a new confirmation dialog
func NewConfirmationModel(title, message string, details []string) *ConfirmationModel {
	return &ConfirmationModel{
		title:       title,
		message:     message,
		details:     details,
		confirmText: "Yes, proceed",
		cancelText:  "Cancel",
		selected:    0, // Default to cancel for safety
	}
}

// NewDeleteConfirmationModel creates a confirmation dialog for deletion
func NewDeleteConfirmationModel(itemName, itemType string, details []string) *ConfirmationModel {
	title := fmt.Sprintf("Delete %s", itemType)
	message := fmt.Sprintf("Are you sure you want to delete '%s'?", itemName)
	
	return &ConfirmationModel{
		title:       title,
		message:     message,
		details:     details,
		confirmText: "Yes, delete",
		cancelText:  "Cancel",
		selected:    0, // Default to cancel
	}
}

// Init implements tea.Model
func (m *ConfirmationModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ConfirmationModel) Update(msg tea.Msg) (*ConfirmationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.selected = 0 // Cancel
		case "right", "l":
			m.selected = 1 // Confirm
		case "tab":
			m.selected = (m.selected + 1) % 2
		case "shift+tab":
			m.selected = (m.selected - 1 + 2) % 2
		case "enter":
			if m.selected == 1 {
				m.confirmed = true
				return m, func() tea.Msg {
					return ConfirmationResult{Confirmed: true}
				}
			} else {
				m.cancelled = true
				return m, func() tea.Msg {
					return ConfirmationResult{Confirmed: false}
				}
			}
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, func() tea.Msg {
				return ConfirmationResult{Confirmed: false}
			}
		case "y":
			m.confirmed = true
			return m, func() tea.Msg {
				return ConfirmationResult{Confirmed: true}
			}
		case "n":
			m.cancelled = true
			return m, func() tea.Msg {
				return ConfirmationResult{Confirmed: false}
			}
		}
	}
	
	return m, nil
}

// View implements tea.Model
func (m *ConfirmationModel) View() string {
	// Calculate dialog dimensions
	dialogWidth := 60
	if m.width > 0 && m.width < dialogWidth+10 {
		dialogWidth = m.width - 10
	}
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	
	// Dialog border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Width(dialogWidth)
	
	var content strings.Builder
	
	// Title
	if m.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Align(lipgloss.Center).
			Width(dialogWidth - 4)
		content.WriteString(titleStyle.Render(m.title))
		content.WriteString("\n\n")
	}
	
	// Message
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Width(dialogWidth - 4).
			Align(lipgloss.Center)
		content.WriteString(messageStyle.Render(m.message))
		content.WriteString("\n")
	}
	
	// Details
	if len(m.details) > 0 {
		content.WriteString("\n")
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(dialogWidth - 4)
		
		for _, detail := range m.details {
			content.WriteString(detailStyle.Render("• " + detail))
			content.WriteString("\n")
		}
	}
	
	// Warning
	content.WriteString("\n")
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true).
		Align(lipgloss.Center).
		Width(dialogWidth - 4)
	content.WriteString(warningStyle.Render("⚠️  This action cannot be undone"))
	content.WriteString("\n\n")
	
	// Buttons
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)
	confirmStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)
	
	if m.selected == 0 {
		// Cancel is selected
		cancelStyle = cancelStyle.
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("7"))
		confirmStyle = confirmStyle.
			Foreground(lipgloss.Color("241"))
	} else {
		// Confirm is selected
		cancelStyle = cancelStyle.
			Foreground(lipgloss.Color("241"))
		confirmStyle = confirmStyle.
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("196"))
	}
	
	cancelButton := cancelStyle.Render(m.cancelText)
	confirmButton := confirmStyle.Render(m.confirmText)
	
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		cancelButton,
		"  ",
		confirmButton,
	)
	
	buttonContainer := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(buttonContainer.Render(buttons))
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(dialogWidth - 4)
	content.WriteString(helpStyle.Render("[←→] navigate  [enter] select  [y/n] quick choice  [esc] cancel"))
	
	dialog := borderStyle.Render(content.String())
	
	// Center the dialog on screen
	if m.height > 0 {
		dialogHeight := strings.Count(dialog, "\n") + 1
		topPadding := (m.height - dialogHeight) / 2
		if topPadding > 0 {
			dialog = strings.Repeat("\n", topPadding) + dialog
		}
	}
	
	if m.width > 0 && dialogWidth < m.width {
		leftPadding := (m.width - dialogWidth) / 2
		if leftPadding > 0 {
			lines := strings.Split(dialog, "\n")
			for i, line := range lines {
				if strings.TrimSpace(line) != "" {
					lines[i] = strings.Repeat(" ", leftPadding) + line
				}
			}
			dialog = strings.Join(lines, "\n")
		}
	}
	
	return dialog
}

// SetSize updates the model size
func (m *ConfirmationModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetConfirmText sets custom text for the confirm button
func (m *ConfirmationModel) SetConfirmText(text string) {
	m.confirmText = text
}

// SetCancelText sets custom text for the cancel button
func (m *ConfirmationModel) SetCancelText(text string) {
	m.cancelText = text
}

// IsConfirmed returns true if the user confirmed the action
func (m *ConfirmationModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled the action
func (m *ConfirmationModel) IsCancelled() bool {
	return m.cancelled
}