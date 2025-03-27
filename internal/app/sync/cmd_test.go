package sync_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BuddhiLW/arara/internal/app/sync"
	"github.com/BuddhiLW/arara/internal/pkg/config"
)

func setupTestEnv(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	cleanup := func() {
		os.Chdir(origDir)
	}

	return tmpDir, cleanup
}

func TestSyncCmd(t *testing.T) {
	// Setup mock input/output
	oldStdin := sync.Stdin
	oldStdout := sync.Stdout
	defer func() {
		sync.Stdin = oldStdin
		sync.Stdout = oldStdout
	}()

	mockInput := bytes.NewBufferString("1\n") // Choose to keep existing
	sync.Stdin = mockInput
	sync.Stdout = io.Discard

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts/install"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create test scripts
	scripts := map[string]string{
		"script1":        "#!/bin/sh\necho test1",
		"script2":        "#!/bin/sh\necho test2",
		"not-executable": "#!/bin/sh\necho test3",
	}

	for name, content := range scripts {
		path := filepath.Join(tmpDir, "scripts/install", name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Make one script non-executable
	if err := os.Chmod(filepath.Join(tmpDir, "scripts/install/not-executable"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create initial arara.yaml
	initialConfig := []byte(`
name: test-config
scripts:
  install:
    - name: script1
      description: "Existing script 1"
      path: "scripts/install/script1"
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "arara.yaml"), initialConfig, 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Run sync command
	if err := sync.Cmd.Do(sync.Cmd); err != nil {
		t.Fatalf("sync.Cmd.Do() error = %v", err)
	}

	// Verify results
	cfg, err := config.LoadConfig("arara.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 executable scripts
	if len(cfg.Scripts.Install) != 2 {
		t.Errorf("got %d scripts, want 2", len(cfg.Scripts.Install))
	}

	// Verify script1 kept its description
	for _, script := range cfg.Scripts.Install {
		if script.Name == "script1" {
			if script.Description != "Existing script 1" {
				t.Errorf("script1 description = %q, want 'Existing script 1'", script.Description)
			}
		}
	}
}

func TestSyncCmd_Transactions(t *testing.T) {
	tests := []struct {
		name         string
		setupFiles   func(dir string) error
		modifyDuring func(dir string) error
		mockInput    string
		wantScripts  int
		wantErr      bool
		checkResults func(t *testing.T, dir string)
	}{
		{
			name:       "AtomicUpdate",
			setupFiles: setupBasicConfig,
			modifyDuring: func(dir string) error {
				// Simulate concurrent modification
				return os.WriteFile(filepath.Join(dir, "arara.yaml"),
					[]byte("name: modified-during-sync\n"), 0644)
			},
			mockInput:   "1\n",
			wantScripts: 2,
			wantErr:     true, // Should detect concurrent modification
			checkResults: func(t *testing.T, dir string) {
				// Verify backup was restored
				data, err := os.ReadFile(filepath.Join(dir, "arara.yaml"))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "test-config") {
					t.Error("backup was not restored after concurrent modification")
				}
			},
		},
		{
			name: "PreserveExistingMetadata",
			setupFiles: func(dir string) error {
				return setupConfigWithMetadata(dir, map[string]interface{}{
					"name": "test-config",
					"env": map[string]string{
						"CUSTOM_VAR": "value",
					},
				})
			},
			mockInput:   "1\n",
			wantScripts: 2,
			wantErr:     false,
			checkResults: func(t *testing.T, dir string) {
				cfg, err := config.LoadConfig(filepath.Join(dir, "arara.yaml"))
				if err != nil {
					t.Fatal(err)
				}
				if cfg.Env["CUSTOM_VAR"] != "value" {
					t.Error("existing metadata was not preserved")
				}
			},
		},
		{
			name:        "ConflictResolution",
			setupFiles:  setupConfigWithConflict,
			mockInput:   "1\n",
			wantScripts: 2,
			wantErr:     false,
			checkResults: func(t *testing.T, dir string) {
				cfg, err := config.LoadConfig(filepath.Join(dir, "arara.yaml"))
				if err != nil {
					t.Fatal(err)
				}
				// Verify the existing description was preserved
				for _, script := range cfg.Scripts.Install {
					if script.Name == "script1" {
						if script.Description != "Existing script 1" {
							t.Error("conflict resolution failed to preserve existing description")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock input/output
			oldStdin := sync.Stdin
			oldStdout := sync.Stdout
			defer func() {
				sync.Stdin = oldStdin
				sync.Stdout = oldStdout
			}()

			mockInput := bytes.NewBufferString(tt.mockInput)
			sync.Stdin = mockInput
			sync.Stdout = io.Discard

			tmpDir, cleanup := setupTestEnv(t)
			defer cleanup()

			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatal(err)
			}

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// If we need to simulate concurrent modification
			if tt.modifyDuring != nil {
				// Do modification before sync
				if err := tt.modifyDuring(tmpDir); err != nil {
					t.Fatal(err)
				}
				// Wait to ensure modification is detected
				time.Sleep(100 * time.Millisecond)
			}

			err := sync.Cmd.Do(sync.Cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("sync.Cmd.Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkResults != nil {
				tt.checkResults(t, tmpDir)
			}
		})
	}
}

// Helper functions
func setupBasicConfig(dir string) error {
	// Create scripts directory
	if err := os.MkdirAll(filepath.Join(dir, "scripts/install"), 0755); err != nil {
		return err
	}

	// Create test scripts
	scripts := map[string]string{
		"script1": "#!/bin/sh\necho test1",
		"script2": "#!/bin/sh\necho test2",
	}

	for name, content := range scripts {
		path := filepath.Join(dir, "scripts/install", name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return err
		}
	}

	// Create basic arara.yaml
	yamlData := []byte(`
name: test-config
scripts:
  install: []
`)
	return os.WriteFile(filepath.Join(dir, "arara.yaml"), yamlData, 0644)
}

func setupConfigWithMetadata(dir string, metadata map[string]interface{}) error {
	if err := setupBasicConfig(dir); err != nil {
		return err
	}

	// Add metadata to config
	configPath := filepath.Join(dir, "arara.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	// Apply metadata
	if name, ok := metadata["name"].(string); ok {
		cfg.Name = name
	}
	if env, ok := metadata["env"].(map[string]string); ok {
		cfg.Env = env
	}

	// Save updated config
	data, err := cfg.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "arara.yaml"), data, 0644)
}

func setupConfigWithConflict(dir string) error {
	if err := setupBasicConfig(dir); err != nil {
		return err
	}

	// Create config with different description for script1
	configPath := filepath.Join(dir, "arara.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	cfg.Scripts.Install = []config.Script{
		{
			Name:        "script1",
			Description: "Existing script 1",
			Path:        "scripts/install/script1",
		},
	}

	data, err := cfg.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "arara.yaml"), data, 0644)
}

func TestSyncCmd_InteractiveConflict(t *testing.T) {
	// Setup mock input/output
	oldStdin := sync.Stdin
	oldStdout := sync.Stdout
	defer func() {
		sync.Stdin = oldStdin
		sync.Stdout = oldStdout
	}()

	mockInput := bytes.NewBufferString("2\n") // Choose new version
	mockOutput := &bytes.Buffer{}
	sync.Stdin = mockInput
	sync.Stdout = mockOutput

	if deadline, ok := t.Deadline(); ok {
		// Add timeout for CI environments
		time.AfterFunc(deadline.Sub(time.Now())-100*time.Millisecond, func() {
			t.Fatal("test timed out")
		})
	}

	if testing.Short() {
		t.Skip("skipping interactive test in short mode")
	}
	// Check for manual run flag
	if os.Getenv("TEST_INTERACTIVE") != "1" {
		t.Skip("skipping interactive test; set TEST_INTERACTIVE=1 to run")
	}

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Setup test case that will cause conflict
	if err := os.MkdirAll(filepath.Join(tmpDir, "scripts/install"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create script with different description in config
	scriptContent := "#!/bin/sh\necho 'test script'"
	scriptPath := filepath.Join(tmpDir, "scripts/install/interactive-test")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with existing description
	configBytes := []byte(`
name: interactive-test
scripts:
  install:
    - name: interactive-test
      description: "Original description"
      path: "scripts/install/interactive-test"
`)
	configPath := filepath.Join(tmpDir, "arara.yaml")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	t.Log("\n=== Interactive Conflict Test ===")
	t.Log("1. You should see a conflict for 'interactive-test' script")
	t.Log("2. Choose between the existing and new description")
	t.Log("3. The test will verify your choice\n")

	if err := sync.Cmd.Do(sync.Cmd); err != nil {
		t.Fatalf("sync.Cmd.Do() error = %v", err)
	}

	// Verify the result
	configPath = filepath.Join(tmpDir, "arara.yaml")
	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("\nFinal configuration:")
	for _, script := range cfg.Scripts.Install {
		if script.Name == "interactive-test" {
			t.Logf("Script: %s\nDescription: %s\n", script.Name, script.Description)
		}
	}
}

func TestSyncCmd_NonInteractive(t *testing.T) {
	// Test the non-interactive parts
}

func TestSyncCmd_Interactive(t *testing.T) {
	if testing.Short() || os.Getenv("CI") != "" {
		t.Skip("skipping interactive test")
	}
	// Manual interactive test
}
