package report

import (
	"fmt"
	"strings"
)

// FormatUnifiedDiff creates a simple unified diff for display.
func FormatUnifiedDiff(filename, before, after string) (string, error) {
	if before == after {
		return "", nil
	}

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("--- %s\n", filename))
	builder.WriteString(fmt.Sprintf("+++ %s\n", filename))
	builder.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(beforeLines), len(afterLines)))

	maxLines := len(beforeLines)
	if len(afterLines) > maxLines {
		maxLines = len(afterLines)
	}

	for i := 0; i < maxLines; i++ {
		var beforeLine string
		var afterLine string

		if i < len(beforeLines) {
			beforeLine = beforeLines[i]
		}
		if i < len(afterLines) {
			afterLine = afterLines[i]
		}

		if beforeLine == afterLine {
			builder.WriteString(" " + beforeLine + "\n")
			continue
		}
		if beforeLine != "" {
			builder.WriteString("-" + beforeLine + "\n")
		}
		if afterLine != "" {
			builder.WriteString("+" + afterLine + "\n")
		}
	}

	return builder.String(), nil
}
