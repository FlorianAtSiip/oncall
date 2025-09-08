package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type apiResponseTimesMsg string

func getApiResponseTimesCmd() tea.Cmd {
	return func() tea.Msg {
		apiEndpoints := map[string]string{
			"Ticketing API": "https://ticketing.siip.io/health",
			"IAM API":       "https://iam.siip.io/health",
		}

		var results []string
		for name, url := range apiEndpoints {
			// Capture body separately, then measure time
			bodyCmd := exec.Command("curl", "-s", url)
			bodyBytes, bodyErr := bodyCmd.CombinedOutput()
			if bodyErr != nil {
				results = append(results, fmt.Sprintf("%s: Error - %v", name, bodyErr))
				continue
			}
			body := strings.TrimSpace(string(bodyBytes))

			timeCmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{time_total}", url)
			timeBytes, timeErr := timeCmd.CombinedOutput()
			if timeErr != nil {
				results = append(results, fmt.Sprintf("%s: Error - %v", name, timeErr))
				continue
			}
			secStr := strings.TrimSpace(string(timeBytes))
			sec, _ := strconv.ParseFloat(secStr, 64)
			ms := int(sec * 1000.0)

			formatted := fmt.Sprintf("%s: %dms", name, ms)

			if name == "Ticketing API" {
				// Robust status detection without JSON lib
				status := extractJsonValue(body, "status")
				lower := strings.ToLower(body)
				okDetected := strings.EqualFold(status, "ok") || strings.EqualFold(status, "up") || strings.Contains(lower, `"status":"ok"`) || strings.Contains(lower, `"status":"up"`)
				failDetected := strings.EqualFold(status, "fail") || strings.EqualFold(status, "down") || strings.Contains(lower, `"status":"fail"`) || strings.Contains(lower, `"status":"down"`)
				statusText := "unknown"
				statusStyle := defaultStyle
				if okDetected {
					statusText = "OK"
					statusStyle = statusResolvedStyle
				} else if failDetected || status != "" {
					statusText = "FAIL"
					statusStyle = statusUnresolvedStyle
				}
				formatted += "\n  General: " + statusStyle.Render(statusText)
			} else if name == "IAM API" {
				// Overall status
				status := extractJsonValue(body, "status")
				if status != "" {
					statusStyle := defaultStyle
					if strings.EqualFold(status, "ok") || strings.EqualFold(status, "up") {
						statusStyle = statusResolvedStyle
					} else {
						statusStyle = statusUnresolvedStyle
					}
					formatted += "\n  Status: " + statusStyle.Render(strings.ToUpper(status))
				}
				// Groups list from root
				groups := extractJsonValue(body, "groups")
				var groupNames []string
				if groups != "" {
					raw := strings.TrimSpace(groups)
					raw = strings.TrimPrefix(raw, "[")
					raw = strings.TrimSuffix(raw, "]")
					parts := strings.Split(raw, ",")
					for i := range parts {
						name := strings.TrimSpace(parts[i])
						name = strings.Trim(name, `"`)
						name = strings.Trim(name, "]}")
						if name != "" {
							groupNames = append(groupNames, name)
						}
					}
				}
				if len(groupNames) > 0 {
					formatted += "\n  Groups: " + levelInfoStyle.Render(strings.Join(groupNames, ", "))
					// Fetch each group endpoint and list its status
					for _, g := range groupNames {
						gURL := url + "/" + g
						gBodyCmd := exec.Command("curl", "-s", gURL)
						gBodyBytes, _ := gBodyCmd.CombinedOutput()
						gBody := strings.TrimSpace(string(gBodyBytes))
						gStatus := extractJsonValue(gBody, "status")
						gLower := strings.ToLower(gBody)
						gOK := strings.EqualFold(gStatus, "ok") || strings.EqualFold(gStatus, "up") || strings.Contains(gLower, `"status":"ok"`) || strings.Contains(gLower, `"status":"up"`)
						gStyle := statusUnresolvedStyle
						gText := strings.ToUpper(strings.TrimSpace(gStatus))
						if gOK {
							gStyle = statusResolvedStyle
							if gText == "" {
								gText = "OK"
							}
						} else if gText == "" {
							gText = "FAIL"
						}
						formatted += "\n    - " + g + ": " + gStyle.Render(gText)
					}
				}
			}

			results = append(results, formatted)
		}
		return apiResponseTimesMsg(strings.Join(results, "\n"))
	}
}

// Robust JSON extractor for simple key lookup without using encoding/json.
// Handles string scalars, numbers, booleans, and balanced []/{} values.
func extractJsonValue(jsonStr, key string) string {
	searchKey := fmt.Sprintf(`"%s":`, key)
	idx := strings.Index(jsonStr, searchKey)
	if idx == -1 {
		return ""
	}
	i := idx + len(searchKey)
	// skip spaces
	for i < len(jsonStr) && (jsonStr[i] == ' ' || jsonStr[i] == '\t' || jsonStr[i] == '\n' || jsonStr[i] == '\r') {
		i++
	}
	if i >= len(jsonStr) {
		return ""
	}
	// String value
	if jsonStr[i] == '"' {
		i++
		start := i
		for i < len(jsonStr) {
			if jsonStr[i] == '\\' {
				i += 2
				continue
			}
			if jsonStr[i] == '"' {
				return jsonStr[start:i]
			}
			i++
		}
		return jsonStr[start:]
	}
	// Array or object
	if jsonStr[i] == '[' || jsonStr[i] == '{' {
		open := jsonStr[i]
		close := byte(']')
		if open == '{' {
			close = '}'
		}
		depth := 0
		start := i
		for i < len(jsonStr) {
			c := jsonStr[i]
			if c == '"' {
				// skip quoted sections
				i++
				for i < len(jsonStr) {
					if jsonStr[i] == '\\' {
						i += 2
						continue
					}
					if jsonStr[i] == '"' {
						break
					}
					i++
				}
			} else {
				if c == open {
					depth++
				}
				if c == close {
					depth--
				}
				if depth == 0 {
					// include closing bracket and stop
					return strings.TrimSpace(jsonStr[start : i+1])
				}
			}
			i++
		}
		return strings.TrimSpace(jsonStr[start:])
	}
	// Number/boolean/null: read until comma or '}'
	start := i
	for i < len(jsonStr) && jsonStr[i] != ',' && jsonStr[i] != '}' && jsonStr[i] != '\n' && jsonStr[i] != '\r' {
		i++
	}
	return strings.TrimSpace(jsonStr[start:i])
}
