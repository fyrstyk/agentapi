package screentracker

import (
	"strings"

	"github.com/coder/agentapi/lib/msgfmt"
)

// screenDiff compares two screen states and returns new content added since oldScreen.
// Uses positional line comparison to avoid false positives from set-based matching,
// which caused old content to be included when new lines happened to match old lines
// at different positions (issue #205).
func screenDiff(oldScreen, newScreen string, agentType msgfmt.AgentType) string {
	oldLines := strings.Split(oldScreen, "\n")
	newLines := strings.Split(newScreen, "\n")

	// Skip header lines for Opencode agent type to avoid false positives.
	// The header contains dynamic content (token count, context percentage, cost)
	// that changes between screens, causing line comparison mismatches.
	headerOffset := 0
	if len(newLines) >= 2 && agentType == msgfmt.AgentTypeOpencode {
		headerOffset = 2
	}

	// Find the first line index (positional) where old and new screens diverge.
	// New content starts at this point.
	firstDivergence := len(newLines)
	for i := headerOffset; i < len(newLines); i++ {
		if i >= len(oldLines) || oldLines[i] != newLines[i] {
			firstDivergence = i
			break
		}
	}

	newSectionLines := newLines[firstDivergence:]

	if len(newSectionLines) == 0 {
		return ""
	}

	// remove leading and trailing lines which are empty or have only whitespace
	startLine := 0
	endLine := len(newSectionLines) - 1
	for i := range newSectionLines {
		if strings.TrimSpace(newSectionLines[i]) != "" {
			startLine = i
			break
		}
	}
	for i := len(newSectionLines) - 1; i >= 0; i-- {
		if strings.TrimSpace(newSectionLines[i]) != "" {
			endLine = i
			break
		}
	}
	return strings.Join(newSectionLines[startLine:endLine+1], "\n")
}
