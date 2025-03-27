package compat

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// CompatSpec defines the compatibility requirements for a script
type CompatSpec struct {
	OS       string `yaml:"os"`       // Operating system name (e.g., debian, ubuntu, darwin)
	Arch     string `yaml:"arch"`     // Architecture (e.g., amd64, arm64)
	Shell    string `yaml:"shell"`    // Shell (e.g., bash, zsh)
	PkgMgr   string `yaml:"pkgmgr"`   // Package manager (e.g., apt, yum, pacman)
	Kernel   string `yaml:"kernel"`   // Kernel version requirement
	Custom   []any  `yaml:"custom"`   // Custom user-defined validation
}

// ValidatorFunc is a function that performs a specific compatibility check
type ValidatorFunc func(value string) bool

// validatorRegistry stores validator functions for different compatibility fields
var validatorRegistry = struct {
	sync.RWMutex
	validators map[string]ValidatorFunc
}{
	validators: map[string]ValidatorFunc{},
}

// RegisterValidator registers a validator function for a specific compatibility field
func RegisterValidator(field string, validator ValidatorFunc) {
	validatorRegistry.Lock()
	defer validatorRegistry.Unlock()
	validatorRegistry.validators[field] = validator
}

// getValidator returns the validator function for a specific field
func getValidator(field string) (ValidatorFunc, bool) {
	validatorRegistry.RLock()
	defer validatorRegistry.RUnlock()
	validator, ok := validatorRegistry.validators[field]
	return validator, ok
}

// init initializes the default validators
func init() {
	// OS validator
	RegisterValidator("os", func(value string) bool {
		if value == "" {
			return true // No requirement specified
		}

		// Get OS information
		osInfo, err := getOSInfo()
		if err != nil {
			return false
		}

		// Check if the required OS matches
		return strings.EqualFold(osInfo["ID"], value) || 
		       strings.Contains(strings.ToLower(osInfo["ID_LIKE"]), strings.ToLower(value))
	})

	// Architecture validator
	RegisterValidator("arch", func(value string) bool {
		if value == "" {
			return true // No requirement specified
		}
		return strings.EqualFold(runtime.GOARCH, value)
	})

	// Shell validator
	RegisterValidator("shell", func(value string) bool {
		if value == "" {
			return true // No requirement specified
		}

		shell := os.Getenv("SHELL")
		return strings.HasSuffix(shell, value)
	})

	// Package manager validator
	RegisterValidator("pkgmgr", func(value string) bool {
		if value == "" {
			return true // No requirement specified
		}

		// Check if the package manager is in PATH
		_, err := exec.LookPath(value)
		return err == nil
	})

	// Kernel validator
	RegisterValidator("kernel", func(value string) bool {
		if value == "" {
			return true // No requirement specified
		}

		// Get kernel version
		out, err := exec.Command("uname", "-r").Output()
		if err != nil {
			return false
		}

		kernel := strings.TrimSpace(string(out))
		// Simple prefix check - can be enhanced with semver comparison
		return strings.HasPrefix(kernel, value)
	})
}

// Check validates if the current system environment meets the compatibility requirements
func Check(compat CompatSpec) bool {
	// Check OS compatibility
	if validator, ok := getValidator("os"); ok {
		if !validator(compat.OS) {
			return false
		}
	}

	// Check architecture compatibility
	if validator, ok := getValidator("arch"); ok {
		if !validator(compat.Arch) {
			return false
		}
	}

	// Check shell compatibility
	if validator, ok := getValidator("shell"); ok {
		if !validator(compat.Shell) {
			return false
		}
	}

	// Check package manager compatibility
	if validator, ok := getValidator("pkgmgr"); ok {
		if !validator(compat.PkgMgr) {
			return false
		}
	}

	// Check kernel version compatibility
	if validator, ok := getValidator("kernel"); ok {
		if !validator(compat.Kernel) {
			return false
		}
	}

	// Check custom validators
	customReqs := make([]interface{}, 0, len(compat.Custom))
	for _, c := range compat.Custom {
		customReqs = append(customReqs, c)
	}
	if !CheckCustom(customReqs) {
		return false
	}

	// All checks passed
	return true
}

// getOSInfo parses /etc/os-release to get OS information
func getOSInfo() (map[string]string, error) {
	osInfo := make(map[string]string)

	// On macOS, manually set values since /etc/os-release doesn't exist
	if runtime.GOOS == "darwin" {
		osInfo["ID"] = "darwin"
		osInfo["ID_LIKE"] = "darwin macos"
		osInfo["NAME"] = "macOS"
		return osInfo, nil
	}

	// Read /etc/os-release
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	// Parse the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")
		osInfo[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /etc/os-release: %w", err)
	}

	return osInfo, nil
}