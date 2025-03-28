package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/arara/internal/pkg/vars"
	"github.com/rwxrob/bonzai/persisters/inyaml"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
	"gopkg.in/yaml.v3"
)

// Config represents the global arara configuration
type Config struct {
	Namespaces []string          `yaml:"namespaces"`
	Configs    map[string]NSInfo `yaml:"configs"`
}

// DotfilesConfig represents a local dotfiles configuration (arara.yaml)
type DotfilesConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env,omitempty"`
	Namespace   string            `yaml:"namespace"`

	Dependencies []string `yaml:"dependencies,omitempty"`

	Setup struct {
		BackupDirs  []string `yaml:"backup_dirs"`
		CoreLinks   []Link   `yaml:"core_links"`
		ConfigLinks []Link   `yaml:"config_links"`
	} `yaml:"setup"`

	Build struct {
		Steps []Step `yaml:"steps"`
	} `yaml:"build"`

	Scripts struct {
		Install []Script `yaml:"install,omitempty"`
	} `yaml:"scripts,omitempty"`
}

type Link struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type Step struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Command     string        `yaml:"command,omitempty"`
	Commands    []string      `yaml:"commands,omitempty"`
	Compat      *CompatConfig `yaml:"compat,omitempty"`
}

type Script struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Path        string        `yaml:"path"`
	Compat      *CompatConfig `yaml:"compat,omitempty"`
}

// String implements fmt.Stringer for interactive selection
func (s Script) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Description)
}

type CompatConfig struct {
	OS     string        `yaml:"os,omitempty"`
	Arch   string        `yaml:"arch,omitempty"`
	Shell  string        `yaml:"shell,omitempty"`
	PkgMgr string        `yaml:"pkgmgr,omitempty"`
	Kernel string        `yaml:"kernel,omitempty"`
	Custom []interface{} `yaml:"custom,omitempty"`
}

func LoadConfig(path string) (*DotfilesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config DotfilesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Only validate namespace if it's a local config and we're not in a test environment
	if filepath.Base(path) == "arara.yaml" && os.Getenv("TEST_MODE") != "1" {
		// Load global config to validate namespace
		gc, err := NewGlobalConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load global config: %w", err)
		}

		// Validate namespace exists
		if config.Namespace != "" {
			found := false
			for _, ns := range gc.Config.Namespaces {
				if ns == config.Namespace {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("undefined namespace: %s", config.Namespace)
			}
		}
	}

	return &config, nil
}

func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "arara")
}

// Marshal returns the YAML representation of the config
func (c *DotfilesConfig) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

// GlobalConfig represents the global arara configuration
type GlobalConfig struct {
	Config
	persister *inyaml.Persister
}

// Save persists the global config
func (gc *GlobalConfig) Save() error {
	data, err := yaml.Marshal(gc.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	gc.persister.Set("config", string(data))
	return nil
}

// GetDotfilesPath returns the path to the active namespace's dotfiles
func GetDotfilesPath() (string, error) {
	gc, err := NewGlobalConfig()
	if err != nil {
		return "", err
	}

	ns := gc.GetActiveNamespace()
	if ns == nil {
		return "", fmt.Errorf("no active namespace")
	}

	return ns.Path, nil
}

// GetActiveNamespace returns the active namespace's configuration
func (gc *GlobalConfig) GetActiveNamespace() *NamespaceConfig {
	activeNS := bonzaiVars.Fetch(vars.ActiveNamespaceEnv, vars.ActiveNamespaceVar, "")
	if activeNS == "" {
		return nil
	}
	if info, ok := gc.Configs[activeNS]; ok {
		return &NamespaceConfig{
			Name:     activeNS,
			Path:     info.Path,
			LocalBin: info.LocalBin,
		}
	}
	return nil
}

type NamespaceConfig struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	LocalBin string `yaml:"local-bin"`
}

type NSInfo struct {
	Path     string   `yaml:"path"`
	LocalBin string   `yaml:"local-bin"`
	Dirs     []string `yaml:"backup_dirs"`
}

var NewGlobalConfig = func() (*GlobalConfig, error) {
	persister := inyaml.NewUserConfig("arara", "config.yaml")

	gc := &GlobalConfig{
		persister: persister,
		Config: Config{
			// Initialize only what's needed for namespace management
			Namespaces: make([]string, 0),
			Configs:    make(map[string]NSInfo),
		},
	}

	// Load existing config
	if err := gc.load(); err != nil {
		return nil, err
	}

	return gc, nil
}

func (gc *GlobalConfig) load() error {
	data := gc.persister.Get("config")
	if data == "" {
		return nil
	}

	if err := yaml.Unmarshal([]byte(data), &gc.Config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// Add namespace management methods to GlobalConfig
func (gc *GlobalConfig) AddNamespace(name, path, localBin string) error {
	// Check if namespace already exists
	for _, ns := range gc.Namespaces {
		if ns == name {
			return fmt.Errorf("namespace %s already exists", name)
		}
	}

	// Add namespace
	gc.Namespaces = append(gc.Namespaces, name)
	gc.Configs[name] = NSInfo{
		Path:     path,
		LocalBin: localBin,
	}

	return gc.Save()
}
