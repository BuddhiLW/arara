package create

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/edit"
)

// Cmd represents the create command
var Cmd = &bonzai.Cmd{
	Name:  "create",
	Alias: "c",
	Short: "create new resources",
	Long: `
Create new resources in the active namespace.

Currently supports:
* install - Create a new install script
* bin - Create a new executable in local-bin
* setup-path - Add local-bin to PATH
`,
	Cmds: []*bonzai.Cmd{help.Cmd, installBinCmd, buildStepCmd, binCmd, setupPathCmd},
}

// installBinCmd creates a new install script
var installBinCmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "create a new installation script",
	Usage: "arara create install <script_name>",
	Long: `
The 'install' subcommand creates a new installation script in your dotfiles
repository, specifically in the 'scripts/install/' directory. These scripts
can later be run using 'arara install <script_name>'.

How it works:
1. Creates the script in $DOTFILES/scripts/install/ (defaults to ~/dotfiles/scripts/install/)
2. Initializes it with a shebang line and basic header comment
3. Makes the script executable (chmod +x)
4. Opens the script in your preferred editor ($EDITOR, defaults to vim)

If the script already exists, it will simply be opened for editing without
overwriting the existing content.

Arguments:
  <script_name>  - Name of the installation script to create (e.g., docker, emacs, etc.)
                   This will also be the name used to run it via 'arara install <script_name>'

Environment Variables:
  DOTFILES       - Path to your dotfiles repository (defaults to ~/dotfiles)
  EDITOR         - Your preferred text editor (defaults to vim)

Examples:
  arara create install docker    # Create scripts/install/docker
  arara create install doom      # Create scripts/install/doom for Doom Emacs
  arara create install rust      # Create scripts/install/rust for Rust language
`,
	MinArgs: 1,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		scriptName := args[0]

		// Get the DOTFILES environment variable
		dotfilesDir := os.Getenv("DOTFILES")
		if dotfilesDir == "" {
			// Default to $HOME/dotfiles if not set
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			dotfilesDir = filepath.Join(homeDir, "dotfiles")
		}

		// Create the scripts/install directory if it doesn't exist
		installDir := filepath.Join(dotfilesDir, "scripts", "install")
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fmt.Errorf("failed to create install scripts directory: %w", err)
		}

		// Full path to the script
		scriptPath := filepath.Join(installDir, scriptName)

		// Check if script already exists
		_, err := os.Stat(scriptPath)
		if os.IsNotExist(err) {
			// Create new script with shebang
			f, err := os.Create(scriptPath)
			if err != nil {
				return fmt.Errorf("failed to create script file: %w", err)
			}

			// Write shebang
			_, err = f.WriteString("#!/usr/bin/bash\n\n# Installation script for " + scriptName + "\n\n")
			if err != nil {
				f.Close()
				return fmt.Errorf("failed to write to script file: %w", err)
			}

			f.Close()

			// Make script executable
			if err := os.Chmod(scriptPath, 0755); err != nil {
				return fmt.Errorf("failed to make script executable: %w", err)
			}

			fmt.Printf("Created new installation script: %s\n", scriptPath)
		} else if err != nil {
			return fmt.Errorf("failed to check if script exists: %w", err)
		} else {
			fmt.Printf("Opening existing script: %s\n", scriptPath)
		}

		// Open the script in editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim" // Default to vim if EDITOR is not set
		}

		cmd := exec.Command(editor, scriptPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	},
}

