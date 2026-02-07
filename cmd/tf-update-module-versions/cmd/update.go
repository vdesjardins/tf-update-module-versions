package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vdesjardins/terraform-module-versions/internal/color"
	"github.com/vdesjardins/terraform-module-versions/internal/filter"
	"github.com/vdesjardins/terraform-module-versions/internal/finder"
	"github.com/vdesjardins/terraform-module-versions/internal/registry"
	"github.com/vdesjardins/terraform-module-versions/internal/report"
	"github.com/vdesjardins/terraform-module-versions/internal/source"
	"github.com/vdesjardins/terraform-module-versions/internal/updater"
	versionpkg "github.com/vdesjardins/terraform-module-versions/internal/version"
)

var (
	modulePatterns       []string
	globalVersion        string
	updateConstraint     string
	updateConstraintFile string
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <path>",
	Short: "Update all Terraform modules to their latest versions",
	Long:  "Find all Terraform modules and update them to the latest available version from their registries",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	dirPath := args[0]

	// Validate path
	if _, err := os.Stat(dirPath); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Parse constraints
	var constraints versionpkg.Constraints
	if updateConstraint != "" && updateConstraintFile != "" {
		return fmt.Errorf("cannot use both --constraint and --constraint-file")
	}

	if updateConstraint != "" {
		var err error
		constraints, err = versionpkg.ParseConstraints(updateConstraint)
		if err != nil {
			return fmt.Errorf("invalid constraint: %w", err)
		}
	} else if updateConstraintFile != "" {
		// Read constraint file
		data, err := os.ReadFile(updateConstraintFile)
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
		output.Fprintf(os.Stderr, color.Cyan, "Applied constraints: %v\n", constraints)
	}

	// Build module filter
	moduleFilter, err := buildModuleFilter()
	if err != nil {
		return err
	}

	// Find all modules with versions
	output.Fprintf(os.Stderr, color.Blue, "Finding modules in %s...\n", dirPath)
	usages, err := finder.FindModulesWithVersions(dirPath, moduleFilter)
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(usages) == 0 {
		fmt.Println("No modules with version constraints found.")
		return nil
	}

	output.Fprintf(os.Stderr, color.Green, "Found %d module invocations\n", len(usages))

	// Analyze sources
	output.Fprintf(os.Stderr, color.Blue, "Analyzing module sources...\n")
	resolver := source.NewResolver()
	sources := make(map[string]*source.Source)
	supportedSources := []*source.Source{}

	for _, usage := range usages {
		if _, exists := sources[usage.Usage.Source]; !exists {
			src, err := resolver.Resolve(usage.Usage.Source)
			if err != nil {
				output.Fprintf(os.Stderr, color.BoldYellow, "Warning: failed to parse source %s: %v\n", usage.Usage.Source, err)
				continue
			}
			sources[usage.Usage.Source] = src

			if src.Supported {
				supportedSources = append(supportedSources, src)
			}
		}
	}

	// Fetch latest versions
	output.Fprintf(os.Stderr, color.Blue, "Fetching latest versions from registries...\n")
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

	// Print what will be updated
	printer := report.NewPrinter(summary)
	printer.Print()

	// Actually perform updates
	output.Fprintf(os.Stderr, color.Blue, "\nApplying updates...\n")
	fileUpdater := updater.NewFileUpdater()

	updatesApplied := 0
	for _, mod := range summary.Modules {
		if mod.UpdateCount == 0 {
			continue
		}

		// Determine target version based on filter strategy
		targetVersion := mod.LatestVersion
		if moduleFilter != nil {
			strategy, matched := moduleFilter.GetVersionStrategy(mod.Source)
			if !matched {
				// Module didn't match filter, skip it
				continue
			}

			// If strategy is specified, use it to select version
			if strategy != "" {
				availableVersions := latestVersions[mod.Source]
				if availableVersions != nil && len(availableVersions) > 0 {
					// Get current version (pick first from map since they're all the same version)
					var currentVersion string
					for ver := range mod.CurrentVersions {
						currentVersion = ver
						break
					}

					selectedVersion, err := versionpkg.SelectVersion(
						currentVersion,
						availableVersions,
						versionpkg.Strategy(strategy),
						constraints,
					)
					if err != nil {
						output.Fprintf(os.Stderr, color.BoldYellow, "Warning: could not select version for %s: %v\n", mod.Source, err)
						continue
					}
					targetVersion = selectedVersion
				}
			}
		}

		// Update each current version
		for currentVer := range mod.CurrentVersions {
			if currentVer == targetVersion {
				continue
			}

			updates, err := fileUpdater.UpdateDirectory(dirPath, mod.Source, currentVer, targetVersion)
			if err != nil {
				output.Fprintf(os.Stderr, color.BoldYellow, "Warning: failed to update %s: %v\n", mod.Source, err)
				continue
			}

			for file, count := range updates {
				fmt.Printf("%s✓ %s: %s %s → %s (%d changes)%s\n", output.Success("✓"), file, mod.Source, currentVer, targetVersion, count, "\033[0m")
				updatesApplied += count
			}
		}
	}

	fmt.Printf("\nFiles Updated: %d\n", updatesApplied)
	fmt.Printf("Total Changes: %d\n\n", updatesApplied)

	return nil
}

// buildModuleFilter creates ModuleFilter from parsed flags
func buildModuleFilter() (*filter.ModuleFilter, error) {
	// Check mutual exclusivity
	if len(modulePatterns) > 0 && globalVersion != "" {
		return nil, fmt.Errorf("cannot use both --module and --version flags")
	}

	mf := &filter.ModuleFilter{
		ModulePatterns: make(map[string]string),
		GlobalVersion:  globalVersion,
		WarnUnmatched:  true,
	}

	// Parse --module patterns
	for _, pattern := range modulePatterns {
		parts := strings.SplitN(pattern, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid module pattern %q: must be 'pattern=version_type'", pattern)
		}

		patternStr := strings.TrimSpace(parts[0])
		versionType := strings.TrimSpace(parts[1])

		// Validate version type
		if !versionpkg.IsValidStrategy(versionType) {
			return nil, fmt.Errorf("invalid version type %q in pattern %q: must be 'minor' or 'latest'", versionType, pattern)
		}

		// Validate regex if it looks like regex
		if _, err := filter.NewMatcher(patternStr); err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", patternStr, err)
		}

		mf.ModulePatterns[patternStr] = versionType
	}

	// Validate global version if provided
	if globalVersion != "" && !versionpkg.IsValidStrategy(globalVersion) {
		return nil, fmt.Errorf("invalid version type %q: must be 'minor' or 'latest'", globalVersion)
	}

	// If neither flag provided, return nil filter (update all to latest)
	if len(mf.ModulePatterns) == 0 && mf.GlobalVersion == "" {
		return nil, nil
	}

	return mf, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)

	flags := updateCmd.Flags()
	flags.StringSliceVar(&modulePatterns, "module", []string{},
		`Filter modules to update. Format: "pattern=version_type" where version_type is 'minor' or 'latest'.
Example: --module vault-starter=minor --module ".*vpc.*"=latest`)

	flags.StringVar(&globalVersion, "version", "",
		`Update all modules to this version type: 'minor' or 'latest'.
Mutually exclusive with --module`)

	flags.StringVar(&updateConstraint, "constraint", "",
		`Version constraints to filter available versions. Format: ">=1.0.0,<2.0.0".
Example: --constraint ">=1.2.3"`)

	flags.StringVar(&updateConstraintFile, "constraint-file", "",
		`Path to file containing version constraints (one per line).
Mutually exclusive with --constraint`)
}
