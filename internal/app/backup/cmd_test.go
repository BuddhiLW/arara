package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupCmd(t *testing.T) {
	// Save original HOME environment variable to restore later
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create a temporary test environment
	tmpDir := t.TempDir()
	
	// Set HOME to our test directory
	os.Setenv("HOME", tmpDir)
	
	// Create test directories that should be backed up
	configDir := filepath.Join(tmpDir, ".config")
	localDir := filepath.Join(tmpDir, ".local")
	
	// Create test files inside these directories
	if err := os.MkdirAll(filepath.Join(configDir, "test"), 0755); err != nil {
		t.Fatalf("Failed to create test .config directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "test", "config.txt"), []byte("test config"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	if err := os.MkdirAll(filepath.Join(localDir, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create test .local directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "bin", "testscript.sh"), []byte("#!/bin/bash\necho 'test'"), 0755); err != nil {
		t.Fatalf("Failed to create test local file: %v", err)
	}
	
	// Execute the backup command
	err := Cmd.Do(Cmd, []string{}...)
	if err != nil {
		t.Fatalf("Failed to execute backup command: %v", err)
	}
	
	// Verify backup was created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read home directory: %v", err)
	}
	
	// Find the backup directory (should start with dotbk-)
	var backupDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "dotbk-") {
			backupDir = filepath.Join(tmpDir, entry.Name())
			break
		}
	}
	
	if backupDir == "" {
		t.Fatal("No backup directory was created")
	}
	
	// Verify backed up directories exist
	if _, err := os.Stat(filepath.Join(backupDir, ".config", "test", "config.txt")); err != nil {
		t.Errorf("Backed up .config file not found: %v", err)
	}
	
	if _, err := os.Stat(filepath.Join(backupDir, ".local", "bin", "testscript.sh")); err != nil {
		t.Errorf("Backed up .local file not found: %v", err)
	}
	
	// Verify original directories were moved (no longer exist)
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Errorf("Original .config directory still exists, expected it to be moved")
	}
	
	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		t.Errorf("Original .local directory still exists, expected it to be moved")
	}
}