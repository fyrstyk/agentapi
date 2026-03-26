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

	// Skip dynamic header lines to avoid false positives from content that
	// changes on every render (spinners, cursors, token counts, etc.).
	headerOffset := 0
	switch agentType {
	case msgfmt.AgentTypeClaude:
		// Claude Code TUI has a dynamic spinner/cursor on line 0:
		//   e.g. "    █                           e"
		// Skipping it prevents firstDivergence=0 on every render.
		headerOffset = 1
	case msgfmt.AgentTypeOpencode:
		// Opencode header contains dynamic token count and cost (2 lines):
		//   ┃  # Getting Started with Claude CLI                                   ┃
		//   ┃  /share to create a shareable link                 12.6K/6% ($0.05)  ┃
		if len(newLines) >= 2 {
			headerOffset = 2
		}
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

	// For Claude: truncate at the input box boundary.
	// Claude Code's TUI renders a full-width ─ divider above the ❯ prompt,
	// followed by old buffer content below the visible area. Everything from
	// that divider onwards is UI chrome or stale buffer artifacts, not content.
	// The divider is a line consisting entirely of ─ (U+2500) characters.
	if agentType == msgfmt.AgentTypeClaude {
		for i, line := range newSectionLines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) >= 40 && strings.Count(trimmed, "─") == len(trimmed) {
				newSectionLines = newSectionLines[:i]
				break
			}
		}
	}

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
