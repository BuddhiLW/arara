package namespace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/edit"
	"github.com/rwxrob/bonzai/vars"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	v "github.com/BuddhiLW/arara/internal/pkg/vars"
)

// Cmd manages dotfiles namespaces
var Cmd = &bonzai.Cmd{
	Name:  "namespace",
	Alias: "ns",
	Short: "manage dotfiles namespaces",
	Cmds: []*bonzai.Cmd{
		listCmd,
		switchCmd,
		help.Cmd,
		editCmd,
		addCmd,
		removeCmd,
	},
}

// listCmd lists available namespaces
var listCmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "ls",
	Short: "list available namespaces",
	Do: func(x *bonzai.Cmd, args ...string) error {
		gc, err := config.NewGlobalConfig()
		if err != nil {
			return err
		}

		active := vars.Fetch(v.ActiveNamespaceEnv, v.ActiveNamespaceVar, "")

		fmt.Println("Available namespaces:")
		for _, ns := range gc.Config.Namespaces {
			if ns == active {
				fmt.Printf("* %s (active)\n", ns)
			} else {
				fmt.Printf("  %s\n", ns)
			}
		}
		return nil
	},
}

// switchCmd switches the active namespace
var switchCmd = &bonzai.Cmd{
	Name:    "switch",
	Alias:   "sw",
	Short:   "switch active namespace",
	NumArgs: 1,
	Do: func(x *bonzai.Cmd, args ...string) error {
		gc, err := config.NewGlobalConfig()
		if err != nil {
			return err
		}

		ns := args[0]
		info, ok := gc.Config.Configs[ns]
		if !ok {
			return fmt.Errorf("namespace not found: %s", ns)
		}

		vars.Data.Set(v.ActiveNamespaceVar, ns)
		vars.Data.Set(v.DotfilesPathVar, info.Path)

		fmt.Printf("Switched to namespace: %s\n", ns)
		fmt.Printf("Dotfiles path: %s\n", info.Path)
		return nil
	},
}

var editCmd = &bonzai.Cmd{
	Name:  "edit",
	Alias: "e",
	Short: "edit global namespace configuration",
	Long: `
Edit the global namespace configuration file (~/.config/arara/config.yaml).
This file defines available namespaces and their dotfiles paths.

Format:
  namespaces:
    - blw
    - jack

  configs:
    blw:
      path: /path/to/blw-dotfiles
      local-bin: "blw"   # optional
    jack:
      path: /path/to/jack-dotfiles
      local-bin: "jack-bins"
`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		configPath := filepath.Join(config.GetConfigDir(), "config.yaml")

		// Create config file if it doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `# Arara namespace configuration
namespaces: []
configs: {}
`
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}

		return edit.Files(configPath)
	},
}

var addCmd = &bonzai.Cmd{
	Name:    "add",
	Alias:   "a",
	Short:   "add a new namespace",
	Usage:   "add <name> <path>",
	NumArgs: 2,
	Long: `
Add a new namespace to the global configuration.

Arguments:
  name: Name of the namespace
  path: Path to the dotfiles repository

Example:
  arara namespace add work ~/work-dotfiles
  arara namespace add personal ~/dotfiles
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		name := args[0]
		path, err := filepath.Abs(args[1])
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		// Validate path exists
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		gc, err := config.NewGlobalConfig()
		if err != nil {
			return err
		}

		// Check if namespace already exists
		for _, ns := range gc.Config.Namespaces {
			if ns == name {
				return fmt.Errorf("namespace already exists: %s", name)
			}
		}

		// Add namespace
		gc.Config.Namespaces = append(gc.Config.Namespaces, name)
		gc.Config.Configs[name] = config.NSInfo{
			Path: path,
		}

		if err := gc.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Added namespace '%s' pointing to %s\n", name, path)
		return nil
	},
}

var removeCmd = &bonzai.Cmd{
	Name:    "remove",
	Alias:   "rm",
	Short:   "remove a namespace",
	Usage:   "remove <name>",
	NumArgs: 1,
	Long: `
Remove a namespace from the global configuration.
This only removes the namespace from arara's configuration,
it does not delete any files.

Example:
  arara namespace remove work
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		name := args[0]

		gc, err := config.NewGlobalConfig()
		if err != nil {
			return err
		}

		// Check if namespace exists
		found := false
		for i, ns := range gc.Config.Namespaces {
			if ns == name {
				// Remove from slice
				gc.Config.Namespaces = append(gc.Config.Namespaces[:i], gc.Config.Namespaces[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("namespace not found: %s", name)
		}

		// Remove from configs map
		delete(gc.Config.Configs, name)

		if err := gc.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// If this was the active namespace, clear it
		active := vars.Fetch(v.ActiveNamespaceEnv, v.ActiveNamespaceVar, "")
		if active == name {
			vars.Data.Set(v.ActiveNamespaceVar, "")
			vars.Data.Set(v.DotfilesPathVar, "")
			fmt.Println("Cleared active namespace")
		}

		fmt.Printf("Removed namespace '%s'\n", name)
		return nil
	},
}
