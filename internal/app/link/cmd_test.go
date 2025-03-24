package link

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkCmd(t *testing.T) {
	// Save original environment variables to restore later
	originalHome := os.Getenv("HOME")
	originalDotfiles := os.Getenv("DOTFILES")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("DOTFILES", originalDotfiles)
	}()

	// Create a temporary test environment
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	dotfilesDir := filepath.Join(tmpDir, "dotfiles")
	
	// Set environment variables for testing
	os.Setenv("HOME", homeDir)
	os.Setenv("DOTFILES", dotfilesDir)
	
	// Create necessary directories and files in dotfiles directory
	
	// Create .config and .local directories
	configSrcDir := filepath.Join(dotfilesDir, ".config")
	if err := os.MkdirAll(filepath.Join(configSrcDir, "test"), 0755); err != nil {
		t.Fatalf("Failed to create .config directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configSrcDir, "test", "config.txt"), []byte("test config"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	localSrcDir := filepath.Join(dotfilesDir, ".local")
	if err := os.MkdirAll(filepath.Join(localSrcDir, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create .local directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSrcDir, "bin", "script.sh"), []byte("echo 'hello'"), 0755); err != nil {
		t.Fatalf("Failed to create test local file: %v", err)
	}
	
	// Create config files
	for _, file := range []string{".bashrc", ".vim", ".doom.d"} {
		srcFile := filepath.Join(dotfilesDir, file)
		if file == ".vim" || file == ".doom.d" {
			// These are directories
			if err := os.MkdirAll(srcFile, 0755); err != nil {
				t.Fatalf("Failed to create %s directory: %v", file, err)
			}
		} else {
			// Create the file
			if err := os.WriteFile(srcFile, []byte("# Test "+file), 0644); err != nil {
				t.Fatalf("Failed to create %s file: %v", file, err)
			}
		}
	}
	
	// Create nested config files
	tmuxConfDir := filepath.Join(dotfilesDir, ".config", "tmux")
	vimrcDir := filepath.Join(dotfilesDir, ".config", "vim")
	x11Dir := filepath.Join(dotfilesDir, ".config", "X11")
	
	if err := os.MkdirAll(tmuxConfDir, 0755); err != nil {
		t.Fatalf("Failed to create tmux config directory: %v", err)
	}
	if err := os.MkdirAll(vimrcDir, 0755); err != nil {
		t.Fatalf("Failed to create vim config directory: %v", err)
	}
	if err := os.MkdirAll(x11Dir, 0755); err != nil {
		t.Fatalf("Failed to create X11 config directory: %v", err)
	}
	
	if err := os.WriteFile(filepath.Join(tmuxConfDir, ".tmux.conf"), []byte("# Test tmux config"), 0644); err != nil {
		t.Fatalf("Failed to create tmux config file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vimrcDir, ".vimrc"), []byte("\" Test vimrc"), 0644); err != nil {
		t.Fatalf("Failed to create vimrc file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(x11Dir, "xinitrc"), []byte("# Test xinitrc"), 0644); err != nil {
		t.Fatalf("Failed to create xinitrc file: %v", err)
	}
	
	// Create user's home directory
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Execute the link command
	err := Cmd.Do(Cmd, []string{}...)
	if err != nil {
		t.Fatalf("Failed to execute link command: %v", err)
	}
	
	// Verify core links were created
	for _, dir := range []struct {
		src string
		dst string
	}{
		{configSrcDir, filepath.Join(homeDir, ".config")},
		{localSrcDir, filepath.Join(homeDir, ".local")},
	} {
		verifySymlink(t, dir.src, dir.dst)
	}
	
	// Verify config links were created
	for _, file := range []struct {
		src string
		dst string
	}{
		{filepath.Join(dotfilesDir, ".bashrc"), filepath.Join(homeDir, ".bashrc")},
		{filepath.Join(dotfilesDir, ".vim"), filepath.Join(homeDir, ".vim")},
		{filepath.Join(dotfilesDir, ".doom.d"), filepath.Join(homeDir, ".doom.d")},
		{filepath.Join(dotfilesDir, ".config/tmux/.tmux.conf"), filepath.Join(homeDir, ".tmux.conf")},
		{filepath.Join(dotfilesDir, ".config/vim/.vimrc"), filepath.Join(homeDir, ".vimrc")},
		{filepath.Join(dotfilesDir, ".config/X11/xinitrc"), filepath.Join(homeDir, ".xinitrc")},
	} {
		verifySymlink(t, file.src, file.dst)
	}
}

// Helper function to verify a symlink
func verifySymlink(t *testing.T, src, dst string) {
	t.Helper()
	
	// Check if target exists
	if _, err := os.Lstat(dst); err != nil {
		t.Errorf("Symlink target %s doesn't exist: %v", dst, err)
		return
	}
	
	// Check if it's a symlink
	info, err := os.Lstat(dst)
	if err != nil {
		t.Errorf("Failed to get file info for %s: %v", dst, err)
		return
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Expected %s to be a symlink, but it's not", dst)
		return
	}
	
	// Read the link target
	linkDest, err := os.Readlink(dst)
	if err != nil {
		t.Errorf("Failed to read symlink %s: %v", dst, err)
		return
	}
	
	if linkDest != src {
		t.Errorf("Symlink %s points to %s, expected %s", dst, linkDest, src)
	}
}