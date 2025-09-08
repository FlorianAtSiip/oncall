package main

import (
	lipgloss "github.com/charmbracelet/lipgloss"
)

var (
	headerStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")) // Blue
	issueIDStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))            // Magenta
	titleStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))            // White
	statusResolvedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))            // Green
	statusUnresolvedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // Red
	levelErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))             // Red
	levelWarningStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))            // Yellow
	levelInfoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // Blue
)

var (
	paneTitleStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Padding(0, 2)
)

var (
	runningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))                                  // Green
	pendingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))                                  // Yellow
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))                                   // Red
	defaultStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))                                  // White
	highlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("19")).Foreground(lipgloss.Color("15")) // Blue background, white text
)

var (
	logViewerHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Padding(0, 1).Bold(true)
	logViewerFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)
)
