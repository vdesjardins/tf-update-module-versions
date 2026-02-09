package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Config represents the TOML configuration file.
type Config struct {
	Diff  DiffConfig  `toml:"diff"`
	Cache CacheConfig `toml:"cache"`
}

type DiffConfig struct {
	Tool string `toml:"tool"`
}

type CacheConfig struct {
	Dir string `toml:"dir"`
	TTL string `toml:"ttl"`
}

func loadConfigFile() (*Config, string, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return nil, "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, path, nil
		}
		return nil, path, err
	}
	if info.IsDir() {
		return nil, path, fmt.Errorf("config path is a directory: %s", path)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, path, err
	}

	return &cfg, path, nil
}

func defaultConfigPath() (string, error) {
	configHome, err := xdgConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(configHome, "terraform-module-versions", "config.toml"), nil
}

func defaultCacheDir() (string, error) {
	cacheHome, err := xdgCacheHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheHome, "terraform-module-versions"), nil
}

func xdgConfigHome() (string, error) {
	if env := os.Getenv("XDG_CONFIG_HOME"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}

func xdgCacheHome() (string, error) {
	if env := os.Getenv("XDG_CACHE_HOME"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, ".cache"), nil
}

func applyConfigDefaults(cmd *cobra.Command, cfg *Config) error {
	cacheDirDefault, err := defaultCacheDir()
	if err != nil {
		return err
	}

	if !flagChanged(cmd, "cache-dir") {
		if cfg != nil && cfg.Cache.Dir != "" {
			cacheDir = cfg.Cache.Dir
		} else {
			cacheDir = cacheDirDefault
		}
	}

	if !flagChanged(cmd, "cache-ttl") {
		if cfg != nil && cfg.Cache.TTL != "" {
			ttl, err := time.ParseDuration(cfg.Cache.TTL)
			if err != nil {
				return fmt.Errorf("invalid cache.ttl in config: %w", err)
			}
			cacheTTL = ttl
		} else {
			cacheTTL = 24 * time.Hour
		}
	}

	if cmd != nil && cmd.Name() == "update" {
		if flag := findFlag(cmd, "diff-tool"); flag != nil && !flag.Changed {
			if cfg != nil && cfg.Diff.Tool != "" {
				diffTool = cfg.Diff.Tool
			}
		}
	}

	return nil
}

func findFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if cmd == nil {
		return nil
	}
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.InheritedFlags().Lookup(name)
}

func flagChanged(cmd *cobra.Command, name string) bool {
	flag := findFlag(cmd, name)
	if flag == nil {
		return false
	}
	return flag.Changed
}
