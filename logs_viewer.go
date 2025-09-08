package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type podLogsMsg string

type podLogViewerModel struct {
	podName  string
	logs     string
	viewport viewport.Model
	ready    bool
}

func newPodLogViewerModel(podName string) podLogViewerModel {
	return podLogViewerModel{podName: podName}
}

func (m podLogViewerModel) Init() tea.Cmd {
	return nil
}

func (m podLogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(logViewerHeaderStyle.Render(" "))
		footerHeight := lipgloss.Height(logViewerFooterStyle.Render(" "))
		verticalMarginHeight := headerHeight + footerHeight
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.logs)
			m.ready = true
			m.viewport.MouseWheelEnabled = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		m.viewport.SetContent(m.logs)
	case podLogsMsg:
		m.logs = string(msg)
		m.viewport.SetContent(m.logs)
	case errMsg:
		m.logs = fmt.Sprintf("Error fetching logs: %v", msg)
		if m.ready {
			m.viewport.SetContent(m.logs)
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m podLogViewerModel) View() string {
	if !m.ready {
		return "Loading logs..."
	}
	header := logViewerHeaderStyle.Render(fmt.Sprintf("Logs for %s", m.podName))
	footer := logViewerFooterStyle.Render("Scroll with arrow keys / mouse wheel | Esc: Back")
	return lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), footer)
}

func getPodLogsCmd(podName string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("kubectl", "logs", podName, "--tail=500")
		timeout := 10 * time.Second
		timer := time.AfterFunc(timeout, func() { _ = cmd.Process.Kill() })
		defer timer.Stop()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errMsg(fmt.Errorf("failed to get logs for pod %s: %w\nKubectl Output: %s", podName, err, string(output)))
		}
		return podLogsMsg(output)
	}
}
