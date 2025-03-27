package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v "github.com/BuddhiLW/arara/internal/pkg/vars"
	"github.com/rwxrob/bonzai/persisters/inyaml"
	"github.com/rwxrob/bonzai/vars"
	"gopkg.in/yaml.v3"
)

// NamespaceConfig represents the global Arara configuration for managing multiple dotfiles
type NamespaceConfig struct {
	Namespaces []string          `yaml:"namespaces"`
	Configs    map[string]NSInfo `yaml:"configs"`
}

// NSInfo holds configuration for a specific namespace
type NSInfo struct {
	Path     string `yaml:"path"`
	LocalBin string `yaml:"local-bin,omitempty"` // Optional, defaults to namespace name
}

// GlobalConfig manages the persistent namespace configuration
type GlobalConfig struct {
	persister *inyaml.Persister
	Config    struct {
		Namespaces []string          `yaml:"namespaces"`
		Configs    map[string]NSInfo `yaml:"configs"`
	} `yaml:"config"`
}

// NewGlobalConfig creates a new global configuration manager
func NewGlobalConfig() (*GlobalConfig, error) {
	persister := inyaml.NewUserConfig("arara", "config.yaml")

	gc := &GlobalConfig{
		persister: persister,
		Config: struct {
			Namespaces []string          `yaml:"namespaces"`
			Configs    map[string]NSInfo `yaml:"configs"`
		}{
			Configs: make(map[string]NSInfo),
		},
	}

	// Load existing config if it exists
	if err := gc.load(); err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	return gc, nil
}

// load reads the configuration from disk
func (gc *GlobalConfig) load() error {
	data := gc.persister.Get("config")
	if data == "" {
		return nil // No existing config
	}

	if err := yaml.Unmarshal([]byte(data), &gc.Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Save persists the configuration to disk
func (gc *GlobalConfig) Save() error {
	data, err := yaml.Marshal(gc.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	gc.persister.Set("config", string(data))
	return nil
}

// AddNamespace registers a new namespace
func (gc *GlobalConfig) AddNamespace(name, path string, localBin string) error {
	if localBin == "" {
		localBin = name // Default to namespace name
	}

	// Validate path exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("invalid path for namespace %s: %w", name, err)
	}

	// Add to namespaces list if not present
	found := false
	for _, ns := range gc.Config.Namespaces {
		if ns == name {
			found = true
			break
		}
	}
	if !found {
		gc.Config.Namespaces = append(gc.Config.Namespaces, name)
	}

	// Update namespace config
	gc.Config.Configs[name] = NSInfo{
		Path:     path,
		LocalBin: localBin,
	}

	return gc.Save()
}

// UpdateShellRC updates shell initialization files with PATH additions
func (gc *GlobalConfig) UpdateShellRC() error {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Build PATH additions
	var paths []string
	for _, info := range gc.Config.Configs {
		binPath := filepath.Join(info.Path, ".local/bin", info.LocalBin)
		if _, err := os.Stat(binPath); err == nil {
			paths = append(paths, binPath)
		}
	}

	// Update .bashrc
	rcFile := filepath.Join(home, ".bashrc")
	if err := gc.updateRCFile(rcFile, paths); err != nil {
		return fmt.Errorf("failed to update .bashrc: %w", err)
	}

	return nil
}

// updateRCFile updates a shell RC file with PATH additions
func (gc *GlobalConfig) updateRCFile(path string, paths []string) error {
	// Read existing content
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Build PATH addition line
	pathLine := "\n<<<< Added by Arara - DO NOT EDIT THIS SECTION\n"
	for _, p := range paths {
		pathLine += fmt.Sprintf("export PATH=\"%s:$PATH\"\n", p)
	}
	pathLine += ">>>> End Arara section\n"

	// Update or add Arara section
	lines := strings.Split(string(content), "\n")
	var newLines []string
	foundSection := false
	inAraraSection := false

	for _, line := range lines {
		if strings.Contains(line, "<<<< Added by Arara") {
			foundSection = true
			inAraraSection = true
			newLines = append(newLines, pathLine)
			continue
		}
		if strings.Contains(line, ">>>> End Arara section") {
			inAraraSection = false
			continue
		}
		if !inAraraSection {
			newLines = append(newLines, line)
		}
	}

	// If no existing section was found, add it
	if !foundSection {
		newLines = append(newLines, pathLine)
	}

	// Write updated content
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(path, []byte(newContent), 0644)
}

// GetDotfilesPath returns the path to the active dotfiles repository
func GetDotfilesPath() string {
	// Try environment variable first
	if path := os.Getenv(v.DotfilesPathEnv); path != "" {
		return path
	}

	// Then try persistent variable
	if path, _ := vars.Data.Get(v.DotfilesPathVar); path != "" {
		return path
	}

	// Finally try current directory
	if pwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(pwd, "arara.yaml")); err == nil {
			return pwd
		}
	}

	return ""
}

// GetActiveNamespace returns the currently active namespace
func GetActiveNamespace() string {
	// Try environment variable first
	if ns := os.Getenv(v.ActiveNamespaceEnv); ns != "" {
		return ns
	}

	// Then try persistent variable
	if ns, _ := vars.Data.Get(v.ActiveNamespaceVar); ns != "" {
		return ns
	}

	return ""
}
