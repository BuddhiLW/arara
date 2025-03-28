package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

// shouldRemoveExisting checks if dst exists, is non-empty, and if a backup directory
// (starting with "dotbk-") exists in home. If so, returns true.
func shouldRemoveExisting(dst, home string) bool {
	info, err := os.Stat(dst)
	if err != nil {
		// dst doesn't exist
		return false
	}
	if !info.IsDir() {
		// if it's not a directory, we consider removal in the caller if needed; here we focus on directories
		return false
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		return false
	}
	if len(entries) == 0 {
		// empty directory is safe
		return false
	}

	// Check for existence of a backup directory in home (backup directories begin with "dotbk-")
	homeEntries, err := os.ReadDir(home)
	if err != nil {
		return false
	}
	for _, entry := range homeEntries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "dotbk-") {
			return true
		}
	}
	return false
}

var Cmd = &bonzai.Cmd{
	Name:  "link",
	Alias: "ln",
	Short: "create symlinks for dotfiles",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		home := os.Getenv("HOME")
		dotfiles := os.Getenv("DOTFILES")
		if dotfiles == "" {
			dotfiles = filepath.Join(home, "dotfiles")
		}

		// Core directory links
		coreLinks := []struct {
			src string
			dst string
		}{
			{filepath.Join(dotfiles, ".config"), filepath.Join(home, ".config")},
			{filepath.Join(dotfiles, ".local"), filepath.Join(home, ".local")},
		}

		for _, link := range coreLinks {
			// if destination exists and is non-empty, and a backup exists, remove it first
			if _, err := os.Lstat(link.dst); err == nil {
				if shouldRemoveExisting(link.dst, home) {
					if err := os.RemoveAll(link.dst); err != nil {
						return fmt.Errorf("failed to remove existing directory %s: %w", link.dst, err)
					}
				}
			}

			if err := os.Symlink(link.src, link.dst); err != nil {
				return fmt.Errorf("failed to create link %s -> %s: %w", link.src, link.dst, err)
			}
			fmt.Printf("Created symlink: %s -> %s\n", link.dst, link.src)
		}

		// Config file links
		configLinks := []struct {
			src string
			dst string
		}{
			{filepath.Join(dotfiles, ".bashrc"), filepath.Join(home, ".bashrc")},
			{filepath.Join(dotfiles, ".vim"), filepath.Join(home, ".vim")},
			{filepath.Join(dotfiles, ".doom.d"), filepath.Join(home, ".doom.d")},
			{filepath.Join(dotfiles, ".config/tmux/.tmux.conf"), filepath.Join(home, ".tmux.conf")},
			{filepath.Join(dotfiles, ".config/vim/.vimrc"), filepath.Join(home, ".vimrc")},
			{filepath.Join(dotfiles, ".config/X11/xinitrc"), filepath.Join(home, ".xinitrc")},
		}

		for _, link := range configLinks {
			// For config links, if a file or symlink already exists, remove it.
			if _, err := os.Lstat(link.dst); err == nil {
				if err := os.RemoveAll(link.dst); err != nil {
					return fmt.Errorf("failed to remove existing file/directory %s: %w", link.dst, err)
				}
			}
			if err := os.Symlink(link.src, link.dst); err != nil {
				return fmt.Errorf("failed to create link %s -> %s: %w", link.src, link.dst, err)
			}
			fmt.Printf("Created symlink: %s -> %s\n", link.dst, link.src)
		}

		return nil
	},
}
