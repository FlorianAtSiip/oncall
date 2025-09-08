package main

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type kubectlPodsDataMsg struct {
	displayOutput string
	podNames      []string
}

// Message carrying current kubectl context name
type kubectlContextMsg string

func getKubectlPodsCmd() tea.Cmd {
	return func() tea.Msg {
		// Command to get clean pod names
		cmdCleanNames := exec.Command("kubectl", "get", "pods", "--no-headers", "-o", "custom-columns=NAME:.metadata.name")
		outputCleanNamesBytes, errCleanNames := cmdCleanNames.CombinedOutput()
		if errCleanNames != nil {
			return errMsg(fmt.Errorf("failed to get clean kubectl pod names: %w", errCleanNames))
		}
		cleanPodNames := strings.Split(strings.TrimSpace(string(outputCleanNamesBytes)), "\n")
		if len(cleanPodNames) == 1 && cleanPodNames[0] == "" {
			cleanPodNames = []string{}
		}

		// Command to get display output (potentially colored)
		cmdDisplay := exec.Command("kubectl", "get", "pods")
		outputDisplayBytes, errDisplay := cmdDisplay.CombinedOutput()
		if errDisplay != nil {
			return errMsg(fmt.Errorf("failed to get kubectl pods for display: %w", errDisplay))
		}
		displayOutput := string(outputDisplayBytes)

		// Colorize for display after getting clean names
		coloredDisplayOutput := colorizeKubectlPodsWithSelection(displayOutput, -1) // -1 for no initial selection highlighting

		return kubectlPodsDataMsg{
			displayOutput: coloredDisplayOutput,
			podNames:      cleanPodNames,
		}
	}
}

// Fetch currently used kubectl context by parsing `kubectl config get-contexts`
func getKubectlContextCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("kubectl", "config", "get-contexts")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errMsg(fmt.Errorf("failed to get kubectl contexts: %w", err))
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			trim := strings.TrimSpace(line)
			if trim == "" || strings.HasPrefix(trim, "CURRENT ") || strings.HasPrefix(trim, "NAME ") {
				continue
			}
			// Active context line starts with '*'
			if strings.HasPrefix(trim, "*") {
				fields := strings.Fields(trim)
				if len(fields) >= 2 {
					return kubectlContextMsg(fields[1])
				}
			}
		}
		return kubectlContextMsg("")
	}
}

func colorizeKubectlPods(output string) string {
	return colorizeKubectlPodsWithSelection(output, -1)
}

func colorizeKubectlPodsWithSelection(output string, selectedIndex int) string {
	lines := strings.Split(output, "\n")
	var coloredLines []string
	if len(lines) > 0 {
		coloredLines = append(coloredLines, lines[0])
		lines = lines[1:]
	}
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		currentLineStyle := defaultStyle
		if i == selectedIndex {
			currentLineStyle = highlightStyle
		}
		fields := strings.Fields(line)
		if len(fields) > 2 {
			status := fields[2]
			switch status {
			case "Running":
				currentLineStyle = currentLineStyle.Copy().Foreground(runningStyle.GetForeground())
			case "Pending", "ContainerCreating", "PodInitializing":
				currentLineStyle = currentLineStyle.Copy().Foreground(pendingStyle.GetForeground())
			case "Error", "Evicted", "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
				currentLineStyle = currentLineStyle.Copy().Foreground(errorStyle.GetForeground())
			}
		}
		coloredLines = append(coloredLines, currentLineStyle.Render(line))
	}
	return strings.Join(coloredLines, "\n")
}
