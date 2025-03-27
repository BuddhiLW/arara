package compat

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
)

// Cmd represents the compatibility command
var Cmd = &bonzai.Cmd{
	Name:  "compat",
	Alias: "c",
	Short: "check system compatibility for scripts",
	Long: `
The compat command provides tools to check if scripts are compatible
with the current system environment.

The compatibility system allows scripts to specify requirements such as:
- Operating system (os)
- Architecture (arch) 
- Shell (shell)
- Package manager (pkgmgr)
- Kernel version (kernel)
- Custom validators

# Configuration

These requirements are specified in the arara.yaml configuration:

  scripts:
    install:
      - name: docker
        path: scripts/install/docker
        compat:
          os: debian
          shell: bash
          pkgmgr: apt
          custom:
            - name: min-memory
              value: 4096

# Subcommands

  check - Check compatibility of a specific script
  list  - List all validators available in the system

# Examples

  arara compat check docker       # Check if docker script is compatible
  arara compat list               # List all available validators
  arara compat list --details     # List validators with detailed descriptions

# Extension Points

Custom validators can be registered by implementing the CustomValidator interface
and registering it with the RegisterCustomValidator function.
	`,
	Cmds: []*bonzai.Cmd{
		checkCmd,
		listCmd,
		help.Cmd, // Include the Bonzai help command
	},
	Def: help.Cmd, // Default to showing help when no subcommand is specified
}

// checkCmd represents the check subcommand
var checkCmd = &bonzai.Cmd{
	Name:  "check",
	Alias: "c",
	Short: "check compatibility of a script",
	Long: `
The check subcommand evaluates if a specified script is compatible
with the current system environment based on its compatibility
requirements defined in arara.yaml.

# Usage
  arara compat check <script-name>

# Arguments
  <script-name> - Name of the script to check compatibility for

# Examples
  arara compat check docker   # Check if docker script is compatible
  arara compat check xmonad   # Check if xmonad setup is compatible

# Output
The command returns:
- Success (exit code 0) if the script is compatible
- Failure (exit code 1) if the script is not compatible

It will also print a detailed report of which compatibility
checks passed or failed.
	`,
	MinArgs: 1,
	MaxArgs: 1,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		scriptName := args[0]
		fmt.Printf("Checking compatibility for script: %s\n\n", scriptName)
		
		// TODO: Implement actual configuration parsing
		// This is a simplified version for now
		
		fmt.Println("System information:")
		fmt.Println("------------------")
		
		// Check OS
		osInfo, err := getOSInfo()
		if err != nil {
			fmt.Printf("OS: Unknown (error: %v)\n", err)
		} else {
			fmt.Printf("OS: %s\n", osInfo["ID"])
		}
		
		// Check Architecture
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
		
		// Check Shell
		shell := os.Getenv("SHELL")
		if shell == "" {
			fmt.Println("Shell: Unknown")
		} else {
			fmt.Printf("Shell: %s\n", filepath.Base(shell))
		}
		
		// Check package managers
		pkgManagers := []string{"apt", "yum", "pacman", "brew"}
		availablePkgMgrs := []string{}
		
		for _, pm := range pkgManagers {
			_, err := exec.LookPath(pm)
			if err == nil {
				availablePkgMgrs = append(availablePkgMgrs, pm)
			}
		}
		
		if len(availablePkgMgrs) == 0 {
			fmt.Println("Package Managers: None detected")
		} else {
			fmt.Printf("Package Managers: %s\n", strings.Join(availablePkgMgrs, ", "))
		}
		
		fmt.Println("\nNotice: This is a simplified compatibility check.")
		fmt.Println("A full implementation will read requirements from arara.yaml")
		
		return nil
	},
}

// listCmd represents the list subcommand
var listCmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "ls",
	Short: "list available compatibility validators",
	Long: `
The list subcommand displays all available compatibility validators
registered in the system, including both built-in and custom validators.

# Usage
  arara compat list [--details]

# Options
  --details  Show detailed information about each validator

# Examples
  arara compat list           # List all validators
  arara compat list --details # List with detailed descriptions

# Built-in Validators
  os     : Operating system (debian, ubuntu, darwin, etc.)
  arch   : CPU architecture (amd64, arm64, etc.)
  shell  : Current shell (bash, zsh, etc.)
  pkgmgr : Package manager (apt, yum, pacman, etc.)
  kernel : Kernel version

# Custom Validators
Custom validators are shown with their registered name and
can be used in the 'custom' section of compatibility requirements.
	`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		showDetails := false
		
		// Check if --details flag is provided
		for _, arg := range args {
			if arg == "--details" {
				showDetails = true
				break
			}
		}
		
		fmt.Println("Available compatibility validators:")
		fmt.Println("----------------------------------")
		
		// Built-in validators
		fmt.Println("\nBuilt-in validators:")
		fmt.Println("  os      - Operating system (e.g., debian, ubuntu, darwin)")
		fmt.Println("  arch    - CPU architecture (e.g., amd64, arm64)")
		fmt.Println("  shell   - Current shell (e.g., bash, zsh)")
		fmt.Println("  pkgmgr  - Package manager (e.g., apt, yum, pacman)")
		fmt.Println("  kernel  - Kernel version")
		
		if showDetails {
			fmt.Println("\nDetails:")
			fmt.Println("  os:")
			fmt.Println("    Checks if the current OS matches the required OS.")
			fmt.Println("    Uses /etc/os-release on Linux and runtime.GOOS on other platforms.")
			fmt.Println("    Supports 'ID' and 'ID_LIKE' from os-release for broader matching.")
			fmt.Println("\n  arch:")
			fmt.Println("    Checks if the current architecture matches the required architecture.")
			fmt.Println("    Uses Go's runtime.GOARCH value.")
			fmt.Println("\n  shell:")
			fmt.Println("    Checks if the current shell matches the required shell.")
			fmt.Println("    Uses the SHELL environment variable.")
			fmt.Println("\n  pkgmgr:")
			fmt.Println("    Checks if the specified package manager is available in the PATH.")
			fmt.Println("\n  kernel:")
			fmt.Println("    Checks if the kernel version matches or exceeds the required version.")
			fmt.Println("    Uses 'uname -r' command output.")
		}
		
		// Custom validators
		customRegistry.RLock()
		customValidators := make([]string, 0, len(customRegistry.validators))
		for name := range customRegistry.validators {
			customValidators = append(customValidators, name)
		}
		customRegistry.RUnlock()
		
		if len(customValidators) > 0 {
			fmt.Println("\nCustom validators:")
			for _, name := range customValidators {
				fmt.Printf("  %s\n", name)
			}
			
			if showDetails && len(customValidators) > 0 {
				fmt.Println("\nCustom validator details:")
				fmt.Println("  To see documentation for custom validators, refer to their")
				fmt.Println("  respective package documentation or implementation.")
			}
		} else {
			fmt.Println("\nNo custom validators registered.")
		}
		
		fmt.Println("\nUsage in arara.yaml:")
		fmt.Println("  compat:")
		fmt.Println("    os: debian")
		fmt.Println("    shell: bash")
		fmt.Println("    custom:")
		fmt.Println("      - name: min-memory")
		fmt.Println("        value: 4096")
		
		return nil
	},
}