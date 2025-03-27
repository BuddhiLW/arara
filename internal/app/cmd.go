package app

import (
	"fmt"
	"os"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
	"gopkg.in/yaml.v3"

	"github.com/BuddhiLW/arara/internal/app/backup"
	"github.com/BuddhiLW/arara/internal/app/build"
	"github.com/BuddhiLW/arara/internal/app/compat"
	"github.com/BuddhiLW/arara/internal/app/create"
	"github.com/BuddhiLW/arara/internal/app/install"
	"github.com/BuddhiLW/arara/internal/app/link"
	"github.com/BuddhiLW/arara/internal/app/list"
	"github.com/BuddhiLW/arara/internal/app/namespace"
	"github.com/BuddhiLW/arara/internal/app/setup"
	"github.com/BuddhiLW/arara/internal/app/sync"
	"github.com/BuddhiLW/arara/internal/pkg/config"
)

const (
	// Environment variables
	ActiveNamespaceEnv = "ARARA_ACTIVE_NAMESPACE"
	DotfilesPathEnv    = "ARARA_DOTFILES_PATH"

	// Variable names
	ActiveNamespaceVar = "active-namespace"
	DotfilesPathVar    = "dotfiles-path"
)

// Placeholder commands - will be implemented later
var installCmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "install additional tools",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("Install command - to be implemented")
		return nil
	},
}

var initCmd = &bonzai.Cmd{
	Name:  "init",
	Short: "initialize new arara.yaml",
	Long: `
The init command creates a new arara.yaml configuration file in the current
directory. This file serves as the central configuration for your dotfiles
management setup.

# Usage
  arara init

# Configuration
The generated configuration includes:
- Basic metadata (name, description)
- Setup configuration (backup directories and symbolic links)
- Build steps (including backup and link commands)
- Template for installation scripts
- Environment variables

This provides a starting point that you can customize for your specific
dotfiles management needs.
`,
	Cmds: []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Check if arara.yaml already exists
		if _, err := os.Stat("arara.yaml"); err == nil {
			return fmt.Errorf("arara.yaml already exists in the current directory")
		}

		// Create a default configuration
		config := createDefaultConfig()

		// Create file
		file, err := os.Create("arara.yaml")
		if err != nil {
			return fmt.Errorf("failed to create arara.yaml: %w", err)
		}
		defer file.Close()

		// Use YAML encoder for better formatting
		encoder := yaml.NewEncoder(file)
		encoder.SetIndent(2)

		if err := encoder.Encode(config); err != nil {
			return fmt.Errorf("failed to encode config: %w", err)
		}

		fmt.Println("Created arara.yaml in the current directory")
		fmt.Println("Customize it to match your dotfiles setup")
		return nil
	},
}

// createDefaultConfig creates a default configuration structure
func createDefaultConfig() config.Config {
	var conf config.Config

	// Basic metadata
	conf.Name = "dotfiles"
	conf.Description = "Personal dotfiles configuration"

	// Environment variables
	conf.Env = map[string]string{
		"DOTFILES": "$HOME/dotfiles",
		"SCRIPTS":  "$DOTFILES/scripts",
	}

	// Setup configuration
	conf.Setup.BackupDirs = []string{
		"$HOME/.config",
		"$HOME/.local",
	}

	conf.Setup.CoreLinks = []config.Link{
		{
			Source: "$DOTFILES/.config",
			Target: "$HOME/.config",
		},
		{
			Source: "$DOTFILES/.local",
			Target: "$HOME/.local",
		},
	}

	conf.Setup.ConfigLinks = []config.Link{
		{
			Source: "$DOTFILES/.bashrc",
			Target: "$HOME/.bashrc",
		},
		{
			Source: "$DOTFILES/.vim",
			Target: "$HOME/.vim",
		},
		{
			Source: "$DOTFILES/.config/vim/.vimrc",
			Target: "$HOME/.vimrc",
		},
	}

	// Build steps
	conf.Build.Steps = []config.Step{
		{
			Name:        "backup",
			Description: "Backup existing dotfiles",
			Command:     "arara setup backup",
		},
		{
			Name:        "link",
			Description: "Create symlinks",
			Command:     "arara setup link",
		},
		{
			Name:        "example-multi-command",
			Description: "Example of multiple commands",
			Commands: []string{
				"mkdir -p $HOME/.config/example",
				"touch $HOME/.config/example/config",
				"echo 'Setup complete' > $HOME/.config/example/status",
			},
			Compat: &config.CompatConfig{
				OS: "linux",
			},
		},
	}

	// Installation scripts
	conf.Scripts.Install = []config.Script{
		{
			Name:        "example",
			Description: "Example installation script",
			Path:        "scripts/install/example",
		},
		{
			Name:        "docker",
			Description: "Install Docker and Docker Desktop",
			Path:        "scripts/install/docker",
			Compat: &config.CompatConfig{
				OS:     "linux",
				Shell:  "bash",
				PkgMgr: "apt",
				Custom: []interface{}{
					map[string]interface{}{
						"name":  "min-memory",
						"value": 4096,
					},
				},
			},
		},
	}

	return conf
}

// Cmd defines the root command for arara CLI
var Cmd = &bonzai.Cmd{
	Name: "arara",
	Cmds: []*bonzai.Cmd{
		backup.Cmd,    // Backup dotfiles
		build.Cmd,     // Execute build steps
		compat.Cmd,    // Check system compatibility
		create.Cmd,    // Create new resources
		help.Cmd,      // Show help information
		install.Cmd,   // Install additional tools
		link.Cmd,      // Create symlinks
		list.Cmd,      // List available scripts
		namespace.Cmd, // Manage namespaces
		setup.Cmd,     // Core setup operations
		sync.Cmd,      // Sync install scripts
	},
	Alias: "ar",
	Vers:  "v0.1.0",
	Short: "dotfiles management tool",
	Long: `Arara is a CLI tool for managing multiple dotfiles installations and configurations.

# Commands:
- build:     Execute or list build steps
- compat:    Check system compatibility for scripts
- create:    Create new resources (install scripts, build steps)
- install:   Install additional tools
- setup:     Core setup operations (backup, link, restore)
- list:      List available installation scripts
- init:      Initialize new arara.yaml configuration
- namespace: Manage and switch between dotfiles namespaces
- help:      Show this help message

Use 'arara help <command> <subcommand>...' for detailed information
about each command.`,
	Vars: bonzai.Vars{
		{
			K: ActiveNamespaceVar,
			V: "",
			E: ActiveNamespaceEnv,
			S: "Currently active dotfiles namespace",
		},
		{
			K: DotfilesPathVar,
			V: "",
			E: DotfilesPathEnv,
			S: "Path to active dotfiles repository",
		},
	},
	Init: func(x *bonzai.Cmd, args ...string) error {
		// Load global config
		gc, err := config.NewGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		// Get active namespace from env/var or use first available
		ns := bonzaiVars.Fetch(ActiveNamespaceEnv, ActiveNamespaceVar, "")
		if ns == "" && len(gc.Config.Namespaces) > 0 {
			ns = gc.Config.Namespaces[0]
			bonzaiVars.Data.Set(ActiveNamespaceVar, ns)
		}

		// Set dotfiles path if namespace is active
		if ns != "" {
			if info, ok := gc.Config.Configs[ns]; ok {
				bonzaiVars.Data.Set(DotfilesPathVar, info.Path)
			}
		}

		return nil
	},
	Comp: comp.Cmds,
}
