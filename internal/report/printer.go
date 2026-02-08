package report

import (
	"fmt"
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

// Print outputs the summary to console
func (p *Printer) Print() {
	fmt.Println(p.color.Sprintf(color.BoldCyan, "\n╔════════════════════════════════════════════════════════════╗"))
	fmt.Println(p.color.Sprintf(color.BoldCyan, "║  Terraform Module Update Summary                           ║"))
	fmt.Println(p.color.Sprintf(color.BoldCyan, "╚════════════════════════════════════════════════════════════╝"))

	// Overview
	fmt.Println(p.color.Sprintf(color.BoldBlue, "\nModule Status Overview"))
	fmt.Println(p.color.Sprintf(color.Blue, "─────────────────────"))
	fmt.Printf("  Supported Modules:        %d\n", p.summary.SuportedCount)
	fmt.Printf("  Unsupported Modules:      %d\n", p.summary.UnsupportedCount)
	fmt.Printf("  Total Module Usages:      %d\n", p.summary.TotalUsages)

	// Detailed reports for supported modules
	if len(p.summary.Modules) > 0 {
		fmt.Println(p.color.Sprintf(color.BoldBlue, "Supported Modules"))
		fmt.Println(p.color.Sprintf(color.Blue, "─────────────────"))
		for _, mod := range p.summary.Modules {
			p.printModuleReport(&mod)
		}
	}

	// Unsupported modules
	if len(p.summary.UnsupportedModules) > 0 {
		fmt.Println(p.color.Sprintf(color.BoldYellow, "Unsupported Modules"))
		fmt.Println(p.color.Sprintf(color.Yellow, "───────────────────"))
		for _, unsup := range p.summary.UnsupportedModules {
			fmt.Printf("\n%s\n", p.color.Error("✗ %s (%s)", unsup.Source, unsup.Type.String()))
			fmt.Printf("  Total Usages: %d\n", unsup.Count)
			fmt.Printf("  Status:       %s\n", p.color.Warning("NOT SUPPORTED (future enhancement)"))
		}
		fmt.Println()
	}

	// Summary stats
	fmt.Println(p.color.Sprintf(color.BoldBlue, "Summary"))
	fmt.Println(p.color.Sprintf(color.Blue, "───────"))
	fmt.Printf("  Total Module Invocations:           %d\n", p.summary.TotalUsages)
	fmt.Printf("  Module Invocations to Update:       %d\n", p.summary.TotalUpdated)
	fmt.Printf("  Module Invocations Already Latest:  %d\n", p.summary.TotalUsages-p.summary.TotalUpdated)

	// Version change details
	if len(p.summary.ByVersionChange) > 0 {
		fmt.Println(p.color.Sprintf(color.BoldBlue, "\nVersion Changes Summary"))
		fmt.Println(p.color.Sprintf(color.Blue, "──────────────────────"))

		// Sort keys for consistent output
		var keys []string
		for k := range p.summary.ByVersionChange {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			count := p.summary.ByVersionChange[key]
			fmt.Printf("  %s (%d changes)\n", p.color.Info("%s", key), count)
		}
	}

	fmt.Println()
}

func (p *Printer) printModuleReport(mod *ModuleReport) {
	fmt.Printf("\n%s\n", p.color.Success("✓ %s (%s)", mod.Source, mod.Type.String()))

	// Current versions
	fmt.Print("  Current Versions:  ")
	var versionLines []string
	for v, count := range mod.CurrentVersions {
		versionLines = append(versionLines, fmt.Sprintf("%s (%d)", v, count))
	}
	sort.Strings(versionLines)
	fmt.Println(strings.Join(versionLines, ", "))

	fmt.Printf("  Latest Version:    %s\n", p.color.Info("%s", mod.LatestVersion))
	fmt.Printf("  Modules to Update: %s\n", p.color.Status("%d", mod.UpdateCount))

	if mod.UpdateCount > 0 {
		fmt.Printf("  Status:            %s\n", p.color.Warning("UPDATE AVAILABLE"))
	} else {
		fmt.Printf("  Status:            %s\n", p.color.Success("ALREADY AT LATEST"))
	}

	// Locations (if tracking was enabled)
	if len(mod.Locations) > 0 && len(mod.Locations) <= 5 {
		fmt.Println("  Files:")
		for _, loc := range mod.Locations {
			fmt.Printf("    - %s\n", loc)
		}
	} else if len(mod.Locations) > 5 {
		fmt.Printf("  Files: %d files (...)\n", len(mod.Locations))
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
