package list

import (
	"fmt"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

var Cmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "l",
	Short: "List available scripts",
	Long: `
    List all available installation scripts defined in arara.yaml.
    Shows script name, description, and path.
    `,
	Cmds: []*bonzai.Cmd{help.Cmd},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Load config
		cfg, err := config.LoadConfig("arara.yaml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Display available scripts
		fmt.Println("Available installation scripts:")
		for _, script := range cfg.Scripts.Install {
			fmt.Printf("  %s: %s\n    Path: %s\n\n",
				script.Name,
				script.Description,
				script.Path)
		}

		return nil
	},
}
