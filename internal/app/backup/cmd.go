package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
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

			// Move the directory
			if err := os.Rename(expandedDir, dst); err != nil {
				return fmt.Errorf("failed to backup %s: %w", expandedDir, err)
			}
			fmt.Printf("Backed up %s to %s\n", expandedDir, dst)
		}

		fmt.Printf("Backup created at: %s\n", backupDir)
		return nil
	},
}
