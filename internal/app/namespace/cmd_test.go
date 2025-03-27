package namespace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/app/namespace"
	"github.com/BuddhiLW/arara/internal/pkg/config"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Save original config dir
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")

	// Set config dir to temp
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Copy testdata to temp dir
	testdataDir := "testdata"
	dotfilesDir := filepath.Join(tmpDir, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy test arara.yaml from testdata
	testConfig, err := os.ReadFile(filepath.Join(testdataDir, "arara.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dotfilesDir, "arara.yaml"), testConfig, 0644); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		bonzaiVars.Data.Clear() // Clear bonzai vars
	}

	return tmpDir, cleanup
}

func TestAddNamespace(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name    string
		nsName  string
		path    string
		wantErr bool
	}{
		{
			name:    "ValidNamespace",
			nsName:  "test",
			path:    filepath.Join(tmpDir, "dotfiles"),
			wantErr: false,
		},
		{
			name:    "DuplicateNamespace",
			nsName:  "test",
			path:    filepath.Join(tmpDir, "dotfiles"),
			wantErr: true,
		},
		{
			name:    "InvalidPath",
			nsName:  "invalid",
			path:    "/nonexistent/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := namespace.Cmd.Cmds[4].Do(nil, tt.nsName, tt.path) // addCmd
			if (err != nil) != tt.wantErr {
				t.Errorf("addCmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify namespace was added
				gc, err := config.NewGlobalConfig()
				if err != nil {
					t.Fatal(err)
				}

				found := false
				for _, ns := range gc.Config.Namespaces {
					if ns == tt.nsName {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("namespace %s not found in config", tt.nsName)
				}

				info, ok := gc.Config.Configs[tt.nsName]
				if !ok {
					t.Errorf("namespace %s config not found", tt.nsName)
				}
				if info.Path != tt.path {
					t.Errorf("namespace %s path = %s, want %s", tt.nsName, info.Path, tt.path)
				}
			}
		})
	}
}

func TestRemoveNamespace(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add test namespace first
	testPath := filepath.Join(tmpDir, "dotfiles")
	if err := namespace.Cmd.Cmds[4].Do(nil, "test", testPath); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		nsName  string
		wantErr bool
	}{
		{
			name:    "ValidNamespace",
			nsName:  "test",
			wantErr: false,
		},
		{
			name:    "NonexistentNamespace",
			nsName:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := namespace.Cmd.Cmds[5].Do(nil, tt.nsName) // removeCmd
			if (err != nil) != tt.wantErr {
				t.Errorf("removeCmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify namespace was removed
				gc, err := config.NewGlobalConfig()
				if err != nil {
					t.Fatal(err)
				}

				for _, ns := range gc.Config.Namespaces {
					if ns == tt.nsName {
						t.Errorf("namespace %s still exists in config", tt.nsName)
					}
				}

				if _, ok := gc.Config.Configs[tt.nsName]; ok {
					t.Errorf("namespace %s config still exists", tt.nsName)
				}
			}
		})
	}
}
