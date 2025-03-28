package deps

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
)

// Add to package-level vars for testing
var (
	Stdin  io.Reader = os.Stdin  // For mocking in tests
	Stdout io.Writer = os.Stdout // For capturing output
)

type PackageManager struct {
	Name          string
	InstallCmd    string
	InstallPrefix []string
	YesFlag       string // Flag to auto-accept prompts
}

var packageManagers = map[string]PackageManager{
	"apt": {
		Name:          "apt",
		InstallCmd:    "install",
		InstallPrefix: []string{"sudo", "apt-get"},
		YesFlag:       "-y -qq",
	},
	"dnf": {
		Name:          "dnf",
		InstallCmd:    "install",
		InstallPrefix: []string{"sudo", "dnf"},
		YesFlag:       "-y",
	},
	"yum": {
		Name:          "yum",
		InstallCmd:    "install",
		InstallPrefix: []string{"sudo", "yum"},
		YesFlag:       "-y",
	},
	"pacman": {
		Name:          "pacman",
		InstallCmd:    "-S",
		InstallPrefix: []string{"sudo", "pacman"},
		YesFlag:       "--noconfirm",
	},
	"brew": {
		Name:          "brew",
		InstallCmd:    "install",
		InstallPrefix: []string{"brew"},
		YesFlag:       "", // Homebrew doesn't prompt by default
	},
}

var Cmd = &bonzai.Cmd{
	Name:  "deps",
	Alias: "dependencies",
	Short: "manage system dependencies",
	Long: `
Manage system dependencies for your dotfiles setup.

This command allows you to:
- Sync dependencies from a file (with #-commented-deps format)
- List current dependencies
- Add new dependencies
- Remove dependencies
- Install dependencies using system package manager

Dependencies are stored in the active namespace's arara.yaml configuration file.
`,
	Cmds: []*bonzai.Cmd{syncCmd, listCmd, addCmd, removeCmd, installCmd, help.Cmd},
}

var syncCmd = &bonzai.Cmd{
	Name:  "sync",
	Alias: "s",
	Usage: "sync [file]",
	Short: "sync dependencies from a file",
	Long: `
Sync dependencies from a file with #-commented-deps format to your arara.yaml configuration.

This will read the specified file and extract all non-commented lines as dependencies.
Lines starting with # or ## are treated as comments and ignored.

Usage:
  arara deps sync path/to/deps-file.txt
`,
	MinArgs: 1,
	MaxArgs: 1,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Get the file path
		filePath := args[0]
		if !filepath.IsAbs(filePath) {
			// If relative path, make it absolute
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}

		// Read the dependencies file
		deps, err := readDepsFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read dependencies file: %w", err)
		}

		// Update the configuration
		return saveDependenciesToConfig(deps)
	},
}

var listCmd = &bonzai.Cmd{
	Name:  "list",
	Alias: "ls",
	Short: "list dependencies",
	Long: `
List all dependencies stored in the active namespace's arara.yaml configuration.

This will display all dependencies that are currently defined.
`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		deps, err := loadDependencies()
		if err != nil {
			return err
		}

		if len(deps) == 0 {
			fmt.Println("No dependencies found")
			return nil
		}

		fmt.Println("Dependencies:")
		for _, dep := range deps {
			fmt.Printf("  %s\n", dep)
		}
		return nil
	},
}

var addCmd = &bonzai.Cmd{
	Name:  "add",
	Alias: "a",
	Short: "add dependencies",
	Usage: "add <package1> [package2...]",
	Long: `
Add one or more dependencies to the active namespace's arara.yaml configuration.

This will add the specified packages to the dependencies list if they don't already exist.

Usage:
  arara deps add git vim tmux
`,
	MinArgs: 1,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Load current dependencies
		currentDeps, err := loadDependencies()
		if err != nil {
			return err
		}

		// Create a map for fast lookups
		depsMap := make(map[string]bool)
		for _, dep := range currentDeps {
			depsMap[dep] = true
		}

		// Add new dependencies
		added := 0
		var newDeps []string
		for _, arg := range args {
			// Split in case an argument contains multiple packages
			for _, dep := range strings.Fields(arg) {
				if !depsMap[dep] {
					newDeps = append(newDeps, dep)
					depsMap[dep] = true
					added++
				}
			}
		}

		if added == 0 {
			fmt.Println("All packages already in dependencies list")
			return nil
		}

		// Save updated dependencies (add only new ones to current list)
		allDeps := append(currentDeps, newDeps...)
		if err := saveDependenciesToConfig(allDeps); err != nil {
			return err
		}

		fmt.Printf("Added %d dependencies\n", added)
		return nil
	},
}

