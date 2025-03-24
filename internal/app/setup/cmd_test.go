package setup

import (
	"testing"

	"github.com/BuddhiLW/arara/internal/app/backup"
	"github.com/BuddhiLW/arara/internal/app/link"
)

func TestSetupCmd(t *testing.T) {
	// Verify that the setup command has the expected subcommands
	if len(Cmd.Cmds) < 3 {
		t.Errorf("Expected at least 3 subcommands in setup command, got %d", len(Cmd.Cmds))
	}
	
	var hasBackupCmd, hasLinkCmd, hasRestoreCmd, hasHelpCmd bool
	
	for _, cmd := range Cmd.Cmds {
		switch cmd.Name {
		case "backup":
			hasBackupCmd = true
			if cmd != backup.Cmd {
				t.Errorf("Expected backup subcommand to be the backup.Cmd instance")
			}
		case "link":
			hasLinkCmd = true
			if cmd != link.Cmd {
				t.Errorf("Expected link subcommand to be the link.Cmd instance")
			}
		case "restore":
			hasRestoreCmd = true
		case "help":
			hasHelpCmd = true
		}
	}
	
	if !hasBackupCmd {
		t.Errorf("Expected setup command to have backup subcommand")
	}
	if !hasLinkCmd {
		t.Errorf("Expected setup command to have link subcommand")
	}
	if !hasRestoreCmd {
		t.Errorf("Expected setup command to have restore subcommand")
	}
	if !hasHelpCmd {
		t.Errorf("Expected setup command to have help subcommand")
	}
}