// buildStepCmd adds a new build step to the arara.yaml file
var buildStepCmd = &bonzai.Cmd{
	Name:  "build-step",
	Alias: "b",
	Short: "add a new build step to arara.yaml",
	Usage: "arara create build-step <step_name> <description> [command]",
	Long: `
The 'build-step' subcommand adds a new step to the build configuration in your
arara.yaml file. These build steps are executed sequentially when running
'arara build install', allowing you to automate your dotfiles setup.

How it works:
1. Locates your arara.yaml configuration file
   (searches current directory, home directory, and $DOTFILES)
2. Identifies the 'steps:' section in the YAML
3. Adds your new step at the end of the section, maintaining proper YAML formatting
4. Preserves existing content and indentation style

The command will automatically detect the YAML structure and indentation
pattern to ensure consistent formatting.

Arguments:
  <step_name>    - The name of the build step (e.g., "docker", "emacs", etc.)
  <description>  - A brief description of what the step does
  [command]      - Optional command to execute (can be added later in the YAML)

Examples:
  # Add a step with a single command
  arara create build-step docker "Install Docker" "arara install docker"

  # Add a step without a command (you can add commands in the YAML later)
  arara create build-step emacs "Setup Emacs configuration"

The YAML output will look like:
  steps:
    - name: "docker"
      description: "Install Docker"
      command: "arara install docker"

If you need multiple commands for a single step, edit the YAML directly
and use the 'commands' field instead of 'command'.
`,
	MinArgs: 2,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		stepName := args[0]
		description := args[1]

		// Get optional command if provided
		var command string
		if len(args) > 2 {
			command = args[2]
		}

		// Look for arara.yaml in current directory or parent directories
		configPath, err := findConfigFile()
		if err != nil {
			return err
		}

		// Read the existing config file
		content, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse the file to find where to insert the new step
		lines := strings.Split(string(content), "\n")
		buildStepsIndex := -1

		for i, line := range lines {
			if strings.TrimSpace(line) == "steps:" {
				buildStepsIndex = i
				break
			}
		}

		if buildStepsIndex == -1 {
			return fmt.Errorf("couldn't find 'steps:' section in the config file")
		}

		// Format the new build step
		var newStep []string
		indent := getIndentation(lines, buildStepsIndex)

		newStep = append(newStep, indent+"- name: \""+stepName+"\"")
		newStep = append(newStep, indent+"  description: \""+description+"\"")

		if command != "" {
			newStep = append(newStep, indent+"  command: \""+command+"\"")
		}

		// Insert the new step after the last existing step
		lastStepIndex := buildStepsIndex
		for i := buildStepsIndex + 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "- name:") {
				lastStepIndex = i
				// Skip to the end of this step
				for j := i + 1; j < len(lines); j++ {
					if j+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j+1]), "- name:") {
						break
					}
					lastStepIndex = j
				}
			}
		}

		// Insert the new step into the lines slice
		updatedLines := make([]string, 0)
		updatedLines = append(updatedLines, lines[:lastStepIndex+1]...)
		updatedLines = append(updatedLines, "")
		updatedLines = append(updatedLines, newStep...)
		if lastStepIndex+1 < len(lines) {
			updatedLines = append(updatedLines, lines[lastStepIndex+1:]...)
		}

		// Write the updated content back to the file
		if err := os.WriteFile(configPath, []byte(strings.Join(updatedLines, "\n")), 0644); err != nil {
			return fmt.Errorf("failed to write updated config file: %w", err)
		}

		fmt.Printf("Added new build step '%s' to %s\n", stepName, configPath)
		return nil
	},
}

// binCmd creates a new executable in local-bin
var binCmd = &bonzai.Cmd{
	Name:  "bin",
	Usage: "<name>",
	Short: "create a new executable in local-bin",
	Long: `
Create a new executable script in the namespace's local-bin directory.
The script will be created in <local-bin>/<name> and made executable.
`,
	Do: createBinScript,
}

