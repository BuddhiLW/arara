package install_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/app/install"
	"github.com/BuddhiLW/arara/internal/pkg/config"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts", "install")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test script
	scriptPath := filepath.Join(scriptsDir, "test-script")
	script := []byte("#!/bin/sh\necho \"Test script executed\"\necho \"SCRIPTS=$SCRIPTS\"")
	if err := os.WriteFile(scriptPath, script, 0755); err != nil {
		t.Fatal(err)
	}

	// Create arara.yaml
	configData := []byte(`
name: test-config
description: Test configuration
namespace: test-ns
env:
  SCRIPTS: scripts
scripts:
  install:
    - name: test
      description: Test script
      path: scripts/install/test-script
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "arara.yaml"), configData, 0644); err != nil {
		t.Fatal(err)
	}

	// Setup global config with namespace
	gc, err := config.NewGlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	if err := gc.AddNamespace("test-ns", tmpDir, ""); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		// Cleanup handled by t.TempDir()
	}

	return tmpDir, cleanup
}

func TestInstallCmd(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set dotfiles path for testing
	t.Setenv("ARARA_DOTFILES_PATH", tmpDir)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "ListScripts",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "ExecuteValidScript",
			args:    []string{"test"},
			wantErr: false,
		},
		{
			name:    "ExecuteNonexistentScript",
			args:    []string{"nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new command instance for each test
			cmd := *install.Cmd
			cmd.Do = install.Cmd.Do // Assign the original Do function

			err := cmd.Do(&cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("install.Cmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteCmd(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set dotfiles path for testing
	t.Setenv("ARARA_DOTFILES_PATH", tmpDir)

	scriptPath := filepath.Join(tmpDir, "scripts", "install", "test-script")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "ValidScript",
			path:    scriptPath,
			wantErr: false,
		},
		{
			name:    "NonexistentScript",
			path:    "/nonexistent/script",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new command instance for each test
			cmd := *install.Cmd.Cmds[1] // executeCmd
			cmd.Do = install.Cmd.Do     // Assign the original Do function

			err := cmd.Do(&cmd, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
