package service

import "strings"

// extractURL extracts a URL from the body text of a reminder.
func extractURL(body string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "URL: ") {
			return strings.TrimPrefix(line, "URL: ")
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line
		}
	}
	return ""
}
