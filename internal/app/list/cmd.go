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
		if _, err := os.Stat("arara.yaml"); err == nil {
			return localCmd.Do(caller, args...)
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
