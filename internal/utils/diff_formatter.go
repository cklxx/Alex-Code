package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Diff background color scheme using lipgloss  
	addedLineBgColor   = lipgloss.Color("#2d5016")   // Dark green background for added lines
	removedLineBgColor = lipgloss.Color("#5d1a1d")   // Dark red background for deleted lines
	contextLineBgColor = lipgloss.Color("#1a1a1a")   // Dark background for context lines
	headerLineColor    = lipgloss.Color("#8b5cf6")   // Purple for diff headers
	
	// Styles for different diff line types - using background colors
	addedLineStyle   = lipgloss.NewStyle().Background(addedLineBgColor)
	removedLineStyle = lipgloss.NewStyle().Background(removedLineBgColor)
	contextLineStyle = lipgloss.NewStyle().Background(contextLineBgColor)
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

		// Check for our new line number format: "  123 +      content" or "  123 -      content"
		if len(line) > 7 && line[0] >= '0' && line[0] <= '9' {
			// Look for the +/- indicator after the line number
			if strings.Contains(line, " +      ") {
				// Added lines (light green)
				formattedLines = append(formattedLines, addedLineStyle.Render(line))
				continue
			} else if strings.Contains(line, " -      ") {
				// Removed lines (light red)
				formattedLines = append(formattedLines, removedLineStyle.Render(line))
				continue
			} else if len(line) > 8 && line[4:12] == "        " {
				// Context lines with line numbers (slight gray tint)
				formattedLines = append(formattedLines, contextLineStyle.Render(line))
				continue
			}
		}
		
		// Fall back to original logic for traditional diff format
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