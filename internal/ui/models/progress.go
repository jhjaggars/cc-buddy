package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel displays progress for long-running operations
type ProgressModel struct {
	progress    progress.Model
	title       string
	steps       []ProgressStep
	currentStep int
	width       int
	height      int
	startTime   time.Time
	completed   bool
	err         error
}

// ProgressStep represents a single step in a multi-step operation
type ProgressStep struct {
	Name        string
	Description string
	Status      StepStatus
	Progress    float64
	Error       error
}

// StepStatus represents the status of a progress step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepInProgress
	StepCompleted
	StepFailed
)

// ProgressUpdateMsg is sent to update progress
type ProgressUpdateMsg struct {
	StepIndex   int
	Progress    float64
	Description string
	Error       error
	Completed   bool
}

// NewProgressModel creates a new progress model
func NewProgressModel(title string, steps []string) *ProgressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50
	
	progressSteps := make([]ProgressStep, len(steps))
	for i, step := range steps {
		progressSteps[i] = ProgressStep{
			Name:   step,
			Status: StepPending,
		}
	}
	
	return &ProgressModel{
		progress:  p,
		title:     title,
		steps:     progressSteps,
		startTime: time.Now(),
	}
}

// Init implements tea.Model
func (m *ProgressModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ProgressModel) Update(msg tea.Msg) (*ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update progress bar width
		progressWidth := m.width - 20
		if progressWidth < 20 {
			progressWidth = 20
		}
		m.progress.Width = progressWidth
		
	case ProgressUpdateMsg:
		if msg.StepIndex >= 0 && msg.StepIndex < len(m.steps) {
			step := &m.steps[msg.StepIndex]
			
			if msg.Error != nil {
				step.Status = StepFailed
				step.Error = msg.Error
				m.err = msg.Error
			} else if msg.Completed {
				step.Status = StepCompleted
				step.Progress = 1.0
				m.completed = msg.Completed
			} else {
				step.Status = StepInProgress
				step.Progress = msg.Progress
				if msg.Description != "" {
					step.Description = msg.Description
				}
			}
			
			m.currentStep = msg.StepIndex
		}
		
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" && !m.completed {
			// TODO: Implement cancellation
			return m, tea.Quit
		}
	}
	
	// Update progress bar
	var cmd tea.Cmd
	if m.currentStep < len(m.steps) {
		cmd = m.progress.SetPercent(m.steps[m.currentStep].Progress)
	}
	
	return m, cmd
}

// View implements tea.Model
func (m *ProgressModel) View() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)
	
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")
	
	// Steps
	for i, step := range m.steps {
		b.WriteString(m.renderStep(i, step))
		
		// Show progress bar for current step
		if i == m.currentStep && step.Status == StepInProgress {
			b.WriteString("\n")
			b.WriteString(m.progress.View())
			
			// Show description if available
			if step.Description != "" {
				b.WriteString("\n")
				descStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				b.WriteString(descStyle.Render(step.Description))
			}
		}
		
		b.WriteString("\n")
	}
	
	// Footer
	b.WriteString("\n")
	
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n[enter] retry  [ctrl+c] cancel")
	} else if m.completed {
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
		elapsed := time.Since(m.startTime).Round(time.Second)
		b.WriteString(successStyle.Render(fmt.Sprintf("✅ Completed successfully in %v", elapsed)))
		b.WriteString("\n\n[enter] continue")
	} else {
		elapsed := time.Since(m.startTime).Round(time.Second)
		b.WriteString(fmt.Sprintf("Elapsed: %v", elapsed))
		b.WriteString("\n\n[ctrl+c] cancel")
	}
	
	return b.String()
}

// SetSize updates the model size
func (m *ProgressModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	
	// Update progress bar width
	progressWidth := width - 20
	if progressWidth < 20 {
		progressWidth = 20
	}
	m.progress.Width = progressWidth
}

// renderStep renders a single progress step
func (m *ProgressModel) renderStep(index int, step ProgressStep) string {
	var icon string
	var style lipgloss.Style
	
	switch step.Status {
	case StepPending:
		icon = "○"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	case StepInProgress:
		icon = "⟳"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	case StepCompleted:
		icon = "✓"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	case StepFailed:
		icon = "✗"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}
	
	text := step.Name
	if step.Status == StepFailed && step.Error != nil {
		text = fmt.Sprintf("%s - %v", step.Name, step.Error)
	}
	
	return fmt.Sprintf("  %s %s", 
		style.Render(icon), 
		style.Render(text))
}

// UpdateStep sends a progress update for a specific step
func (m *ProgressModel) UpdateStep(stepIndex int, progress float64, description string) tea.Cmd {
	return func() tea.Msg {
		return ProgressUpdateMsg{
			StepIndex:   stepIndex,
			Progress:    progress,
			Description: description,
		}
	}
}

// CompleteStep marks a step as completed
func (m *ProgressModel) CompleteStep(stepIndex int) tea.Cmd {
	return func() tea.Msg {
		return ProgressUpdateMsg{
			StepIndex: stepIndex,
			Progress:  1.0,
			Completed: true,
		}
	}
}

// FailStep marks a step as failed
func (m *ProgressModel) FailStep(stepIndex int, err error) tea.Cmd {
	return func() tea.Msg {
		return ProgressUpdateMsg{
			StepIndex: stepIndex,
			Error:     err,
		}
	}
}