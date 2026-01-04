package report

import (
	"fmt"

	"github.com/vdesjardins/terraform-module-versions/internal/finder"
	"github.com/vdesjardins/terraform-module-versions/internal/source"
)

// Builder constructs UpdateSummary from findings and registry results
type Builder struct {
	modules map[string]*ModuleReport
}

// NewBuilder creates a new summary builder
func NewBuilder() *Builder {
	return &Builder{
		modules: make(map[string]*ModuleReport),
	}
}

// AddModuleUsages processes module usages and groups by source
func (b *Builder) AddModuleUsages(usages []finder.ModuleWithPath) {
	for _, usage := range usages {
		if _, exists := b.modules[usage.Usage.Source]; !exists {
			b.modules[usage.Usage.Source] = &ModuleReport{
				Source:          usage.Usage.Source,
				CurrentVersions: make(map[string]int),
				Locations:       []string{},
			}
		}

		mod := b.modules[usage.Usage.Source]
		mod.CurrentVersions[usage.Usage.Version]++
		mod.TotalUsages++

		// Track location if not too many
		if len(mod.Locations) < 100 {
			mod.Locations = append(mod.Locations, usage.FilePath)
		}
	}
}

// AddSourceInfo adds source type information to modules
func (b *Builder) AddSourceInfo(sources map[string]*source.Source) {
	for sourceStr, src := range sources {
		if mod, exists := b.modules[sourceStr]; exists {
			mod.Type = src.Type
			mod.Supported = src.Supported
		}
	}
}

// AddLatestVersions adds latest version info from registry
// Accepts map from fetcher which returns latest first
func (b *Builder) AddLatestVersions(versionsMap map[string][]string) {
	for sourceStr, versions := range versionsMap {
		if mod, exists := b.modules[sourceStr]; exists {
			if len(versions) > 0 {
				// Versions are already sorted latest-first
				mod.LatestVersion = versions[0]

				// Calculate how many would be updated
				for ver, count := range mod.CurrentVersions {
					if ver != mod.LatestVersion {
						mod.UpdateCount += count
					}
				}

				mod.UpcomingVersion = mod.LatestVersion
			}
		}
	}
}

// Build constructs the final UpdateSummary
func (b *Builder) Build() *UpdateSummary {
	summary := &UpdateSummary{
		ByVersionChange: make(map[string]int),
	}

	var supported []ModuleReport
	var unsupported []UnsupportedSource

	totalUsages := 0
	totalUpdated := 0

	for _, mod := range b.modules {
		totalUsages += mod.TotalUsages

		if mod.Supported {
			supported = append(supported, *mod)
			totalUpdated += mod.UpdateCount

			// Build version change map
			for ver, count := range mod.CurrentVersions {
				if ver != mod.LatestVersion {
					changeKey := fmt.Sprintf("%s → %s", ver, mod.LatestVersion)
					summary.ByVersionChange[changeKey] += count
				}
			}
		} else {
			unsupported = append(unsupported, UnsupportedSource{
				Source: mod.Source,
				Type:   mod.Type,
				Count:  mod.TotalUsages,
			})
		}
	}

	summary.Modules = supported
	summary.UnsupportedModules = unsupported
	summary.TotalUsages = totalUsages
	summary.TotalUpdated = totalUpdated
	summary.SuportedCount = len(supported)
	summary.UnsupportedCount = len(unsupported)

	return summary
}

// BuildQuick builds summary without location tracking (faster for large projects)
func BuildQuick(usages []finder.ModuleWithPath, sources map[string]*source.Source, latestVersions map[string][]string) *UpdateSummary {
	builder := NewBuilder()

	// Add usages without location tracking
	moduleMap := make(map[string]*ModuleReport)
	for _, usage := range usages {
		if _, exists := moduleMap[usage.Usage.Source]; !exists {
			moduleMap[usage.Usage.Source] = &ModuleReport{
				Source:          usage.Usage.Source,
				CurrentVersions: make(map[string]int),
				Locations:       []string{},
			}
		}
		moduleMap[usage.Usage.Source].CurrentVersions[usage.Usage.Version]++
		moduleMap[usage.Usage.Source].TotalUsages++
	}

	builder.modules = moduleMap

	// Add source info
	builder.AddSourceInfo(sources)

	// Add latest versions
	builder.AddLatestVersions(latestVersions)

	return builder.Build()
}
