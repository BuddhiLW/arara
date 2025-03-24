package dotfiles

import (
	"os"
	"path/filepath"
)

// Manager handles dotfiles operations
type Manager struct {
	ConfigPath  string
	DotfilesDir string
}

// New creates a new dotfiles manager
func New(configPath, dotfilesDir string) *Manager {
	return &Manager{
		ConfigPath:  configPath,
		DotfilesDir: dotfilesDir,
	}
}

// CreateSymlink creates a symlink with proper error handling
func (m *Manager) CreateSymlink(source, target string) error {
	// Remove existing if it's a symlink
	if info, err := os.Lstat(target); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			os.Remove(target)
		}
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	return os.Symlink(source, target)
}
