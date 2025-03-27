package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
)

var Cmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "install additional tools",
	Long: `
	Install additional tools and configurations from the scripts directory.
	Scripts are defined in arara.yaml and executed with proper environment setup.
	`,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		executeCmd,
	},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Get dotfiles path from vars
		dotfilesPath, err := config.GetDotfilesPath()
		if err != nil {
			return fmt.Errorf("failed to get dotfiles path: %w", err)
		}
		if dotfilesPath == "" {
			return fmt.Errorf("no active dotfiles repository found")
		}

		// Load config to get environment variables
		cfg, err := config.LoadConfig(filepath.Join(dotfilesPath, "arara.yaml"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Set environment variables from config into vars.Data
		for k, v := range cfg.Env {
			expanded := os.ExpandEnv(v)
			bonzaiVars.Data.Set(k, expanded)
		}

		// If no args, list available scripts
		if len(args) == 0 {
			fmt.Println("Available installation scripts:")
			for _, script := range cfg.Scripts.Install {
				fmt.Printf("  %s - %s\n", script.Name, script.Description)
			}
			return nil
		}

		// Find script and execute
		scriptName := args[0]
		for _, script := range cfg.Scripts.Install {
			if script.Name == scriptName {
				scriptPath := filepath.Join(dotfilesPath, script.Path)
				return executeCmd.Do(executeCmd, scriptPath)
			}
		}

		return fmt.Errorf("script not found: %s", scriptName)
	},
}

var executeCmd = &bonzai.Cmd{
	Name:    "execute",
	Alias:   "exec",
	Short:   "execute installation script",
	Usage:   "execute <script-path>",
	NumArgs: 1,
	Vars: bonzai.Vars{
		{
			K: "ARARA_SCRIPT",
			V: "",
			E: "ARARA_SCRIPT",
			S: "Path to script being executed",
		},
		{
			K: "ARARA_SCRIPT_DIR",
			V: "",
			E: "ARARA_SCRIPT_DIR",
			S: "Directory containing the script",
		},
	},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		path := args[0]

		// Check if script exists and is executable
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("script not found: %w", err)
		}

		if info.Mode()&0111 == 0 {
			return fmt.Errorf("script is not executable: %s", path)
		}

		// Build environment with vars.Data
		env := os.Environ()
		if data, err := bonzaiVars.Data.All(); err == nil {
			for _, line := range strings.Split(data, "\n") {
				if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
					env = append(env, fmt.Sprintf("%s=%s", parts[0], parts[1]))
				}
			}
		}

		// Execute script
		cmd := exec.Command(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script execution failed: %w", err)
		}

		return nil
	},
}
