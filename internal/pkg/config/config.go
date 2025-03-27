package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env,omitempty"`

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

	Namespace string `yaml:"namespace"` // Namespace this config belongs to
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

type CompatConfig struct {
	OS     string        `yaml:"os,omitempty"`
	Arch   string        `yaml:"arch,omitempty"`
	Shell  string        `yaml:"shell,omitempty"`
	PkgMgr string        `yaml:"pkgmgr,omitempty"`
	Kernel string        `yaml:"kernel,omitempty"`
	Custom []interface{} `yaml:"custom,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	// First try loading from the specified path
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist at specified path, try user config directory
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(configDir, "arara", "config.yaml")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Load global config to validate namespace
	gc, err := NewGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Validate namespace exists
	if config.Namespace != "" {
		found := false
		for _, ns := range gc.config.Namespaces {
			if ns == config.Namespace {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("undefined namespace: %s", config.Namespace)
		}
	}

	return &config, nil
}
