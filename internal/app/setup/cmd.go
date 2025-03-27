package setup

import (
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/BuddhiLW/arara/internal/app/backup"
	"github.com/BuddhiLW/arara/internal/app/link"
)

// Placeholder for restore command
var restoreCmd = &bonzai.Cmd{
	Name:  "restore",
	Alias: "r",
	Short: "restore from backup",
	Cmds:  []*bonzai.Cmd{help.Cmd},
}

var Cmd = &bonzai.Cmd{
	Name:  "setup",
	Alias: "s",
	Short: "core dotfiles setup operations",
	Cmds: []*bonzai.Cmd{
		backup.Cmd,   // Backup existing dotfiles
		link.Cmd,     // Create symlinks
		restoreCmd,   // Restore from backup
		help.Cmd,     // Show help
	},
}
