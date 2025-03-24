package app

import (
	"fmt"
	
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	
	"github.com/BuddhiLW/arara/internal/app/build"
	"github.com/BuddhiLW/arara/internal/app/create"
	"github.com/BuddhiLW/arara/internal/app/setup"
)

// Placeholder commands - will be implemented later
var installCmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "Install additional tools",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("Install command - to be implemented")
		return nil
	},
}

var listCmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "l",
	Short: "List available scripts",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("List command - to be implemented")
		return nil
	},
}

var initCmd = &bonzai.Cmd{
	Name:  "init",
	Short: "Initialize new arara.yaml",
	Cmds:  []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("Init command - to be implemented")
		return nil
	},
}


// Cmd defines the root command for arara CLI
var Cmd = &bonzai.Cmd{
	Name:  "arara",
	Alias: "ar",
	Usage: "arara [help|build|install|setup|list|init|create]",
	Vers:  "v0.1.0",
	Short: "Dotfiles management tool",
	Long: `Arara is a CLI tool for managing dotfiles installation and configuration.

Commands:
  build     - Execute or list build steps
  create    - Create new resources (install scripts, build steps)
  install   - Install additional tools
  setup     - Core setup operations (backup, link, restore)
  list      - List available installation scripts
  init      - Initialize new arara.yaml configuration
  help      - Show this help message

Examples:
  arara setup backup                  # Backup existing dotfiles
  arara setup link                    # Create symlinks for dotfiles
  arara build install                 # Execute all build steps
  arara create install docker         # Create a new Docker install script
  arara create build-step docker "Install Docker" "arara install docker"

For more details on a specific command:
  arara <command> help

Source: github.com/BuddhiLW/arara
Issues: github.com/BuddhiLW/arara/issues`,
	Cmds: []*bonzai.Cmd{
		build.Cmd,    // Initial dotfiles setup
		create.Cmd,   // Create resources (install scripts, build steps)
		installCmd,   // Install additional tools
		setup.Cmd,    // Core setup operations
		listCmd,      // List available scripts
		initCmd,      // Initialize new arara.yaml
		help.Cmd,     // Show help
	},
	Comp:  comp.Cmds,
	Def:   help.Cmd,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		fmt.Println("Arara - Dotfiles Management Tool (v0.1.0)")
		fmt.Println("-------------------------------------")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  build     - Execute or list build steps")
		fmt.Println("  create    - Create new resources (install scripts, build steps)")
		fmt.Println("  install   - Install additional tools")
		fmt.Println("  setup     - Core setup operations (backup, link, restore)")
		fmt.Println("  list      - List available installation scripts")
		fmt.Println("  init      - Initialize new arara.yaml configuration")
		fmt.Println("  help      - Show this help message")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  arara setup backup                  # Backup existing dotfiles")
		fmt.Println("  arara setup link                    # Create symlinks for dotfiles")
		fmt.Println("  arara build install                 # Execute all build steps")
		fmt.Println("  arara create install docker         # Create a new Docker install script")
		fmt.Println("  arara create build-step docker \"Install Docker\" \"arara install docker\"")
		fmt.Println()
		fmt.Println("For more details on a specific command:")
		fmt.Println("  arara <command> help")
		fmt.Println()
		fmt.Println("Source: github.com/BuddhiLW/arara")
		fmt.Println("Issues: github.com/BuddhiLW/arara/issues")
		
		return nil
	},
}