// setupPathCmd adds local-bin to PATH in .bashrc
var setupPathCmd = &bonzai.Cmd{
	Name:  "setup-path",
	Short: "add local-bin to PATH",
	Long: `
Add the local-bin directory to PATH in your .bashrc file.
This ensures executables created with 'arara create bin' are available in your shell.
`,
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Get all namespace local-bin paths
		gc, err := config.NewGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		// Collect all local-bin paths
		localBinPaths := make([]string, 0)
		basePath := filepath.Join(home, ".local", "bin")
		for _, ns := range gc.Namespaces {
			if info, ok := gc.Configs[ns]; ok && info.LocalBin != "" {
				localBinPaths = append(localBinPaths, filepath.Join(basePath, info.LocalBin))
			}
		}

		// Read .bashrc
		bashrcPath := filepath.Join(home, ".bashrc")
		content, err := os.ReadFile(bashrcPath)
		if err != nil {
			return fmt.Errorf("failed to read .bashrc: %w", err)
		}

		// Check if paths are already in .bashrc
		if strings.Contains(string(content), "# <<<< arara local-bin setup") {
			// Extract existing paths between markers
			start := strings.Index(string(content), "# <<<< arara local-bin setup")
			end := strings.Index(string(content), "# >>>>")
			if start == -1 || end == -1 {
				return fmt.Errorf("malformed arara path setup in .bashrc")
			}

			existingBlock := string(content[start:end])

			// Check which paths need to be added
			newPaths := make([]string, 0)
			for _, path := range localBinPaths {
				if !strings.Contains(existingBlock, path) {
					newPaths = append(newPaths, path)
				}
			}

			if len(newPaths) == 0 {
				fmt.Println("All paths already configured in .bashrc")
				return nil
			}

			// Create updated content
			beforeBlock := string(content[:start])
			afterBlock := string(content[end+6:]) // +4 to skip "# >>>>

			newContent := "# <<<< arara local-bin setup\n"
			// Keep existing paths
			for _, line := range strings.Split(existingBlock, "\n") {
				if strings.Contains(line, "export PATH") {
					newContent += line + "\n"
				}
			}
			// Add new paths
			for _, path := range newPaths {
				newContent += fmt.Sprintf(`export PATH="$PATH:%s"`+"\n", path)
			}
			newContent += "# >>>>\n"

			// Write updated content
			finalContent := beforeBlock + newContent + afterBlock
			if err := os.WriteFile(bashrcPath, []byte(finalContent), 0644); err != nil {
				return fmt.Errorf("failed to update .bashrc: %w", err)
			}

			fmt.Printf("Added new paths to PATH in .bashrc\n")
			fmt.Println("Please run 'source ~/.bashrc' or start a new shell for changes to take effect")
		} else {
			// No existing block, create new one
			newContent := "\n# <<<< arara local-bin setup\n"
			for _, path := range localBinPaths {
				newContent += fmt.Sprintf(`export PATH="$PATH:%s"`+"\n", path)
			}
			newContent += "# >>>>\n"

			// Append to .bashrc
			if err := os.WriteFile(bashrcPath, append(content, []byte(newContent)...), 0644); err != nil {
				return fmt.Errorf("failed to update .bashrc: %w", err)
			}

			fmt.Printf("Added paths to PATH in .bashrc\n")
			fmt.Println("Please run 'source ~/.bashrc' or start a new shell for changes to take effect")
		}

		return nil
	},
}

// Helper function to find the arara.yaml config file
func findConfigFile() (string, error) {
	// First check current directory
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for arara.yaml or .arara.yaml
	possibleNames := []string{"arara.yaml", ".arara.yaml"}

	// Look in current directory
	for _, name := range possibleNames {
		path := filepath.Join(pwd, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	for _, name := range possibleNames {
		path := filepath.Join(homeDir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try dotfiles directory if it exists
	dotfilesDir := os.Getenv("DOTFILES")
	if dotfilesDir != "" {
		for _, name := range possibleNames {
			path := filepath.Join(dotfilesDir, name)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("couldn't find arara.yaml configuration file")
}

// Helper function to determine indentation in the YAML file
func getIndentation(lines []string, lineIndex int) string {
	// Look at the indentation of child items to match it
	for i := lineIndex + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			// Found a step, extract the indentation
			indent := ""
			for _, char := range line {
				if char == ' ' {
					indent += " "
				} else {
					break
				}
			}
			return indent
		}
	}
	// Default indent if we can't determine it
	return "  "
}

func createBinScript(cmd *bonzai.Cmd, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected script name argument")
	}
	name := args[0]

	// Get active namespace config
	gc, err := config.NewGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	ns := gc.GetActiveNamespace()
	if ns == nil {
		return fmt.Errorf("no active namespace")
	}

	if ns.LocalBin == "" {
		return fmt.Errorf("local-bin not configured for namespace %s", ns.Name)
	}

	// Create bin directory if it doesn't exist
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	localBinDir := filepath.Join(configDir, "..", ".local", "bin")
	binDir := filepath.Join(localBinDir, ns.LocalBin)

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Create script file
	scriptPath := filepath.Join(binDir, name)
	if _, err := os.Stat(scriptPath); err == nil {
		return fmt.Errorf("script %s already exists", name)
	}

	content := []byte(`#!/bin/bash

# Description: <add description>

set -euo pipefail

# Your code here
`)

	if err := os.WriteFile(scriptPath, content, 0755); err != nil {
		return fmt.Errorf("failed to create script: %w", err)
	}

	fmt.Printf("Created executable %s\n", scriptPath)

	// Open the script in editor
	if err := edit.Files(scriptPath); err != nil {
		return fmt.Errorf("failed to open script in editor: %w", err)
	}

	return nil
}
