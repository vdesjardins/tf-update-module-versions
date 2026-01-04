package report

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Printer handles output to console
type Printer struct {
	summary *UpdateSummary
}

// NewPrinter creates a new report printer
func NewPrinter(summary *UpdateSummary) *Printer {
	return &Printer{summary: summary}
}

// Print outputs the summary to console
func (p *Printer) Print() {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  Terraform Module Update Summary")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Overview
	fmt.Println("\nModule Status Overview")
	fmt.Println("─────────────────────")
	fmt.Printf("  Supported Modules:        %d\n", p.summary.SuportedCount)
	fmt.Printf("  Unsupported Modules:      %d\n", p.summary.UnsupportedCount)
	fmt.Printf("  Total Module Usages:      %d\n", p.summary.TotalUsages)

	// Detailed reports for supported modules
	if len(p.summary.Modules) > 0 {
		fmt.Println("Supported Modules")
		fmt.Println("─────────────────")
		for _, mod := range p.summary.Modules {
			p.printModuleReport(&mod)
		}
	}

	// Unsupported modules
	if len(p.summary.UnsupportedModules) > 0 {
		fmt.Println("Unsupported Modules")
		fmt.Println("───────────────────")
		for _, unsup := range p.summary.UnsupportedModules {
			fmt.Printf("\n✗ %s (%s)\n", unsup.Source, unsup.Type.String())
			fmt.Printf("  Total Usages: %d\n", unsup.Count)
			fmt.Println("  Status:       NOT SUPPORTED (future enhancement)")
		}
		fmt.Println()
	}

	// Summary stats
	fmt.Println("Summary")
	fmt.Println("───────")
	fmt.Printf("  Total Module Invocations:           %d\n", p.summary.TotalUsages)
	fmt.Printf("  Module Invocations to Update:       %d\n", p.summary.TotalUpdated)
	fmt.Printf("  Module Invocations Already Latest:  %d\n", p.summary.TotalUsages-p.summary.TotalUpdated)

	// Version change details
	if len(p.summary.ByVersionChange) > 0 {
		fmt.Println("\nVersion Changes Summary")
		fmt.Println("──────────────────────")

		// Sort keys for consistent output
		var keys []string
		for k := range p.summary.ByVersionChange {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			count := p.summary.ByVersionChange[key]
			fmt.Printf("  %s (%d changes)\n", key, count)
		}
	}

	fmt.Println()
}

func (p *Printer) printModuleReport(mod *ModuleReport) {
	fmt.Printf("\n✓ %s (%s)\n", mod.Source, mod.Type.String())

	// Current versions
	fmt.Print("  Current Versions:  ")
	var versionLines []string
	for v, count := range mod.CurrentVersions {
		versionLines = append(versionLines, fmt.Sprintf("%s (%d)", v, count))
	}
	sort.Strings(versionLines)
	fmt.Println(strings.Join(versionLines, ", "))

	fmt.Printf("  Latest Version:    %s\n", mod.LatestVersion)
	fmt.Printf("  Modules to Update: %d\n", mod.UpdateCount)

	if mod.UpdateCount > 0 {
		fmt.Printf("  Status:            UPDATE AVAILABLE\n")
	} else {
		fmt.Printf("  Status:            ALREADY AT LATEST\n")
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
	fmt.Fprintf(os.Stderr, "\n❌ Error: %s\n\n", message)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("\n✅ %s\n\n", message)
}
