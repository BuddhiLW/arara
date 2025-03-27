package list_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/app/list"
	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/BuddhiLW/arara/internal/pkg/vars"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Save original config dir and create test config dir
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	origTestMode := os.Getenv("TEST_MODE")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	os.Setenv("TEST_MODE", "1")

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		os.Setenv("TEST_MODE", origTestMode)
		bonzaiVars.Data.Clear()
	}

	return tmpDir, cleanup
}

func TestListCmd_ContextAware(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupFiles    func(dir string) error
		wantNamespace string
		wantErr       bool
		wantAutoAdd   bool
	}{
		{
			name: "NewDotfiles_AutoAdds",
			setupFiles: func(dir string) error {
				dotfilesDir := filepath.Join(dir, "new-dotfiles")
				if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
					return err
				}
				config := []byte(`
namespace: test-dotfiles  # Add namespace field
name: test-dotfiles
scripts:
  install:
    - name: test
      description: "Test script"
`)
				return os.WriteFile(filepath.Join(dotfilesDir, "arara.yaml"), config, 0644)
			},
			wantNamespace: "test-dotfiles",
			wantAutoAdd:   true,
		},
		{
			name: "ExistingNamespace_JustSwitches",
			setupFiles: func(dir string) error {
				// First create and add a namespace
				dotfilesDir := filepath.Join(dir, "existing")
				if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
					return err
				}
				yamlConfig := []byte(`namespace: existing
name: existing-dots`)
				if err := os.WriteFile(filepath.Join(dotfilesDir, "arara.yaml"), yamlConfig, 0644); err != nil {
					return err
				}

				// Add it to global config
				gc, err := config.NewGlobalConfig()
				if err != nil {
					return err
				}
				gc.Config.Namespaces = append(gc.Config.Namespaces, "existing")
				gc.Config.Configs["existing"] = config.NSInfo{Path: dotfilesDir}
				return gc.Save()
			},
			wantNamespace: "existing",
			wantAutoAdd:   false,
		},
		{
			name: "NoAraraYaml_NoAutoAdd",
			setupFiles: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "empty"), 0755)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := filepath.Join(tmpDir, "test-"+tt.name)
			if err := tt.setupFiles(testDir); err != nil {
				t.Fatal(err)
			}

			// Change to the actual dotfiles directory, not just the test root
			var targetDir string
			switch tt.name {
			case "NewDotfiles_AutoAdds":
				targetDir = filepath.Join(testDir, "new-dotfiles")
			case "ExistingNamespace_JustSwitches":
				targetDir = filepath.Join(testDir, "existing")
			default:
				targetDir = testDir
			}

			// Change to test directory
			origDir, _ := os.Getwd()
			if err := os.Chdir(targetDir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)

			// Clear any existing state
			bonzaiVars.Data.Clear()

			// Execute list command
			err := list.Cmd.Do(list.Cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("list.Cmd.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify namespace was set correctly
			activeNS := bonzaiVars.Fetch(vars.ActiveNamespaceEnv, vars.ActiveNamespaceVar, "")
			if activeNS != tt.wantNamespace {
				t.Errorf("active namespace = %v, want %v", activeNS, tt.wantNamespace)
			}

			// Verify namespace was added to global config if expected
			if tt.wantAutoAdd {
				gc, err := config.NewGlobalConfig()
				if err != nil {
					t.Fatal(err)
				}

				found := false
				for _, ns := range gc.Config.Namespaces {
					if ns == tt.wantNamespace {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("namespace %s was not added to global config", tt.wantNamespace)
				}
			}
		})
	}
}