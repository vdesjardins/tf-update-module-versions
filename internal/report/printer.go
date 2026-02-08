package report

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/vdesjardins/terraform-module-versions/internal/color"
)

// Printer handles output to console
type Printer struct {
	summary *UpdateSummary
	color   *color.ColoredOutput
}

// NewPrinter creates a new report printer
func NewPrinter(summary *UpdateSummary) *Printer {
	return &Printer{summary: summary, color: color.New()}
}

// Print outputs the summary to the provided writer.
func (p *Printer) Print(writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintln(writer, p.color.Sprintf(color.BoldCyan, "\n╔════════════════════════════════════════════════════════════╗"))
	fmt.Fprintln(writer, p.color.Sprintf(color.BoldCyan, "║  Terraform Module Update Summary                           ║"))
	fmt.Fprintln(writer, p.color.Sprintf(color.BoldCyan, "╚════════════════════════════════════════════════════════════╝"))

	// Overview
	fmt.Fprintln(writer, p.color.Sprintf(color.BoldBlue, "\nModule Status Overview"))
	fmt.Fprintln(writer, p.color.Sprintf(color.Blue, "─────────────────────"))
	fmt.Fprintf(writer, "  Supported Modules:        %d\n", p.summary.SuportedCount)
	fmt.Fprintf(writer, "  Unsupported Modules:      %d\n", p.summary.UnsupportedCount)
	fmt.Fprintf(writer, "  Total Module Usages:      %d\n", p.summary.TotalUsages)

	// Detailed reports for supported modules
	if len(p.summary.Modules) > 0 {
		fmt.Fprintln(writer, p.color.Sprintf(color.BoldBlue, "\nSupported Modules"))
		fmt.Fprintln(writer, p.color.Sprintf(color.Blue, "─────────────────"))
		for _, mod := range p.summary.Modules {
			p.printModuleReport(writer, &mod)
		}
	}

	// Unsupported modules
	if len(p.summary.UnsupportedModules) > 0 {
		fmt.Fprintln(writer, p.color.Sprintf(color.BoldYellow, "\nUnsupported Modules"))
		fmt.Fprintln(writer, p.color.Sprintf(color.Yellow, "───────────────────"))
		for _, unsup := range p.summary.UnsupportedModules {
			fmt.Fprintf(writer, "\n%s\n", p.color.Error("✗ %s (%s)", unsup.Source, unsup.Type.String()))
			fmt.Fprintf(writer, "  Total Usages: %d\n", unsup.Count)
			fmt.Fprintf(writer, "  Status:       %s\n", p.color.Warning("NOT SUPPORTED (future enhancement)"))
		}
		fmt.Fprintln(writer)
	}

	// Summary stats
	fmt.Fprintln(writer, p.color.Sprintf(color.BoldBlue, "\nSummary"))
	fmt.Fprintln(writer, p.color.Sprintf(color.Blue, "───────"))
	fmt.Fprintf(writer, "  Total Module Invocations:           %d\n", p.summary.TotalUsages)
	fmt.Fprintf(writer, "  Module Invocations to Update:       %d\n", p.summary.TotalUpdated)
	fmt.Fprintf(writer, "  Module Invocations Already Latest:  %d\n", p.summary.TotalUsages-p.summary.TotalUpdated)

	// Version change details
	if len(p.summary.ByVersionChange) > 0 {
		fmt.Fprintln(writer, p.color.Sprintf(color.BoldBlue, "\nVersion Changes Summary"))
		fmt.Fprintln(writer, p.color.Sprintf(color.Blue, "──────────────────────"))

		// Sort keys for consistent output
		var keys []string
		for k := range p.summary.ByVersionChange {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			count := p.summary.ByVersionChange[key]
			fmt.Fprintf(writer, "  %s (%d changes)\n", p.color.Info("%s", key), count)
		}
	}

	fmt.Fprintln(writer)
}

func (p *Printer) printModuleReport(writer io.Writer, mod *ModuleReport) {
	fmt.Fprintf(writer, "\n%s\n", p.color.Success("✓ %s (%s)", mod.Source, mod.Type.String()))

	// Current versions
	fmt.Fprint(writer, "  Current Versions:  ")
	var versionLines []string
	for v, count := range mod.CurrentVersions {
		versionLines = append(versionLines, fmt.Sprintf("%s (%d)", v, count))
	}
	sort.Strings(versionLines)
	fmt.Fprintln(writer, strings.Join(versionLines, ", "))

	fmt.Fprintf(writer, "  Latest Version:    %s\n", p.color.Info("%s", mod.LatestVersion))
	fmt.Fprintf(writer, "  Modules to Update: %s\n", p.color.Status("%d", mod.UpdateCount))

	if mod.UpdateCount > 0 {
		fmt.Fprintf(writer, "  Status:            %s\n", p.color.Warning("UPDATE AVAILABLE"))
	} else {
		fmt.Fprintf(writer, "  Status:            %s\n", p.color.Success("ALREADY AT LATEST"))
	}

	// Locations (if tracking was enabled)
	if len(mod.Locations) > 0 && len(mod.Locations) <= 5 {
		fmt.Fprintln(writer, "  Files:")
		for _, loc := range mod.Locations {
			fmt.Fprintf(writer, "    - %s\n", loc)
		}
	} else if len(mod.Locations) > 5 {
		fmt.Fprintf(writer, "  Files: %d files (...)\n", len(mod.Locations))
	}
}

// PrintError prints an error message
func PrintError(message string) {
	colored := color.New()
	fmt.Fprintf(os.Stderr, "\n%s\n\n", colored.Error("❌ Error: %s", message))
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	colored := color.New()
	fmt.Printf("\n%s\n\n", colored.Success("✅ %s", message))
}
