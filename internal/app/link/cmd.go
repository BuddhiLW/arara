package link

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

var Cmd = &bonzai.Cmd{
	Name:  "link",
	Alias: "ln",
	Short: "Create symlinks for dotfiles",
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
			if err := os.Symlink(link.src, link.dst); err != nil {
				return fmt.Errorf("failed to create link %s -> %s: %w", link.src, link.dst, err)
			}
			fmt.Printf("Created symlink: %s -> %s\n", link.dst, link.src)
		}

		return nil
	},
}