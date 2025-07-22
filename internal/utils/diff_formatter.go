package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Diff color scheme using lipgloss
	addedLineColor   = lipgloss.Color("#22c55e")   // Light green for added lines
	removedLineColor = lipgloss.Color("#ef4444")   // Light red for deleted lines
	contextLineColor = lipgloss.Color("#6b7280")   // Gray for context lines
	headerLineColor  = lipgloss.Color("#8b5cf6")   // Purple for diff headers
	
	// Styles for different diff line types
	addedLineStyle   = lipgloss.NewStyle().Foreground(addedLineColor)
	removedLineStyle = lipgloss.NewStyle().Foreground(removedLineColor)
	contextLineStyle = lipgloss.NewStyle().Foreground(contextLineColor)
	headerLineStyle  = lipgloss.NewStyle().Foreground(headerLineColor).Bold(true)
)

// FormatDiffOutput applies color formatting to git diff output
func FormatDiffOutput(diffOutput string) string {
	lines := strings.Split(diffOutput, "\n")
	var formattedLines []string

	for _, line := range lines {
		if len(line) == 0 {
			formattedLines = append(formattedLines, line)
			continue
		}

		switch line[0] {
		case '+':
			// Added lines (light green)
			formattedLines = append(formattedLines, addedLineStyle.Render(line))
		case '-':
			// Removed lines (light red)
			formattedLines = append(formattedLines, removedLineStyle.Render(line))
		case '@':
			// Diff headers with line numbers (purple)
			if strings.HasPrefix(line, "@@") {
				formattedLines = append(formattedLines, headerLineStyle.Render(line))
			} else {
				formattedLines = append(formattedLines, line)
			}
		case 'd', 'i', 'n':
			// Diff command headers like "diff --git", "index", "new file mode"
			if strings.HasPrefix(line, "diff ") || 
			   strings.HasPrefix(line, "index ") || 
			   strings.HasPrefix(line, "new file mode") ||
			   strings.HasPrefix(line, "deleted file mode") ||
			   strings.HasPrefix(line, "--- ") ||
			   strings.HasPrefix(line, "+++ ") {
				formattedLines = append(formattedLines, headerLineStyle.Render(line))
			} else {
				// Context lines (no color change)
				formattedLines = append(formattedLines, line)
			}
		case ' ':
			// Context lines (slight gray tint)
			formattedLines = append(formattedLines, contextLineStyle.Render(line))
		default:
			// All other lines (no color change)
			formattedLines = append(formattedLines, line)
		}
	}

	return strings.Join(formattedLines, "\n")
}

// IsDiffOutput checks if the given output appears to be from git diff
func IsDiffOutput(output string) bool {
	diffIndicators := []string{
		"diff --git",
		"index ",
		"--- a/",
		"+++ b/",
		"@@",
	}

	for _, indicator := range diffIndicators {
		if strings.Contains(output, indicator) {
			return true
		}
	}

	return false
}