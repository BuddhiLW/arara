package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

// Cmd represents the build command
var Cmd = &bonzai.Cmd{
	Name:  "build",
	Alias: "b",
	Short: "execute or list build steps from arara.yaml",
	Cmds:  []*bonzai.Cmd{help.Cmd, listCmd, installCmd},
}

// listCmd lists all build steps from arara.yaml
var listCmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "ls",
	Short: "list available build steps",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// TODO: Read from arara.yaml and display build steps
		fmt.Println("Available build steps:")
		fmt.Println("  - backup: Backup existing dotfiles")
		fmt.Println("  - link: Create symlinks")
		fmt.Println("  - xmonad: Setup window manager")
		return nil
	},
}

// installCmd executes all build steps from arara.yaml
var installCmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "execute fresh dotfiles installation",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("Executing build steps...")
		
		// Execute backup step
		fmt.Println("1. Backing up existing dotfiles...")
		backupCmd := exec.Command("arara", "setup", "backup")
		backupCmd.Stdout = os.Stdout
		backupCmd.Stderr = os.Stderr
		if err := backupCmd.Run(); err != nil {
			return fmt.Errorf("failed to backup existing dotfiles: %w", err)
		}
		
		// Execute link step
		fmt.Println("2. Creating symlinks...")
		linkCmd := exec.Command("arara", "setup", "link")
		linkCmd.Stdout = os.Stdout
		linkCmd.Stderr = os.Stderr
		if err := linkCmd.Run(); err != nil {
			return fmt.Errorf("failed to create symlinks: %w", err)
		}
		
		// Execute xmonad setup step
		fmt.Println("3. Setting up window manager...")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		
		xmonadConfigDir := filepath.Join(homeDir, ".config", "xmonad")
		if err := os.Chdir(xmonadConfigDir); err != nil {
			return fmt.Errorf("failed to change to xmonad config directory: %w", err)
		}
		
		// Remove existing xmonad repos
		if err := os.RemoveAll("xmonad"); err != nil {
			return fmt.Errorf("failed to remove existing xmonad repo: %w", err)
		}
		if err := os.RemoveAll("xmonad-contrib"); err != nil {
			return fmt.Errorf("failed to remove existing xmonad-contrib repo: %w", err)
		}
		
		// Clone xmonad repositories
		xmonadCmd := exec.Command("git", "clone", "https://github.com/xmonad/xmonad")
		xmonadCmd.Stdout = os.Stdout
		xmonadCmd.Stderr = os.Stderr
		if err := xmonadCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone xmonad repository: %w", err)
		}
		
		xmonadContribCmd := exec.Command("git", "clone", "https://github.com/xmonad/xmonad-contrib")
		xmonadContribCmd.Stdout = os.Stdout
		xmonadContribCmd.Stderr = os.Stderr
		if err := xmonadContribCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone xmonad-contrib repository: %w", err)
		}
		
		// Install Haskell Stack
		stackCmd := exec.Command("bash", "-c", "curl -sSL https://get.haskellstack.org/ | sh -s - -f")
		stackCmd.Stdout = os.Stdout
		stackCmd.Stderr = os.Stderr
		if err := stackCmd.Run(); err != nil {
			return fmt.Errorf("failed to install Haskell Stack: %w", err)
		}
		
		fmt.Println("Build completed successfully!")
		return nil
	},
}