var removeCmd = &bonzai.Cmd{
	Name:  "remove",
	Alias: "rm",
	Short: "remove dependencies",
	Usage: "remove <package1> [package2...]",
	Long: `
Remove one or more dependencies from the active namespace's arara.yaml configuration.

This will remove the specified packages from the dependencies list if they exist.

Usage:
  arara deps remove git vim tmux
`,
	MinArgs: 1,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		// Load current dependencies
		currentDeps, err := loadDependencies()
		if err != nil {
			return err
		}

		// Create a map of dependencies to remove
		toRemove := make(map[string]bool)
		for _, arg := range args {
			// Split in case an argument contains multiple packages
			for _, dep := range strings.Fields(arg) {
				toRemove[dep] = true
			}
		}

		// Filter out the dependencies to remove
		var newDeps []string
		removed := 0
		for _, dep := range currentDeps {
			if !toRemove[dep] {
				newDeps = append(newDeps, dep)
			} else {
				removed++
			}
		}

		if removed == 0 {
			fmt.Println("None of the specified dependencies were found")
			return nil
		}

		// Save updated dependencies
		if err := saveDependenciesToConfig(newDeps); err != nil {
			return err
		}

		fmt.Printf("Removed %d dependencies\n", removed)
		return nil
	},
}

var installCmd = &bonzai.Cmd{
	Name:  "install",
	Alias: "i",
	Short: "install dependencies",
	Usage: "install [package1 package2...]",
	Long: `
Install dependencies using the system's package manager.

If no arguments are provided, this will install all dependencies listed in the 
active namespace's arara.yaml configuration.

If specific packages are provided as arguments, only those packages will be installed.

Supported package managers:
  - apt (Debian, Ubuntu)
  - dnf (Fedora)
  - yum (CentOS, RHEL)
  - pacman (Arch Linux)
  - brew (macOS)

Usage:
  arara deps install           # Install all dependencies from config
  arara deps install git tmux  # Install specific packages
`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		var deps []string
		var err error

		// If no args provided, load all dependencies from config
		if len(args) == 0 {
			deps, err = loadDependencies()
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				fmt.Println("No dependencies found to install")
				return nil
			}
		} else {
			// Process args to get individual packages
			for _, arg := range args {
				// Split in case an argument contains multiple packages
				for _, dep := range strings.Fields(arg) {
					if dep != "" {
						deps = append(deps, dep)
					}
				}
			}
		}

		// Detect package manager
		pm, err := detectPackageManager()
		if err != nil {
			return err
		}

		// Run install command
		fmt.Printf("Installing %d dependencies using %s...\n", len(deps), pm.Name)

		// Build command with the yes flag if available
		cmdArgs := append(pm.InstallPrefix, pm.InstallCmd)

		// Add yes flag if provided
		if pm.YesFlag != "" {
			cmdArgs = append(cmdArgs, pm.YesFlag)
		}

		// Add all dependencies
		cmdArgs = append(cmdArgs, deps...)

		fmt.Printf("Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	},
}

