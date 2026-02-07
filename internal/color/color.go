package color

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// Color represents ANSI color codes
type Color string

const (
	// Foreground colors
	Red    Color = "\033[31m"
	Green  Color = "\033[32m"
	Yellow Color = "\033[33m"
	Blue   Color = "\033[34m"
	Cyan   Color = "\033[36m"
	Reset  Color = "\033[0m"

	// Bold variants
	BoldRed    Color = "\033[1;31m"
	BoldGreen  Color = "\033[1;32m"
	BoldYellow Color = "\033[1;33m"
	BoldBlue   Color = "\033[1;34m"
	BoldCyan   Color = "\033[1;36m"
)

// ColoredOutput manages colored output with TTY detection
type ColoredOutput struct {
	enabled bool
}

// New creates a new ColoredOutput instance
// It detects if stdout/stderr is a TTY and enables colors accordingly
func New() *ColoredOutput {
	enabled := isTTY(os.Stdout) && isTTY(os.Stderr)

	// Check NO_COLOR environment variable (standard for disabling colors)
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
	}

	// Check CLICOLOR environment variable (allow override)
	if os.Getenv("CLICOLOR") == "0" {
		enabled = false
	}

	return &ColoredOutput{enabled: enabled}
}

// isTTY checks if the file descriptor is a TTY
func isTTY(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}

// Sprintf returns a formatted string with color codes if colors are enabled
func (c *ColoredOutput) Sprintf(color Color, format string, args ...interface{}) string {
	if !c.enabled {
		return fmt.Sprintf(format, args...)
	}
	return fmt.Sprintf("%s%s%s", color, fmt.Sprintf(format, args...), Reset)
}

// Fprintf writes formatted colored output to the provided writer
func (c *ColoredOutput) Fprintf(w io.Writer, color Color, format string, args ...interface{}) (int, error) {
	if !c.enabled {
		return fmt.Fprintf(w, format, args...)
	}
	return fmt.Fprintf(w, "%s%s%s", color, fmt.Sprintf(format, args...), Reset)
}

// Success prints a success message in green
func (c *ColoredOutput) Success(format string, args ...interface{}) string {
	return c.Sprintf(BoldGreen, format, args...)
}

// Error prints an error message in bold red
func (c *ColoredOutput) Error(format string, args ...interface{}) string {
	return c.Sprintf(BoldRed, format, args...)
}

// Warning prints a warning message in bold yellow
func (c *ColoredOutput) Warning(format string, args ...interface{}) string {
	return c.Sprintf(BoldYellow, format, args...)
}

// Info prints an info message in cyan
func (c *ColoredOutput) Info(format string, args ...interface{}) string {
	return c.Sprintf(Cyan, format, args...)
}

// Status prints a status message in blue
func (c *ColoredOutput) Status(format string, args ...interface{}) string {
	return c.Sprintf(Blue, format, args...)
}

// Enabled returns whether colors are currently enabled
func (c *ColoredOutput) Enabled() bool {
	return c.enabled
}

// Strip removes all color codes from a string
func Strip(s string) string {
	// Remove ANSI escape codes
	replacer := strings.NewReplacer(
		"\033[31m", "",
		"\033[32m", "",
		"\033[33m", "",
		"\033[34m", "",
		"\033[36m", "",
		"\033[1;31m", "",
		"\033[1;32m", "",
		"\033[1;33m", "",
		"\033[1;34m", "",
		"\033[1;36m", "",
		"\033[0m", "",
	)
	return replacer.Replace(s)
}
