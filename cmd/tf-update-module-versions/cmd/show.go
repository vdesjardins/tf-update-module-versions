package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vdesjardins/terraform-module-versions/internal/finder"
	"github.com/vdesjardins/terraform-module-versions/internal/registry"
	"github.com/vdesjardins/terraform-module-versions/internal/report"
	"github.com/vdesjardins/terraform-module-versions/internal/source"
	versionpkg "github.com/vdesjardins/terraform-module-versions/internal/version"
)

var (
	showConstraint     string
	showConstraintFile string
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show <path>",
	Short: "Show available module updates without making changes",
	Long:  "Analyze Terraform modules and show available updates from registries",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	dirPath := args[0]

	// Validate path
	if _, err := os.Stat(dirPath); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Parse constraints
	var constraints versionpkg.Constraints
	if showConstraint != "" && showConstraintFile != "" {
		return fmt.Errorf("cannot use both --constraint and --constraint-file")
	}

	if showConstraint != "" {
		var err error
		constraints, err = versionpkg.ParseConstraints(showConstraint)
		if err != nil {
			return fmt.Errorf("invalid constraint: %w", err)
		}
	} else if showConstraintFile != "" {
		// Read constraint file
		data, err := os.ReadFile(showConstraintFile)
		if err != nil {
			return fmt.Errorf("failed to read constraint file: %w", err)
		}
		var err2 error
		constraints, err2 = versionpkg.ParseConstraints(string(data))
		if err2 != nil {
			return fmt.Errorf("invalid constraints in file: %w", err2)
		}
	}

	// Display applied constraints if any
	if len(constraints) > 0 {
		fmt.Fprintf(os.Stderr, "Applied constraints: %v\n", constraints)
	}

	// Find all modules with versions
	fmt.Fprintf(os.Stderr, "Finding modules in %s...\n", dirPath)
	usages, err := finder.FindModulesWithVersions(dirPath, nil)
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(usages) == 0 {
		fmt.Println("No modules with version constraints found.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d module invocations\n", len(usages))

	// Analyze sources
	fmt.Fprintf(os.Stderr, "Analyzing module sources...\n")
	resolver := source.NewResolver()
	sources := make(map[string]*source.Source)
	supportedSources := []*source.Source{}

	for _, usage := range usages {
		if _, exists := sources[usage.Usage.Source]; !exists {
			src, err := resolver.Resolve(usage.Usage.Source)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse source %s: %v\n", usage.Usage.Source, err)
				continue
			}
			sources[usage.Usage.Source] = src

			if src.Supported {
				supportedSources = append(supportedSources, src)
			}
		}
	}

	// Fetch latest versions
	fmt.Fprintf(os.Stderr, "Fetching latest versions from registries...\n")
	var fetcher *registry.VersionFetcher
	if cacheStore != nil {
		// Use version fetcher with cache
		client := registry.NewClientWithCache(cacheStore)
		fetcher = registry.NewVersionFetcherWithClient(client, 4)
	} else {
		// Fall back to version fetcher without cache
		fetcher = registry.NewVersionFetcher(4)
	}
	latestVersions := fetcher.FetchMultipleVersions(context.Background(), supportedSources)

	// Build summary
	builder := report.NewBuilder()
	builder.AddModuleUsages(usages)
	builder.AddSourceInfo(sources)
	builder.AddLatestVersions(latestVersions)
	summary := builder.Build()

	// Print report
	printer := report.NewPrinter(summary)
	printer.Print(nil)

	return nil
}

func init() {
	rootCmd.AddCommand(showCmd)

	flags := showCmd.Flags()
	flags.StringVar(&showConstraint, "constraint", "",
		`Version constraints to filter available versions. Format: ">=1.0.0,<2.0.0".
Example: --constraint ">=1.2.3"`)

	flags.StringVar(&showConstraintFile, "constraint-file", "",
		`Path to file containing version constraints (one per line).
Mutually exclusive with --constraint`)
}
