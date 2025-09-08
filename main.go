package main

import (
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const appVersion = "0.0.1"

type model struct {
	width            int
	height           int
	sentryLogs       string
	sentryStats      string
	kubectlPods      string
	apiResponseTimes string
	selectedPane     int // 0: Sentry Errors, 1: Analytics, 2: Pod Status
	selectedPodIndex int
	podNames         []string // To store actual pod names for logs

	logViewer     podLogViewerModel
	showLogViewer bool

	currentKubeContext     string
	podHighUsage           map[string]bool
	lastSentryErrorsUpdate time.Time

	// Splash
	showSplash      bool
	splashTimerDone bool
	initDataArrived bool
}

type tickMsg time.Time
type splashTimerMsg time.Time
type errMsg error

func (m model) Init() tea.Cmd {
	return tea.Batch(
		getSentryErrorLogsCmd(),
		getSentryStatsCmd(),
		getKubectlPodsCmd(),
		getKubectlContextCmd(),
		getApiResponseTimesCmd(),
		splashTimerCmd(),
		tickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	if m.showLogViewer {
		oldLogViewer, logCmd := m.logViewer.Update(msg)
		m.logViewer = oldLogViewer.(podLogViewerModel)
		cmds = append(cmds, logCmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.showLogViewer = false
				return m, nil
			}
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selectedPane == 2 && len(m.podNames) > 0 {
				m.selectedPodIndex--
				if m.selectedPodIndex < 0 {
					m.selectedPodIndex = len(m.podNames) - 1
				}
			}
		case "down", "j":
			if m.selectedPane == 2 && len(m.podNames) > 0 {
				m.selectedPodIndex++
				if m.selectedPodIndex >= len(m.podNames) {
					m.selectedPodIndex = 0
				}
			}
		case "l":
			if m.selectedPane == 2 && len(m.podNames) > 0 && m.selectedPodIndex < len(m.podNames) {
				selectedPod := m.podNames[m.selectedPodIndex]
				m.logViewer = newPodLogViewerModel(selectedPod)
				m.showLogViewer = true
				return m, tea.Batch(
					sendWindowSizeCmd(m.width, m.height),
					getPodLogsCmd(selectedPod),
				)
			}
		case "tab":
			m.selectedPane = (m.selectedPane + 1) % 3
			m.selectedPodIndex = 0
		case "shift+tab":
			m.selectedPane--
			if m.selectedPane < 0 {
				m.selectedPane = 2
			}
			m.selectedPodIndex = 0
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case sentryErrorLogsMsg:
		m.sentryLogs = string(msg)
		m.lastSentryErrorsUpdate = time.Now()
		m.initDataArrived = true
	case sentryStatsMsg:
		m.sentryStats = string(msg)
		m.initDataArrived = true
	case kubectlPodsDataMsg:
		m.kubectlPods = msg.displayOutput
		m.podNames = msg.podNames
		if m.selectedPodIndex >= len(m.podNames) {
			m.selectedPodIndex = 0
		}
		m.initDataArrived = true
	case kubectlContextMsg:
		m.currentKubeContext = string(msg)
	case apiResponseTimesMsg:
		m.apiResponseTimes = string(msg)
		m.initDataArrived = true
	case splashTimerMsg:
		m.splashTimerDone = true
	case tickMsg:
		// hide splash after timer and initial data
		if m.showSplash && m.splashTimerDone && m.initDataArrived {
			m.showSplash = false
		}
		batch := []tea.Cmd{
			getSentryStatsCmd(),
			getApiResponseTimesCmd(),
			getKubectlPodsCmd(),
			getKubectlContextCmd(),
		}
		if time.Since(m.lastSentryErrorsUpdate) >= 60*time.Second || m.lastSentryErrorsUpdate.IsZero() {
			batch = append(batch, getSentryErrorLogsCmd())
		}
		batch = append(batch, tickCmd())
		return m, tea.Batch(batch...)
	case errMsg:
	}
	return m, tea.Batch(cmds...)
}

// Helper to emit a WindowSizeMsg as a command
func sendWindowSizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg { return tea.WindowSizeMsg{Width: width, Height: height} }
}

func (m model) View() string {
	if m.showSplash {
		// Simple centered ASCII splash
		art := []string{
			"   ____        _____      _ _ ",
			"  / __ \\____  / ___/___  (_| )",
			" / / / / __ \\__ \\__/ / / /|/ ",
			"/ /_/ / /_/ /__/ / / /_/ /     ",
			"\\____/ .___/____/  \\__,_/      ",
			"     /_/                         ",
		}
		artStr := strings.Join(art, "\n")
		ver := "v" + appVersion
		title := paneTitleStyle.Render("On-Call")
		verStyled := levelInfoStyle.Render(ver)
		content := title + "\n" + artStr + "\n\n" + verStyled + "\n\nFetching data..."
		box := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.NormalBorder())
		if m.width > 0 && m.height > 0 {
			w := m.width - 4
			if w < 20 {
				w = m.width
			}
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box.Width(w).Render(content))
		}
		return box.Render(content)
	}

	if m.showLogViewer {
		return m.logViewer.View()
	}

	basePaneStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 2)

	focusedPaneStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2)

	keyHintsContentHeight := 3
	keyHintsTotalHeight := keyHintsContentHeight + (basePaneStyle.GetVerticalPadding() * 2) + (basePaneStyle.GetVerticalBorderSize() * 2)

	availableHeightForTopPanes := m.height - keyHintsTotalHeight

	targetHalfWidthContent := (m.width / 2) - (basePaneStyle.GetHorizontalPadding() * 2) - (basePaneStyle.GetHorizontalBorderSize() * 2)
	if targetHalfWidthContent < 0 {
		targetHalfWidthContent = 0
	}

	targetHalfHeightContent := (availableHeightForTopPanes / 2) - (basePaneStyle.GetVerticalPadding() * 2) - (basePaneStyle.GetVerticalBorderSize() * 2)
	if targetHalfHeightContent < 0 {
		targetHalfHeightContent = 0
	}

	ctxSuffix := ""
	if m.currentKubeContext != "" {
		ctxSuffix = " [" + levelInfoStyle.Render(m.currentKubeContext) + "]"
	}

	pane1Content := paneTitleStyle.Render("🛑 Recent Sentry Errors") + "\n" + m.sentryLogs
	pane2Content := paneTitleStyle.Render("📊 Analytics") + "\n" + m.sentryStats + "\n\n" + m.apiResponseTimes
	pane3Content := paneTitleStyle.Render("📦 Pod Status (Live)"+ctxSuffix) + "\n" + colorizeKubectlPodsWithSelection(m.kubectlPods, m.selectedPodIndex)
	pane4Content := "^Q: Quit | ^C: Exit | ?: Help | Tab/Shift+Tab: Switch Panes"

	var pane1, pane2, pane3 string
	if m.selectedPane == 0 {
		pane1 = focusedPaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane1Content)
		pane2 = basePaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane2Content)
		pane3 = basePaneStyle.Width(targetHalfWidthContent).Height(availableHeightForTopPanes - (basePaneStyle.GetVerticalPadding() * 2) - (basePaneStyle.GetVerticalBorderSize() * 2)).Render(pane3Content)
	} else if m.selectedPane == 1 {
		pane1 = basePaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane1Content)
		pane2 = focusedPaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane2Content)
		pane3 = basePaneStyle.Width(targetHalfWidthContent).Height(availableHeightForTopPanes - (basePaneStyle.GetVerticalPadding() * 2) - (basePaneStyle.GetVerticalBorderSize() * 2)).Render(pane3Content)
	} else if m.selectedPane == 2 {
		pane1 = basePaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane1Content)
		pane2 = basePaneStyle.Width(targetHalfWidthContent).Height(targetHalfHeightContent).Render(pane2Content)
		pane3 = focusedPaneStyle.Width(targetHalfWidthContent).Height(availableHeightForTopPanes - (basePaneStyle.GetVerticalPadding() * 2) - (basePaneStyle.GetVerticalBorderSize() * 2)).Render(pane3Content)
	}

	pane4 := basePaneStyle.Width(m.width - (basePaneStyle.GetHorizontalPadding() * 2) - (basePaneStyle.GetHorizontalBorderSize() * 2)).Height(keyHintsContentHeight).Render(pane4Content)

	leftColumn := lipgloss.JoinVertical(lipgloss.Top, pane1, pane2)
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, pane3)

	return lipgloss.JoinVertical(lipgloss.Left, topSection, pane4)
}

func countNonEmptyLines(output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "Issue") {
			count++
		}
	}
	return count
}

func tickCmd() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func splashTimerCmd() tea.Cmd {
	return tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg { return splashTimerMsg(t) })
}

func main() {
	p := tea.NewProgram(model{showSplash: true}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
