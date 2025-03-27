package namespace

import (
	"fmt"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
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
