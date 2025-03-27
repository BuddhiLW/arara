package app

import (
	"testing"

	"github.com/BuddhiLW/arara/internal/app/build"
	"github.com/BuddhiLW/arara/internal/app/create"
	"github.com/BuddhiLW/arara/internal/app/setup"
)

func TestRootCmd(t *testing.T) {
	// Verify the root command has the expected properties
	if Cmd.Name != "arara" {
		t.Errorf("Expected root command name to be 'arara', got '%s'", Cmd.Name)
	}

	if Cmd.Alias != "ar" {
		t.Errorf("Expected root command alias to be 'ar', got '%s'", Cmd.Alias)
	}

	// Verify the root command has the expected subcommands
	if len(Cmd.Cmds) < 6 {
		t.Errorf("Expected at least 6 subcommands in root command, got %d", len(Cmd.Cmds))
	}

	var hasBuildCmd, hasCreateCmd, hasInstallCmd, hasSetupCmd, hasListCmd, hasInitCmd, hasHelpCmd bool

	for _, cmd := range Cmd.Cmds {
		switch cmd.Name {
		case "build":
			hasBuildCmd = true
			if cmd != build.Cmd {
				t.Errorf("Expected build subcommand to be the build.Cmd instance")
			}
		case "create":
			hasCreateCmd = true
			if cmd != create.Cmd {
				t.Errorf("Expected create subcommand to be the create.Cmd instance")
			}
		case "install":
			hasInstallCmd = true
		case "setup":
			hasSetupCmd = true
			if cmd != setup.Cmd {
				t.Errorf("Expected setup subcommand to be the setup.Cmd instance")
			}
		case "list":
			hasListCmd = true
		case "init":
			hasInitCmd = true
		case "help":
			hasHelpCmd = true
		}
	}

	if !hasBuildCmd {
		t.Errorf("Expected root command to have build subcommand")
	}
	if !hasCreateCmd {
		t.Errorf("Expected root command to have create subcommand")
	}
	if !hasInstallCmd {
		t.Errorf("Expected root command to have install subcommand")
	}
	if !hasSetupCmd {
		t.Errorf("Expected root command to have setup subcommand")
	}
	if !hasListCmd {
		t.Errorf("Expected root command to have list subcommand")
	}
	if !hasInitCmd {
		t.Errorf("Expected root command to have init subcommand")
	}
	if !hasHelpCmd {
		t.Errorf("Expected root command to have help subcommand")
	}
}
