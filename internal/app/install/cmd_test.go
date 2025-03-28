package install_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/app/install"
)

func setupTestEnv(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Create scripts directory structure
	scriptsDir := filepath.Join(tmpDir, "scripts", "install")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy test script from testdata
	testScript := filepath.Join("testdata", "scripts", "install", "test-script")
	destScript := filepath.Join(scriptsDir, "test-script")

	scriptData, err := os.ReadFile(testScript)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(destScript, scriptData, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test arara.yaml
	if err := os.WriteFile(filepath.Join(tmpDir, "arara.yaml"), []byte(`
namespace: test
name: test-dotfiles
scripts:
  install:
    - name: test
      description: "Test script"
      path: "scripts/install/test-script"
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up test namespace in global config
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Set XDG_CONFIG_HOME to our test config directory
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", configDir)
	
	// Set up active namespace and test mode
	os.Setenv("ARARA_ACTIVE_NAMESPACE", "test")
	os.Setenv("ARARA_DOTFILES_PATH", tmpDir)
	os.Setenv("TEST_MODE", "1")

	cleanup := func() {
		os.RemoveAll(tmpDir)
		os.Setenv("XDG_CONFIG_HOME", origConfigHome)
		os.Unsetenv("ARARA_ACTIVE_NAMESPACE")
		os.Unsetenv("ARARA_DOTFILES_PATH")
		os.Unsetenv("TEST_MODE")
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

	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:    "ValidScript",
			script:  "test",
			wantErr: false,
		},
		{
			name:    "NonexistentScript",
			script:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := install.Cmd.Do(install.Cmd, tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("install.Cmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
