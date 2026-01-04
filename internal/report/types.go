package report

import (
	"github.com/vdesjardins/terraform-module-versions/internal/source"
)

// ModuleReport represents a summary report for one module source
type ModuleReport struct {
	Source          string                // Module source
	Type            source.SourceTypeEnum // Registry type
	Supported       bool                  // Whether we can fetch versions
	CurrentVersions map[string]int        // Version -> count of usages
	LatestVersion   string                // Latest available version
	TotalUsages     int                   // Total module invocations
	UpdateCount     int                   // Count that will be updated
	UpcomingVersion string                // What version will be updated to
	Locations       []string              // File paths with this module
}

// UnsupportedSource represents a module source we can't update
type UnsupportedSource struct {
	Source string
	Type   source.SourceTypeEnum
	Count  int // Number of usages
}

// UpdateSummary is the final report of all findings
type UpdateSummary struct {
	Modules            []ModuleReport
	UnsupportedModules []UnsupportedSource
	TotalUsages        int            // Total across all modules
	TotalUpdated       int            // Total that would be changed
	ByVersionChange    map[string]int // "1.0.0 → 2.0.0": count
	SuportedCount      int            // Count of supported modules
	UnsupportedCount   int            // Count of unsupported modules
}
