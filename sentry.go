package main

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type sentryIssue struct {
	ID       string
	ShortID  string
	Title    string
	LastSeen string
	Status   string
	Level    string
}

type sentryErrorLogsMsg string

type sentryStatsMsg string

func getSentryErrorLogsCmd() tea.Cmd {
	return func() tea.Msg {
		cmdTicketing := exec.Command("sentry-cli", "issues", "list", "--org", "siip", "--project", "siip-ticketing", "--query", "age:-24h is:unresolved")
		outputTicketing, errTicketing := cmdTicketing.CombinedOutput()
		if errTicketing != nil {
			return errMsg(fmt.Errorf("failed to get ticketing sentry logs: %w", errTicketing))
		}

		cmdIam := exec.Command("sentry-cli", "issues", "list", "--org", "siip", "--project", "siip-iam-service", "--query", "age:-24h is:unresolved")
		outputIam, errIam := cmdIam.CombinedOutput()
		if errIam != nil {
			return errMsg(fmt.Errorf("failed to get IAM sentry logs: %w", errIam))
		}

		parsedTicketing := parseSentryIssues(string(outputTicketing))
		parsedIam := parseSentryIssues(string(outputIam))

		formattedOutput := formatSentryIssues(parsedTicketing) + "\n\n" + formatSentryIssues(parsedIam)

		return sentryErrorLogsMsg(formattedOutput)
	}
}

func getSentryStatsCmd() tea.Cmd {
	return func() tea.Msg {
		cmdTicketing := exec.Command("sentry-cli", "issues", "list", "--org", "siip", "--project", "siip-ticketing")
		outputTicketingBytes, errTicketing := cmdTicketing.CombinedOutput()
		if errTicketing != nil {
			return errMsg(fmt.Errorf("failed to get ticketing sentry issues: %w", errTicketing))
		}
		ticketingIssueCount := countNonEmptyLines(string(outputTicketingBytes))

		cmdIam := exec.Command("sentry-cli", "issues", "list", "--org", "siip", "--project", "siip-iam-service")
		outputIamBytes, errIam := cmdIam.CombinedOutput()
		if errIam != nil {
			return errMsg(fmt.Errorf("failed to get IAM sentry issues: %w", errIam))
		}
		iamIssueCount := countNonEmptyLines(string(outputIamBytes))

		statsOutput := fmt.Sprintf("Ticketing Issues (total): %d\nIAM Issues (total): %d",
			ticketingIssueCount,
			iamIssueCount,
		)
		return sentryStatsMsg(statsOutput)
	}
}

func parseSentryIssues(output string) []sentryIssue {
	lines := strings.Split(output, "\n")
	var issues []sentryIssue

	headerLineIdx := -1
	separatorLineIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Issue ID") && strings.Contains(line, "Title") {
			headerLineIdx = i
		} else if strings.HasPrefix(line, "+") && headerLineIdx != -1 {
			separatorLineIdx = i
			break
		}
	}

	if headerLineIdx == -1 || separatorLineIdx == -1 {
		return issues
	}

	header := lines[headerLineIdx]

	issueIDStart := strings.Index(header, "Issue ID")
	shortIDStart := strings.Index(header, "Short ID")
	titleStart := strings.Index(header, "Title")
	lastSeenStart := strings.Index(header, "Last seen")
	statusStart := strings.Index(header, "Status")
	levelStart := strings.Index(header, "Level")

	columnStarts := []int{issueIDStart, shortIDStart, titleStart, lastSeenStart, statusStart, levelStart}
	var validColumnStarts []int
	for _, cs := range columnStarts {
		if cs != -1 {
			validColumnStarts = append(validColumnStarts, cs)
		}
	}
	sort.Ints(validColumnStarts)

	for i := separatorLineIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "+") || strings.TrimSpace(line) == "" {
			continue
		}

		issue := sentryIssue{}

		if issueIDStart != -1 && len(line) > issueIDStart {
			issue.ID = strings.TrimSpace(extractColumn(line, issueIDStart, shortIDStart, validColumnStarts))
		}
		if shortIDStart != -1 && len(line) > shortIDStart {
			issue.ShortID = strings.TrimSpace(extractColumn(line, shortIDStart, titleStart, validColumnStarts))
		}
		if titleStart != -1 && len(line) > titleStart {
			issue.Title = strings.TrimSpace(extractColumn(line, titleStart, lastSeenStart, validColumnStarts))
		}
		if lastSeenStart != -1 && len(line) > lastSeenStart {
			issue.LastSeen = strings.TrimSpace(extractColumn(line, lastSeenStart, statusStart, validColumnStarts))
		}
		if statusStart != -1 && len(line) > statusStart {
			issue.Status = strings.TrimSpace(extractColumn(line, statusStart, levelStart, validColumnStarts))
		}
		if levelStart != -1 && len(line) > levelStart {
			issue.Level = strings.TrimSpace(line[levelStart:])
		}

		issues = append(issues, issue)
	}
	return issues
}

func extractColumn(line string, start int, nextStart int, validColumnStarts []int) string {
	actualEnd := len(line)
	for _, colStart := range validColumnStarts {
		if colStart > start {
			actualEnd = colStart
			break
		}
	}
	if start >= len(line) {
		return ""
	}
	end := actualEnd
	if nextStart != -1 && nextStart < end {
		end = nextStart
	}
	if end > len(line) {
		end = len(line)
	}
	return line[start:end]
}

func formatSentryIssues(issues []sentryIssue) string {
	var formatted []string
	formatted = append(formatted, headerStyle.Render("Recent Sentry Issues:"))
	if len(issues) == 0 {
		formatted = append(formatted, "  No unresolved issues found.")
		return strings.Join(formatted, "\n")
	}
	for _, issue := range issues {
		statusStyle := defaultStyle
		switch issue.Status {
		case "resolved", "ignored":
			statusStyle = statusResolvedStyle
		case "unresolved":
			statusStyle = statusUnresolvedStyle
		}
		levelStyle := defaultStyle
		switch issue.Level {
		case "error", "fatal":
			levelStyle = levelErrorStyle
		case "warning":
			levelStyle = levelWarningStyle
		case "info", "debug":
			levelStyle = levelInfoStyle
		}
		formatted = append(formatted, fmt.Sprintf(
			"  %s %s | %s | %s | %s",
			issueIDStyle.Render(issue.ShortID),
			titleStyle.Render(issue.Title),
			lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(issue.LastSeen),
			statusStyle.Render(issue.Status),
			levelStyle.Render(issue.Level),
		))
	}
	return strings.Join(formatted, "\n")
}
