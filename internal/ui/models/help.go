package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel displays context-sensitive help
type HelpModel struct {
	context HelpContext
	width   int
	height  int
	visible bool
}

// HelpContext represents different help contexts
type HelpContext int

const (
	MainHelpContext HelpContext = iota
	ListHelpContext
	CreateHelpContext
	ProgressHelpContext
	ConfirmationHelpContext
)

// HelpEntry represents a single help item
type HelpEntry struct {
	Key         string
	Description string
}

// NewHelpModel creates a new help model
func NewHelpModel() *HelpModel {
	return &HelpModel{
		context: MainHelpContext,
		visible: false,
	}
}

// Init implements tea.Model
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *HelpModel) Update(msg tea.Msg) (*HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "h", "help":
			m.visible = !m.visible
		case "esc":
			if m.visible {
				m.visible = false
			}
		}
	}
	
	return m, nil
}

// View implements tea.Model
func (m *HelpModel) View() string {
	if !m.visible {
		return ""
	}
	
	// Get help entries for current context
	entries := m.getHelpEntries()
	
	// Calculate dialog dimensions
	dialogWidth := 60
	if m.width > 0 && m.width < dialogWidth+10 {
		dialogWidth = m.width - 10
	}
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	
	// Build content
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Width(dialogWidth - 4)
	content.WriteString(titleStyle.Render("Help - " + m.getContextName()))
	content.WriteString("\n\n")
	
	// Help entries
	maxKeyWidth := 0
	for _, entry := range entries {
		if len(entry.Key) > maxKeyWidth {
			maxKeyWidth = len(entry.Key)
		}
	}
	
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true).
		Width(maxKeyWidth)
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))
	
	for _, entry := range entries {
		key := keyStyle.Render(entry.Key)
		desc := descStyle.Render(entry.Description)
		content.WriteString(fmt.Sprintf("  %s  %s\n", key, desc))
	}
	
	// Footer
	content.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(dialogWidth - 4)
	content.WriteString(footerStyle.Render("[?] toggle help  [esc] close"))
	
	// Style the dialog
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Width(dialogWidth)
	
	dialog := borderStyle.Render(content.String())
	
	// Center the dialog
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

// SetContext sets the help context
func (m *HelpModel) SetContext(context HelpContext) {
	m.context = context
}

// SetSize updates the model size
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsVisible returns true if help is currently visible
func (m *HelpModel) IsVisible() bool {
	return m.visible
}

// Show displays the help overlay
func (m *HelpModel) Show() {
	m.visible = true
}

// Hide hides the help overlay
func (m *HelpModel) Hide() {
	m.visible = false
}

// getContextName returns the name of the current context
func (m *HelpModel) getContextName() string {
	switch m.context {
	case MainHelpContext:
		return "Main View"
	case ListHelpContext:
		return "Environment List"
	case CreateHelpContext:
		return "Create Environment"
	case ProgressHelpContext:
		return "Progress View"
	case ConfirmationHelpContext:
		return "Confirmation Dialog"
	default:
		return "General"
	}
}

// getHelpEntries returns help entries for the current context
func (m *HelpModel) getHelpEntries() []HelpEntry {
	switch m.context {
	case MainHelpContext, ListHelpContext:
		return []HelpEntry{
			{"↑↓", "Navigate environments"},
			{"enter", "Open terminal in environment"},
			{"n", "Create new environment"},
			{"d", "Delete selected environment"},
			{"r", "Refresh environment list"},
			{"q", "Quit application"},
			{"ctrl+c", "Interrupt/Quit"},
			{"?", "Toggle this help"},
		}
		
	case CreateHelpContext:
		return []HelpEntry{
			{"tab", "Next field"},
			{"shift+tab", "Previous field"},
			{"space", "Select option"},
			{"enter", "Continue/Create"},
			{"esc", "Cancel creation"},
			{"?", "Toggle this help"},
		}
		
	case ProgressHelpContext:
		return []HelpEntry{
			{"ctrl+c", "Cancel operation"},
			{"enter", "Continue (when completed)"},
			{"?", "Toggle this help"},
		}
		
	case ConfirmationHelpContext:
		return []HelpEntry{
			{"←→", "Navigate options"},
			{"tab", "Navigate options"},
			{"enter", "Select option"},
			{"y", "Confirm"},
			{"n", "Cancel"},
			{"esc", "Cancel"},
			{"?", "Toggle this help"},
		}
		
	default:
		return []HelpEntry{
			{"?", "Show context-sensitive help"},
			{"q", "Quit"},
			{"ctrl+c", "Interrupt"},
		}
	}
}