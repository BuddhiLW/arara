package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/arara/internal/pkg/config"
)

func TestReadDepsFile(t *testing.T) {
	// Create a temporary file with dependencies
	content := `
# This is a comment
git
# Another comment
vim
tmux
# Final comment
curl
# Multiple deps on one line
- lolcat
- fortune fortune-mod fortunes-br display-dhammapada fortune-anarchism fortune-mod
- gawk
- qutebrowser
`
	tmpfile, err := os.CreateTemp("", "deps-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Test reading the dependencies file
	deps, err := readDepsFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read dependencies file: %v", err)
	}

	// Check that comments were properly filtered
	expected := []string{
		"git", "vim", "tmux", "curl", 
		"lolcat", "fortune", "fortune-mod", "fortunes-br", "display-dhammapada", 
		"fortune-anarchism", "fortune-mod", "gawk", "qutebrowser",
	}
	
	// Create map for easier checking
	expectedMap := make(map[string]bool)
	for _, dep := range expected {
		expectedMap[dep] = true
	}
	
	// Count the actual number of unique dependencies in our expected list
	expectedCount := len(expectedMap)
	
	// Create map of actual dependencies for comparison
	depsMap := make(map[string]bool)
	for _, dep := range deps {
		depsMap[dep] = true
	}
	
	// Check that we got the right number of unique dependencies
	if len(depsMap) != expectedCount {
		t.Errorf("Expected %d unique dependencies, got %d", expectedCount, len(depsMap))
	}
	
	// Check that all expected dependencies are present
	for _, dep := range deps {
		if !expectedMap[dep] {
			t.Errorf("Unexpected dependency found: %s", dep)
		}
	}
}

// MockTransaction is a test helper to avoid file system operations
type MockTransaction struct {
	Modified bool
}

func (t *MockTransaction) commit() error {
	return nil
}

func (t *MockTransaction) rollback() error {
	return nil
}

func (t *MockTransaction) checkModified() (bool, error) {
	return t.Modified, nil
}

// Mock the beginTransaction function for testing
func mockBeginTransaction(path string) (*transaction, error) {
	return &transaction{
		configPath: path,
		backupPath: path + ".bak",
		origHash:   []byte("testhash"),
	}, nil
}

func TestDetectPackageManager(t *testing.T) {
	// This test is limited since we can't easily mock exec.LookPath
	// We'll just verify it doesn't error on the current system
	_, err := detectPackageManager()
	if err != nil && !strings.Contains(err.Error(), "no supported package manager found") {
		t.Errorf("Unexpected error from detectPackageManager: %v", err)
	}
}

func TestPackageManagerCommands(t *testing.T) {
	// Test that the package manager commands are properly constructed
	testCases := []struct {
		manager PackageManager
		deps    []string
		want    []string
	}{
		{
			manager: packageManagers["apt"],
			deps:    []string{"git", "vim"},
			want:    []string{"sudo", "apt-get", "install", "git", "vim"},
		},
		{
			manager: packageManagers["pacman"],
			deps:    []string{"git", "vim"},
			want:    []string{"sudo", "pacman", "-S", "git", "vim"},
		},
		{
			manager: packageManagers["brew"],
			deps:    []string{"git", "vim"},
			want:    []string{"brew", "install", "git", "vim"},
		},
	}

	for _, tc := range testCases {
		var cmdArgs []string
		cmdArgs = append(cmdArgs, tc.manager.InstallPrefix...)
		cmdArgs = append(cmdArgs, tc.manager.InstallCmd)
		cmdArgs = append(cmdArgs, tc.deps...)

		if len(cmdArgs) != len(tc.want) {
			t.Errorf("Expected %d command args, got %d", len(tc.want), len(cmdArgs))
			continue
		}

		for i, arg := range tc.want {
			if cmdArgs[i] != arg {
				t.Errorf("Expected cmd arg %s at position %d, got %s", arg, i, cmdArgs[i])
			}
		}
	}
}

// Setup helpers for testing without accessing the file system
func setupTestConfigAndVars(t *testing.T) (func(), string) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "arara-deps-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a test config file
	configPath := filepath.Join(tmpDir, "arara.yaml")
	cfg := config.DotfilesConfig{
		Name:        "test-namespace",
		Description: "Test configuration",
		Dependencies: []string{
			"git",
			"vim",
		},
	}

	data, err := cfg.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return cleanup, tmpDir
}