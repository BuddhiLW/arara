package create

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/arara/internal/pkg/config"
)

func TestCreateCmd(t *testing.T) {
	// Verify that the create command has the expected subcommands
	if len(Cmd.Cmds) != 4 {
		t.Errorf("Expected 4 subcommands in create command, got %d", len(Cmd.Cmds))
	}

	var hasInstallCmd, hasBuildStepCmd, hasHelpCmd, hasBinCmd bool

	for _, cmd := range Cmd.Cmds {
		switch cmd.Name {
		case "install":
			hasInstallCmd = true
		case "build-step":
			hasBuildStepCmd = true
		case "help":
			hasHelpCmd = true
		case "bin":
			hasBinCmd = true
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
	if !hasBinCmd {
		t.Errorf("Expected create command to have bin subcommand")
	}

	t.Run("create bin script", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Setup mock global config
		gc := &config.GlobalConfig{
			Config: config.Config{
				Namespaces: []string{"test"},
				Configs: map[string]config.NSInfo{
					"test": {
						Path:     tmpDir,
						LocalBin: "bin",
					},
				},
			},
		}

		// Mock global config
		oldNewGlobalConfig := config.NewGlobalConfig
		config.NewGlobalConfig = func() (*config.GlobalConfig, error) {
			return gc, nil
		}
		defer func() {
			config.NewGlobalConfig = oldNewGlobalConfig
		}()

		// Set active namespace
		os.Setenv("ARARA_ACTIVE_NAMESPACE", "test")
		os.Setenv("TEST_MODE", "1")

		// Create bin directory
		binDir := filepath.Join(tmpDir, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create bin script by calling the binCmd directly
		if err := binCmd.Do(binCmd, "test-script"); err != nil {
			t.Fatal(err)
		}

		// Verify script was created
		scriptPath := filepath.Join(tmpDir, "bin", "test-script")
		if _, err := os.Stat(scriptPath); err != nil {
			t.Errorf("script not created: %v", err)
		}

		// Verify script is executable
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&0111 == 0 {
			t.Error("script not executable")
		}

		// Clean up
		os.Unsetenv("ARARA_ACTIVE_NAMESPACE")
		os.Unsetenv("TEST_MODE")
	})
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
