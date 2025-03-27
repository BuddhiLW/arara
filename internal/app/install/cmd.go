package install

import (
	"fmt"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

var Cmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "install additional tools",
	Long: `
	Install additional tools and configurations from the scripts directory.
	Scripts are defined in arara.yaml and executed with proper environment setup.
	`,
	Cmds: []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Load config
		cfg, err := config.LoadConfig("arara.yaml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Validate script arguments
		if len(args) == 0 {
			return fmt.Errorf("please specify scripts to install")
		}

		// Execute requested scripts
		for _, scriptName := range args {
			// Find script in config
			var scriptPath string
			for _, script := range cfg.Scripts.Install {
				if script.Name == scriptName {
					scriptPath = script.Path
					break
				}
			}

			if scriptPath == "" {
				return fmt.Errorf("script not found: %s", scriptName)
			}

			// Execute script
			fmt.Printf("Installing %s...\n", scriptName)
			if err := executeScript(scriptPath); err != nil {
				return fmt.Errorf("failed to install %s: %w", scriptName, err)
			}
		}

		return nil
	},
}

func executeScript(path string) error {
	// TODO: Implement script execution with proper environment
	return nil
}
