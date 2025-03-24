package create

import (
	"testing"
)

func TestCreateCmd(t *testing.T) {
	// Verify that the create command has the expected subcommands
	if len(Cmd.Cmds) != 3 {
		t.Errorf("Expected 3 subcommands in create command, got %d", len(Cmd.Cmds))
	}
	
	var hasInstallCmd, hasBuildStepCmd, hasHelpCmd bool
	
	for _, cmd := range Cmd.Cmds {
		switch cmd.Name {
		case "install":
			hasInstallCmd = true
		case "build-step":
			hasBuildStepCmd = true
		case "help":
			hasHelpCmd = true
		}
	}
	
	if !hasInstallCmd {
		t.Errorf("Expected create command to have install subcommand")
	}
	if !hasBuildStepCmd {
		t.Errorf("Expected create command to have build-step subcommand")
	}
	if !hasHelpCmd {
		t.Errorf("Expected create command to have help subcommand")
	}
}

func TestInstallBinCmd_Parameters(t *testing.T) {
	// Verify installBinCmd has the correct parameters
	if installBinCmd.Name != "install" {
		t.Errorf("Expected name to be 'install', got '%s'", installBinCmd.Name)
	}
	
	if installBinCmd.Alias != "i" {
		t.Errorf("Expected alias to be 'i', got '%s'", installBinCmd.Alias)
	}
	
	if installBinCmd.MinArgs != 1 {
		t.Errorf("Expected MinArgs to be 1, got %d", installBinCmd.MinArgs)
	}
}

func TestBuildStepCmd_Parameters(t *testing.T) {
	// Verify buildStepCmd has the correct parameters
	if buildStepCmd.Name != "build-step" {
		t.Errorf("Expected name to be 'build-step', got '%s'", buildStepCmd.Name)
	}
	
	if buildStepCmd.Alias != "b" {
		t.Errorf("Expected alias to be 'b', got '%s'", buildStepCmd.Alias)
	}
	
	if buildStepCmd.MinArgs != 2 {
		t.Errorf("Expected MinArgs to be 2, got %d", buildStepCmd.MinArgs)
	}
}

func TestGetIndentation(t *testing.T) {
	// Test getIndentation function
	lines := []string{
		"build:",
		"  steps:",
		"    - name: \"test\"",
		"      description: \"Test step\"",
	}
	
	indent := getIndentation(lines, 1) // From "steps:" line
	expected := "    "
	
	if indent != expected {
		t.Errorf("Expected indentation to be '%s', got '%s'", expected, indent)
	}
	
	// Test with no indentation patterns
	emptyLines := []string{
		"build:",
		"  steps:",
	}
	
	defaultIndent := getIndentation(emptyLines, 1)
	if defaultIndent != "  " {
		t.Errorf("Expected default indentation to be '  ', got '%s'", defaultIndent)
	}
}