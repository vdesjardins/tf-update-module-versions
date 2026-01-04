package finder

import (
	"io/fs"
	"path/filepath"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/vdesjardins/terraform-module-versions/internal/filter"
)

// FindModulesWithVersions recursively finds all Terraform modules with explicit version constraints
// Only returns modules that have a version attribute specified
// If filter is provided, only returns modules matching the filter criteria
func FindModulesWithVersions(root string, moduleFilter *filter.ModuleFilter) ([]ModuleWithPath, error) {
	var results []ModuleWithPath

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process directories
		if !d.IsDir() {
			return nil
		}

		// Load the terraform module configuration for this directory
		module, _ := tfconfig.LoadModule(path)

		// Extract module calls with version constraints
		for _, call := range module.ModuleCalls {
			// Only include modules with explicit version specified
			if call.Version == "" {
				continue
			}

			// Apply filter if provided
			if moduleFilter != nil {
				_, matches := moduleFilter.GetVersionStrategy(call.Source)
				if !matches {
					continue
				}
			}

			results = append(results, ModuleWithPath{
				FilePath: path,
				Usage: ModuleUsage{
					Source:    call.Source,
					Version:   call.Version,
					FilePath:  path,
					BlockName: call.Name,
				},
			})
		}

		return nil
	})

	return results, err
}

// FindAllModules recursively finds all Terraform modules (regardless of version constraint)
func FindAllModules(root string) ([]ModuleWithPath, error) {
	var results []ModuleWithPath

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		module, _ := tfconfig.LoadModule(path)

		for _, call := range module.ModuleCalls {
			results = append(results, ModuleWithPath{
				FilePath: path,
				Usage: ModuleUsage{
					Source:    call.Source,
					Version:   call.Version,
					FilePath:  path,
					BlockName: call.Name,
				},
			})
		}

		return nil
	})

	return results, err
}
