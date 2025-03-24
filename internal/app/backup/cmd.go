package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

var Cmd = &bonzai.Cmd{
	Name:  "backup",
	Alias: "bk",
	Short: "Backup existing dotfiles",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		backupDir := filepath.Join(os.Getenv("HOME"),
			fmt.Sprintf("dotbk-%d", time.Now().Unix()))

		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup dir: %w", err)
		}

		// Move .config and .local
		dirsToBackup := []string{".config", ".local"}
		for _, dir := range dirsToBackup {
			src := filepath.Join(os.Getenv("HOME"), dir)
			dst := filepath.Join(backupDir, dir)

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to backup %s: %w", dir, err)
			}
		}

		fmt.Printf("Backup created at: %s\n", backupDir)
		return nil
	},
}
