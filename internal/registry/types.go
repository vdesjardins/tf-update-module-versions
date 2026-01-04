package registry

// Module represents a module in the terraform registry
type Module struct {
	Source   string     `json:"source"`
	Versions []*Version `json:"versions"`
}

// Version represents a specific version of a module in the registry
type Version struct {
	Version            string      `json:"version"`
	Root               RootModule  `json:"root"`
	RegistryModuleInfo *ModuleInfo `json:"module_info,omitempty"`
}

// RootModule represents the root module information
type RootModule struct {
	Providers []Provider `json:"providers"`
}

// Provider represents a provider requirement
type Provider struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ModuleInfo represents registry metadata for a module version
type ModuleInfo struct {
	Source      string `json:"source"`
	PublishedAt string `json:"published_at"`
}

// registryResponse is the response structure from the registry API
type registryResponse struct {
	Modules []Module `json:"modules"`
}
