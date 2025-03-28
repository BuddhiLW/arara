package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/futil"
)

var Cmd = &bonzai.Cmd{
	Name:  "backup",
	Alias: "bk",
	Short: "backup existing dotfiles",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Load configuration
		cfg, err := config.LoadConfig("arara.yaml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create backup directory with timestamp
		backupDir := filepath.Join(os.Getenv("HOME"),
			fmt.Sprintf("dotbk-%d", time.Now().Unix()))

		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup dir: %w", err)
		}

		// Backup directories specified in config
		for _, dir := range cfg.Setup.BackupDirs {
			// Expand environment variables in path
			expandedDir := os.ExpandEnv(dir)

			// Get the base name of the directory
			baseName := filepath.Base(expandedDir)

			// Create destination path
			dst := filepath.Join(backupDir, baseName)

			// Skip if source doesn't exist
			if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
				fmt.Printf("Skipping non-existent directory: %s\n", expandedDir)
				continue
			}

			// Try renaming first (faster if on same filesystem)
			err := os.Rename(expandedDir, dst)
			if err != nil {
				// If rename fails, try copying
				if err := futil.Replace(dst, expandedDir); err != nil {
					return fmt.Errorf("failed to backup %s: %w", expandedDir, err)
				}
				// After successful copy, remove the original
				if err := os.RemoveAll(expandedDir); err != nil {
					return fmt.Errorf("failed to remove original after backup %s: %w", expandedDir, err)
				}
			}
			fmt.Printf("Backed up %s to %s\n", expandedDir, dst)
		}

		fmt.Printf("Backup created at: %s\n", backupDir)
		return nil
	},
}

// copyDir recursively copies a directory tree
func copyDir(src string, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination dir
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursive copy for directories
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents and preserve mode
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
