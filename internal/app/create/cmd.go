package create

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

// Cmd represents the create command
var Cmd = &bonzai.Cmd{
	Name:  "create",
	Alias: "c",
	Short: "create new arara resources",
	Long: `
The create command helps you add new resources to your dotfiles configuration.
It streamlines the creation of installation scripts and build steps.

Subcommands:
  install    - Create a new installation script in scripts/install/
              These scripts can be executed via 'arara install <name>'
  
  build-step - Add a new build step to your arara.yaml configuration
              Build steps are executed during 'arara build install'

Examples:
  arara create install docker
      Creates a new installation script for Docker and opens it in your editor
  
  arara create build-step docker "Install Docker" "arara install docker"
      Adds a new build step to arara.yaml to install Docker during builds

The create command helps maintain a consistent structure for your
dotfiles repository and simplifies adding new functionality.
`,
	Cmds:  []*bonzai.Cmd{help.Cmd, installBinCmd, buildStepCmd},
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