package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/vdesjardins/terraform-module-versions/internal/cache"
	"github.com/vdesjardins/terraform-module-versions/internal/color"
)

var (
	version    = "dev"
	commit     = "unknown"
	buildTime  = "unknown"
	cacheDir   = ""
	cacheTTL   = 24 * time.Hour
	cacheClear = false
	cacheStore cache.Store
	output     *color.ColoredOutput
	pager      *color.Pager
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terraform-module-versions",
	Short: "Manage Terraform module versions",
	Long:  "A tool to manage Terraform module versions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize colored output
		output = color.New()

		// Initialize cache if cache operations are needed
		if cacheDir == "" {
			// Default to ~/.cache/terraform-module-versions
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to determine home directory: %w", err)
			}
			cacheDir = filepath.Join(home, ".cache", "terraform-module-versions")
		}

		// Create cache directory if it doesn't exist
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}

		// Clear cache if requested
		if cacheClear {
			diskStore, err := cache.NewDiskStore(cacheDir)
			if err != nil {
				return fmt.Errorf("failed to create cache store for clearing: %w", err)
			}
			if err := diskStore.Clear(); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}
			diskStore.Close()
			output.Fprintf(os.Stderr, color.Green, "Cache cleared\n")
		}

		// Initialize the global cache store
		store, err := cache.NewDiskStore(cacheDir)
		if err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
		cacheStore = store

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if pager != nil {
			pager.Close()
		}
		// Clean up cache store resources
		if cacheStore != nil {
			cacheStore.Close()
		}
		return nil
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Initialize output if not already done (for early errors)
		if output == nil {
			output = color.New()
		}
		output.Fprintf(os.Stderr, color.BoldRed, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "Directory for cache storage (default: $HOME/.cache/terraform-module-versions)")
	rootCmd.PersistentFlags().DurationVar(&cacheTTL, "cache-ttl", 24*time.Hour, "Cache entry TTL (e.g., 24h, 1h30m)")
	rootCmd.PersistentFlags().BoolVar(&cacheClear, "cache-clear", false, "Clear the cache before running")

	cobra.AddTemplateFunc("heading", func(text string) string {
		return color.New().Sprintf(color.BoldCyan, "%s", text)
	})

	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{.}}

{{end}}{{heading "Usage"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{heading "Aliases"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{heading "Examples"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{heading "Available Commands"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{heading "Flags"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{heading "Global Flags"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`)

	rootCmd.SetUsageTemplate(`{{heading "Usage"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableLocalFlags}}

{{heading "Flags"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{heading "Global Flags"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`)
}

// SetVersion allows setting the version at runtime
func SetVersion(v, c, b string) {
	version = v
	commit = c
	buildTime = b
	rootCmd.Version = fmt.Sprintf("v%s (commit: %s, built: %s)", version, commit, buildTime)
}
