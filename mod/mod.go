package mod

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// ModuleInfo holds the source and optional version of a Terraform module.
type ModuleInfo struct {
	Source  string
	Version string
}

type RegistryModule struct {
	Source   string             `json:"source"`
	Versions []*RegistryVersion `json:"versions"`
}

type RegistryRootModule struct {
	Providers []RegistryProvider `json:"providers"`
}

type RegistryProvider struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RegistryVersion represents a version entry returned by a registry.
type RegistryVersion struct {
	Version            string             `json:"version"`
	Root               RegistryRootModule `json:"root"`
	RegistryModuleInfo *RegistryModuleInfo
}

type RegistryModuleInfo struct {
	Source      string `json:"source"`
	PublishedAt string `json:"published_at"`
}

func FindModules(root string) ([]ModuleInfo, error) {
	var mods []ModuleInfo
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		module, _ := tfconfig.LoadModule(path)

		moduleCalls := module.ModuleCalls

		for _, call := range moduleCalls {
			mods = append(mods, ModuleInfo{
				Source:  call.Source,
				Version: call.Version,
			})
		}
		return nil
	})
	return mods, err
}

// UpstreamVersions fetches versions for a module source after validating format
func UpstreamVersions(source string) (*RegistryModule, error) {
	modulePaths := strings.Split(source, "//")
	parts := strings.Split(modulePaths[0], "/")

	if len(parts) != 3 && len(parts) != 4 {
		return nil, fmt.Errorf("unsupported module source: %s", source)
	}

	return UpstreamModule(source)
}

func UpstreamModule(source string) (*RegistryModule, error) {
	host := ""

	modulePaths := strings.Split(source, "//")
	parts := strings.Split(modulePaths[0], "/")
	if len(parts) == 3 {
		host = "registry.terraform.io"
	} else {
		host = parts[0]
		parts = parts[1:]
	}
	namespace, name, provider := parts[0], parts[1], parts[2]
	apiURL := fmt.Sprintf("https://%s/v1/modules/%s/%s/%s", host, namespace, name, provider)

	return fetchRegistryVersions(apiURL)
}

func fetchRegistryVersions(api string) (*RegistryModule, error) {
	apiURL := fmt.Sprintf("%s/versions", api)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry API returned %d", resp.StatusCode)
	}
	var payload struct {
		Modules []RegistryModule `json:"modules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Modules) == 0 {
		return nil, nil
	}

	err = fetchRegistryModuleInfo(api, &payload.Modules[0])

	return &payload.Modules[0], err
}

func fetchRegistryModuleInfo(api string, module *RegistryModule) error {
	for _, v := range module.Versions {
		apiURL := fmt.Sprintf("%s/%s", api, v.Version)
		resp, err := http.Get(apiURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("registry API returned %d", resp.StatusCode)
		}
		var info RegistryModuleInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return err
		}
		v.RegistryModuleInfo = &info
	}

	return nil
}
