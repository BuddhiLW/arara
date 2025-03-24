package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Valid config content
	validConfig := `
name: "test-dotfiles"
description: "Test dotfiles configuration"

setup:
  backup_dirs:
    - "$HOME/.config"
    - "$HOME/.local"
  
  core_links:
    - source: "$DOTFILES/.config"
      target: "$HOME/.config"
    - source: "$DOTFILES/.local"
      target: "$HOME/.local"

  config_links:
    - source: "$DOTFILES/.bashrc"
      target: "$HOME/.bashrc"
    - source: "$DOTFILES/.vimrc"
      target: "$HOME/.vimrc"

build:
  steps:
    - name: "backup"
      description: "Backup existing dotfiles"
      command: "arara setup backup"

    - name: "link"
      description: "Create symlinks"
      command: "arara setup link"

    - name: "custom"
      description: "Custom commands"
      commands:
        - "echo 'Step 1'"
        - "echo 'Step 2'"
`

	// Write valid config to temp file
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Test valid config loading
	t.Run("ValidConfig", func(t *testing.T) {
		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load valid config: %v", err)
		}

		// Verify loaded config values
		if cfg.Name != "test-dotfiles" {
			t.Errorf("Expected name 'test-dotfiles', got '%s'", cfg.Name)
		}

		if cfg.Description != "Test dotfiles configuration" {
			t.Errorf("Expected description 'Test dotfiles configuration', got '%s'", cfg.Description)
		}

		// Check backup dirs
		expected := []string{"$HOME/.config", "$HOME/.local"}
		if !reflect.DeepEqual(cfg.Setup.BackupDirs, expected) {
			t.Errorf("Expected backup dirs %v, got %v", expected, cfg.Setup.BackupDirs)
		}

		// Check links
		if len(cfg.Setup.CoreLinks) != 2 {
			t.Errorf("Expected 2 core links, got %d", len(cfg.Setup.CoreLinks))
		}

		if len(cfg.Setup.ConfigLinks) != 2 {
			t.Errorf("Expected 2 config links, got %d", len(cfg.Setup.ConfigLinks))
		}

		// Check build steps
		if len(cfg.Build.Steps) != 3 {
			t.Errorf("Expected 3 build steps, got %d", len(cfg.Build.Steps))
		}

		// Check command vs commands field
		if cfg.Build.Steps[0].Command != "arara setup backup" {
			t.Errorf("Expected command 'arara setup backup', got '%s'", cfg.Build.Steps[0].Command)
		}

		expectedCommands := []string{"echo 'Step 1'", "echo 'Step 2'"}
		if !reflect.DeepEqual(cfg.Build.Steps[2].Commands, expectedCommands) {
			t.Errorf("Expected commands %v, got %v", expectedCommands, cfg.Build.Steps[2].Commands)
		}
	})

	// Test invalid file path
	t.Run("InvalidPath", func(t *testing.T) {
		_, err := LoadConfig("/path/that/does/not/exist.yaml")
		if err == nil {
			t.Error("Expected error for non-existent config file, got nil")
		}
	})

	// Test invalid YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644); err != nil {
			t.Fatalf("Failed to create invalid yaml file: %v", err)
		}

		_, err := LoadConfig(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}