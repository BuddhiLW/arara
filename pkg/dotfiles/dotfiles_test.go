package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Test constructor
	configPath := "/test/config.yaml"
	dotfilesDir := "/test/dotfiles"
	
	manager := New(configPath, dotfilesDir)
	
	if manager.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath to be %s, got %s", configPath, manager.ConfigPath)
	}
	
	if manager.DotfilesDir != dotfilesDir {
		t.Errorf("Expected DotfilesDir to be %s, got %s", dotfilesDir, manager.DotfilesDir)
	}
}

func TestCreateSymlink(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a manager instance
	manager := New("config.yaml", tmpDir)
	
	// Create a source file
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	
	sourceFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(sourceFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Test creating a symlink
	t.Run("CreateSymlink", func(t *testing.T) {
		targetDir := filepath.Join(tmpDir, "target")
		targetFile := filepath.Join(targetDir, "test_link.txt")
		
		err := manager.CreateSymlink(sourceFile, targetFile)
		if err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
		
		// Verify the symlink was created
		linkDest, err := os.Readlink(targetFile)
		if err != nil {
			t.Fatalf("Failed to read symlink: %v", err)
		}
		
		if linkDest != sourceFile {
			t.Errorf("Expected symlink to point to %s, got %s", sourceFile, linkDest)
		}
	})
	
	// Test overwriting an existing symlink
	t.Run("OverwriteExistingSymlink", func(t *testing.T) {
		// Create a temporary file to link to initially
		tempFile := filepath.Join(tmpDir, "temp.txt")
		if err := os.WriteFile(tempFile, []byte("temp content"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		
		// Create the target symlink pointing to the temp file
		targetFile := filepath.Join(tmpDir, "existing_link.txt")
		if err := os.Symlink(tempFile, targetFile); err != nil {
			t.Fatalf("Failed to create initial symlink: %v", err)
		}
		
		// Now update it to point to our source file
		err := manager.CreateSymlink(sourceFile, targetFile)
		if err != nil {
			t.Fatalf("Failed to update existing symlink: %v", err)
		}
		
		// Verify the symlink was updated
		linkDest, err := os.Readlink(targetFile)
		if err != nil {
			t.Fatalf("Failed to read updated symlink: %v", err)
		}
		
		if linkDest != sourceFile {
			t.Errorf("Expected updated symlink to point to %s, got %s", sourceFile, linkDest)
		}
	})
	
	// Test creating parent directories
	t.Run("CreateParentDirectories", func(t *testing.T) {
		nestedTargetDir := filepath.Join(tmpDir, "nested", "target", "dir")
		nestedTargetFile := filepath.Join(nestedTargetDir, "nested_link.txt")
		
		err := manager.CreateSymlink(sourceFile, nestedTargetFile)
		if err != nil {
			t.Fatalf("Failed to create symlink with nested directories: %v", err)
		}
		
		// Verify the symlink was created
		linkDest, err := os.Readlink(nestedTargetFile)
		if err != nil {
			t.Fatalf("Failed to read nested symlink: %v", err)
		}
		
		if linkDest != sourceFile {
			t.Errorf("Expected nested symlink to point to %s, got %s", sourceFile, linkDest)
		}
	})
}