package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	bonzaiVars "github.com/rwxrob/bonzai/vars"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/BuddhiLW/arara/internal/pkg/vars"
)

func autoAddNamespace(path string) error {
	// Check if directory contains arara.yaml
	configPath := filepath.Join(path, "arara.yaml")
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("no arara.yaml found in %s", path)
	}

	// Load config to get name
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	// Use directory name as namespace if not specified
	nsName := cfg.Name
	if nsName == "" {
		nsName = filepath.Base(path)
	}

	// Check if namespace already exists
	gc, err := config.NewGlobalConfig()
	if err != nil {
		return err
	}

	for _, ns := range gc.Config.Namespaces {
		if ns == nsName {
			// Already exists, just switch to it
			bonzaiVars.Data.Set(vars.ActiveNamespaceVar, nsName)
			bonzaiVars.Data.Set(vars.DotfilesPathVar, path)
			return nil
		}
	}

	// Add new namespace
	gc.Config.Namespaces = append(gc.Config.Namespaces, nsName)
	gc.Config.Configs[nsName] = config.NSInfo{
		Path: path,
	}

	if err := gc.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Switch to new namespace
	bonzaiVars.Data.Set(vars.ActiveNamespaceVar, nsName)
	bonzaiVars.Data.Set(vars.DotfilesPathVar, path)

	fmt.Printf("Added and switched to namespace '%s' (%s)\n", nsName, path)
	return nil
}

var Cmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "l",
	Short: "list available scripts",
	Long: `
List available installation scripts defined in arara.yaml.
By default, tries to use local arara.yaml if present, otherwise uses active namespace.

Commands:
  local   - Force list from current directory's arara.yaml
  global  - Force list from active namespace configuration
`,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		localCmd,
		globalCmd,
	},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Try local first
		if pwd, err := os.Getwd(); err == nil {
			if _, err := os.Stat("arara.yaml"); err == nil {
				// Found local arara.yaml, try to auto-add namespace
				if err := autoAddNamespace(pwd); err != nil {
					return err
				}
				// List scripts from local config
				cfg, err := config.LoadConfig("arara.yaml")
				if err != nil {
					return fmt.Errorf("failed to load local config: %w", err)
				}

				fmt.Println("Available installation scripts (local):")
				for _, script := range cfg.Scripts.Install {
					fmt.Printf("  %s - %s\n", script.Name, script.Description)
				}
				return nil
			}
		}
		// Fallback to global
		return globalCmd.Do(caller, args...)
	},
}

var localCmd = &bonzai.Cmd{
	Name:  "local",
	Short: "list scripts from local arara.yaml",
	Do: func(caller *bonzai.Cmd, args ...string) error {
		cfg, err := config.LoadConfig("arara.yaml")
		if err != nil {
			return fmt.Errorf("failed to load local config: %w", err)
		}

		fmt.Println("Available installation scripts (local):")
		for _, script := range cfg.Scripts.Install {
			fmt.Printf("  %s - %s\n", script.Name, script.Description)
		}
		return nil
	},
}

var globalCmd = &bonzai.Cmd{
	Name:  "global",
	Short: "list scripts from active namespace",
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Get active namespace first
		activeNS := bonzaiVars.Fetch(vars.ActiveNamespaceEnv, vars.ActiveNamespaceVar, "")
		if activeNS == "" {
			return fmt.Errorf("no active namespace set. Use 'arara namespace switch <name>' first")
		}

		// Get dotfiles path from active namespace
		dotfilesPath := config.GetDotfilesPath()
		if dotfilesPath == "" {
			return fmt.Errorf("no dotfiles path found for namespace: %s", activeNS)
		}

		cfg, err := config.LoadConfig(filepath.Join(dotfilesPath, "arara.yaml"))
		if err != nil {
			return fmt.Errorf("failed to load config for namespace %s: %w", activeNS, err)
		}

		fmt.Printf("Available installation scripts (namespace: %s):\n", activeNS)
		for _, script := range cfg.Scripts.Install {
			fmt.Printf("  %s - %s\n", script.Name, script.Description)
		}
		return nil
	},
}