// readDepsFile reads a dependency file and returns a list of dependencies
// ignoring lines that start with # or ##
// Each dependency should be on its own line or separated by spaces
func readDepsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle multiple dependencies on one line by splitting on whitespace
		lineDeps := strings.Fields(line)
		for _, dep := range lineDeps {
			// Ignore any YAML formatting or list markers
			dep = strings.TrimPrefix(dep, "-")
			dep = strings.TrimSpace(dep)
			if dep != "" {
				deps = append(deps, dep)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

// loadDependencies loads the dependencies from the active namespace's arara.yaml
func loadDependencies() ([]string, error) {
	// Get the active namespace
	activeNS := bonzaiVars.Fetch("ARARA_ACTIVE_NAMESPACE", "active-namespace", "")
	if activeNS == "" {
		return nil, fmt.Errorf("no active namespace set. Use 'arara namespace switch <n>' first")
	}

	// Get dotfiles path from active namespace
	dotfilesPath, err := config.GetDotfilesPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get dotfiles path: %w", err)
	}
	if dotfilesPath == "" {
		return nil, fmt.Errorf("no dotfiles path found for namespace: %s", activeNS)
	}

	// Load the configuration
	cfg, err := config.LoadConfig(filepath.Join(dotfilesPath, "arara.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config for namespace %s: %w", activeNS, err)
	}

	// Process dependencies to ensure each entry is a single package
	var flatDeps []string
	for _, dep := range cfg.Dependencies {
		// Split any multi-word dependencies into individual packages
		for _, singleDep := range strings.Fields(dep) {
			if singleDep != "" {
				flatDeps = append(flatDeps, singleDep)
			}
		}
	}

	return flatDeps, nil
}

// saveDependenciesToConfig saves dependencies to the active namespace's arara.yaml
func saveDependenciesToConfig(deps []string) error {
	// Get the active namespace
	activeNS := bonzaiVars.Fetch("ARARA_ACTIVE_NAMESPACE", "active-namespace", "")
	if activeNS == "" {
		return fmt.Errorf("no active namespace set. Use 'arara namespace switch <n>' first")
	}

	// Get dotfiles path from active namespace
	dotfilesPath, err := config.GetDotfilesPath()
	if err != nil {
		return fmt.Errorf("failed to get dotfiles path: %w", err)
	}
	if dotfilesPath == "" {
		return fmt.Errorf("no dotfiles path found for namespace: %s", activeNS)
	}

	configPath := filepath.Join(dotfilesPath, "arara.yaml")

	// Begin transaction for atomic updates
	tx, err := beginTransaction(configPath)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Roll back on failure
	defer func() {
		if err != nil {
			tx.rollback()
		}
	}()

	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update dependencies
	cfg.Dependencies = deps

	// Check for concurrent modifications
	if modified, err := tx.checkModified(); err != nil {
		return fmt.Errorf("failed to check for concurrent modification: %w", err)
	} else if modified {
		return fmt.Errorf("config was modified during update, please try again")
	}

	// Save updated config
	data, err := cfg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Commit transaction
	if err := tx.commit(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	fmt.Printf("Updated dependencies in %s\n", configPath)
	return nil
}

// detectPackageManager detects which package manager is available on the system
func detectPackageManager() (PackageManager, error) {
	// On macOS, prefer brew
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			return packageManagers["brew"], nil
		}
	}

	// Check for Linux package managers
	if runtime.GOOS == "linux" {
		// Check for pacman (Arch)
		if _, err := exec.LookPath("pacman"); err == nil {
			return packageManagers["pacman"], nil
		}

		// Check for apt (Debian/Ubuntu)
		if _, err := exec.LookPath("apt-get"); err == nil {
			return packageManagers["apt"], nil
		}

		// Check for dnf (Fedora)
		if _, err := exec.LookPath("dnf"); err == nil {
			return packageManagers["dnf"], nil
		}

		// Check for yum (CentOS/RHEL)
		if _, err := exec.LookPath("yum"); err == nil {
			return packageManagers["yum"], nil
		}
	}

	return PackageManager{}, fmt.Errorf("no supported package manager found")
}

// transaction handles atomic updates to arara.yaml
type transaction struct {
	configPath string
	backupPath string
	origHash   []byte
}

// beginTransaction starts a new transaction by creating a backup
func beginTransaction(configPath string) (*transaction, error) {
	// Calculate original file hash
	hash, err := fileHash(configPath)
	if err != nil {
		return nil, err
	}

	// Create backup
	backupPath := configPath + fmt.Sprintf(".bak.%d", os.Getpid())
	if err := copyFile(configPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	return &transaction{
		configPath: configPath,
		backupPath: backupPath,
		origHash:   hash,
	}, nil
}

// commit finalizes the transaction
func (t *transaction) commit() error {
	// Remove backup file
	return os.Remove(t.backupPath)
}

// rollback restores from backup
func (t *transaction) rollback() error {
	return os.Rename(t.backupPath, t.configPath)
}

// checkModified verifies if file was modified during transaction
func (t *transaction) checkModified() (bool, error) {
	currentHash, err := fileHash(t.configPath)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(t.origHash, currentHash), nil
}

// fileHash calculates the SHA-256 hash of a file
func fileHash(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer f.Close()

	h := bufio.NewReader(f)
	buf := make([]byte, 1024)
	hasher := exec.Command("sha256sum")
	stdin, err := hasher.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()

		for {
			n, err := h.Read(buf)
			if n > 0 {
				stdin.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	out, err := hasher.Output()
	if err != nil {
		return nil, err
	}

	// Extract just the hash part
	hash := strings.Split(string(out), " ")[0]
	return []byte(hash), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = bufio.NewReader(source).WriteTo(destination)
	return err
}
