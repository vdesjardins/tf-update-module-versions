package finder

// ModuleUsage represents a module usage in a Terraform configuration
type ModuleUsage struct {
	Source    string // e.g., "hashicorp/vault-starter/aws"
	Version   string // e.g., "0.1.3"
	FilePath  string // Absolute path to the .tf file
	BlockName string // Module block name, e.g., "example" from module "example"
}

// ModuleWithPath is a convenience type combining a file path with module usage
type ModuleWithPath struct {
	FilePath string
	Usage    ModuleUsage
}
