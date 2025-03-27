package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGlobalConfig(t *testing.T) {
	// Save original HOME environment variable
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	gc, err := NewGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	if gc.config.Configs == nil {
		t.Error("Expected Configs map to be initialized")
	}

	if gc.persister == nil {
		t.Error("Expected persister to be initialized")
	}
}

func TestAddNamespace(t *testing.T) {
	// Create temporary test environment
	tmpDir := t.TempDir()
	dotfilesPath := filepath.Join(tmpDir, "dotfiles")
	if err := os.MkdirAll(dotfilesPath, 0755); err != nil {
		t.Fatalf("Failed to create test dotfiles directory: %v", err)
	}

	gc, err := NewGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	tests := []struct {
		name     string
		nsName   string
		path     string
		localBin string
		wantErr  bool
	}{
		{
			name:     "Valid namespace",
			nsName:   "test",
			path:     dotfilesPath,
			localBin: "test-bin",
			wantErr:  false,
		},
		{
			name:     "Invalid path",
			nsName:   "invalid",
			path:     "/nonexistent/path",
			localBin: "test-bin",
			wantErr:  true,
		},
		{
			name:     "Default local-bin",
			nsName:   "default",
			path:     dotfilesPath,
			localBin: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gc.AddNamespace(tt.nsName, tt.path, tt.localBin)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify namespace was added
				found := false
				for _, ns := range gc.config.Namespaces {
					if ns == tt.nsName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Namespace %s not found in namespaces list", tt.nsName)
				}

				// Verify config was added
				info, exists := gc.config.Configs[tt.nsName]
				if !exists {
					t.Errorf("Config for namespace %s not found", tt.nsName)
					return
				}

				if info.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, info.Path)
				}

				expectedBin := tt.localBin
				if expectedBin == "" {
					expectedBin = tt.nsName
				}
				if info.LocalBin != expectedBin {
					t.Errorf("Expected local-bin %s, got %s", expectedBin, info.LocalBin)
				}
			}
		})
	}
}

func TestUpdateShellRC(t *testing.T) {
	// Create temporary test environment
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create test home directory: %v", err)
	}

	// Set HOME environment variable
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", homeDir)

	// Create test dotfiles directories with bin folders
	dotfiles1 := filepath.Join(tmpDir, "dotfiles1")
	dotfiles2 := filepath.Join(tmpDir, "dotfiles2")
	bin1 := filepath.Join(dotfiles1, ".local/bin/ns1")
	bin2 := filepath.Join(dotfiles2, ".local/bin/ns2")

	for _, dir := range []string{bin1, bin2} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create bin directory %s: %v", dir, err)
		}
	}

	// Create test bashrc
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	origContent := "# Original bashrc content\nexport PATH=\"/usr/local/bin:$PATH\"\n"
	if err := os.WriteFile(bashrcPath, []byte(origContent), 0644); err != nil {
		t.Fatalf("Failed to create test bashrc: %v", err)
	}

	// Initialize global config
	gc, err := NewGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	// Add test namespaces
	if err := gc.AddNamespace("ns1", dotfiles1, "ns1"); err != nil {
		t.Fatalf("Failed to add namespace 1: %v", err)
	}
	if err := gc.AddNamespace("ns2", dotfiles2, "ns2"); err != nil {
		t.Fatalf("Failed to add namespace 2: %v", err)
	}

	// Test updating bashrc
	if err := gc.UpdateShellRC(); err != nil {
		t.Fatalf("Failed to update shell RC: %v", err)
	}

	// Read updated bashrc
	content, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("Failed to read updated bashrc: %v", err)
	}

	// Verify content
	updatedContent := string(content)

	// Check that original content is preserved
	if !strings.Contains(updatedContent, origContent) {
		t.Error("Original bashrc content was not preserved")
	}

	// Check that Arara section exists
	if !strings.Contains(updatedContent, "<<<< Added by Arara") {
		t.Error("Arara section start marker not found")
	}
	if !strings.Contains(updatedContent, ">>>> End Arara section") {
		t.Error("Arara section end marker not found")
	}

	// Check that both bin paths are added
	for _, binPath := range []string{bin1, bin2} {
		if !strings.Contains(updatedContent, fmt.Sprintf("export PATH=\"%s:$PATH\"", binPath)) {
			t.Errorf("Expected bin path %s not found in bashrc", binPath)
		}
	}

	// Test updating again (should replace existing section)
	if err := gc.UpdateShellRC(); err != nil {
		t.Fatalf("Failed to update shell RC second time: %v", err)
	}

	// Read updated content again
	content, err = os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("Failed to read updated bashrc: %v", err)
	}

	// Count Arara sections (should only be one)
	count := strings.Count(string(content), "<<<< Added by Arara")
	if count != 1 {
		t.Errorf("Expected exactly one Arara section, found %d", count)
	}
}
