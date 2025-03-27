package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/pkg/config"
)

func setupTestConfig(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test config file
	configData := []byte(`
name: test-config
description: Test configuration
namespace: test-ns

setup:
  backup_dirs:
    - $HOME/.config
  core_links:
    - source: $DOTFILES/.config
      target: $HOME/.config
  config_links:
    - source: $DOTFILES/.bashrc
      target: $HOME/.bashrc

build:
  steps:
    - name: test-step
      description: Test step
      command: echo "test"
`)

	configPath := filepath.Join(tmpDir, "arara.yaml")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create global config for namespace validation
	gc, err := config.NewGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	if err := gc.AddNamespace("test-ns", tmpDir, ""); err != nil {
		t.Fatalf("Failed to add test namespace: %v", err)
	}

	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}

	return configPath, cleanup
}

func TestLoadConfig(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "ValidConfig",
			path:    configPath,
			wantErr: false,
		},
		{
			name:    "InvalidPath",
			path:    "/nonexistent/path/config.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadConfig(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg == nil {
				t.Error("LoadConfig() returned nil config without error")
			}

			if !tt.wantErr {
				if cfg.Name != "test-config" {
					t.Errorf("LoadConfig() got name = %v, want %v", cfg.Name, "test-config")
				}
				if cfg.Namespace != "test-ns" {
					t.Errorf("LoadConfig() got namespace = %v, want %v", cfg.Namespace, "test-ns")
				}
			}
		})
	}
}
