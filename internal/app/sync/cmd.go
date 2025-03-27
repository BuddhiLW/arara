package sync

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"

	"bufio"
	"bytes"
	"strconv"

	"github.com/BuddhiLW/arara/internal/pkg/config"
)

// Add to package-level vars for testing
var (
	Stdin  io.Reader = os.Stdin  // For mocking in tests
	Stdout io.Writer = os.Stdout // For capturing output
)

// transaction handles atomic updates to arara.yaml
type transaction struct {
	configPath string
	backupPath string
	origHash   []byte
}

// begin starts a new transaction by creating a backup
func beginTransaction(configPath string) (*transaction, error) {
	// Calculate original file hash
	origFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer origFile.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, origFile); err != nil {
		return nil, fmt.Errorf("failed to hash config: %w", err)
	}

	// Create backup
	backupPath := configPath + fmt.Sprintf(".bak.%d", time.Now().UnixNano())
	if err := copyFile(configPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	return &transaction{
		configPath: configPath,
		backupPath: backupPath,
		origHash:   hash.Sum(nil),
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
	currentFile, err := os.Open(t.configPath)
	if err != nil {
		return false, fmt.Errorf("failed to open current config: %w", err)
	}
	defer currentFile.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, currentFile); err != nil {
		return false, fmt.Errorf("failed to hash current config: %w", err)
	}

	currentHash := hash.Sum(nil)
	return !bytes.Equal(currentHash, t.origHash), nil
}

// Add this type for script conflict resolution
type scriptConflict struct {
	name     string
	existing config.Script
	new      config.Script
}

func (s scriptConflict) String() string {
	return fmt.Sprintf("[%s] Existing: %q vs New: %q",
		s.name, s.existing.Description, s.new.Description)
}

// Non-interactive function for testing
func syncScripts(cfg *config.DotfilesConfig, scriptsDir string) ([]config.Script, []scriptConflict, error) {
	// Initialize scripts map to preserve existing configurations
	existingScripts := make(map[string]config.Script)
	for _, script := range cfg.Scripts.Install {
		existingScripts[script.Name] = script
	}

	// Find all executable files in scripts/install
	entries, err := os.ReadDir(scriptsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("failed to read scripts directory: %w", err)
	}

	// Build new scripts list
	var newScripts []config.Script
	var conflicts []scriptConflict
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if file is executable
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0111 == 0 { // Check executable bit
			continue
		}

		name := entry.Name()
		path := filepath.Join(scriptsDir, name)

		// Create new script config
		newScript := config.Script{
			Name:        name,
			Description: fmt.Sprintf("Script from %s", path),
			Path:        path,
		}

		if existing, exists := existingScripts[name]; exists {
			// Check if configs differ beyond just the path
			existing.Path = path // Update path
			if existing.Description != newScript.Description {
				conflicts = append(conflicts, scriptConflict{
					name:     name,
					existing: existing,
					new:      newScript,
				})
				continue
			}
			newScripts = append(newScripts, existing)
		} else {
			newScripts = append(newScripts, newScript)
		}
	}

	return newScripts, conflicts, nil
}

// chooseFrom is our mockable version of choose.From
func chooseFrom(options []string) (int, string, error) {
	width := len(fmt.Sprint(len(options)))
	for i, v := range options {
		fmt.Fprintf(Stdout, "%*d. %v\n", width, i+1, v)
	}

	scanner := bufio.NewScanner(Stdin)
	for {
		fmt.Fprint(Stdout, "#? ")
		if !scanner.Scan() {
			return -1, "", scanner.Err()
		}
		resp := scanner.Text()
		if resp == "q" {
			return -1, "", nil
		}
		n, err := strconv.Atoi(resp)
		if err == nil && n > 0 && n <= len(options) {
			return n - 1, options[n-1], nil
		}
	}
}

// Interactive resolution for real usage
func resolveConflictsInteractive(conflicts []scriptConflict) (map[string]config.Script, error) {
	resolved := make(map[string]config.Script)

	for _, conflict := range conflicts {
		options := []string{
			fmt.Sprintf("Keep existing: %s", conflict.existing.Description),
			fmt.Sprintf("Use new: %s", conflict.new.Description),
		}

		fmt.Fprintf(Stdout, "\nConflict for script %q:\n", conflict.name)
		idx, _, err := chooseFrom(options)
		if err != nil {
			return nil, fmt.Errorf("conflict resolution failed: %w", err)
		}
		if idx == -1 { // User quit
			return nil, fmt.Errorf("conflict resolution cancelled by user")
		}

		if idx == 0 {
			resolved[conflict.name] = conflict.existing
		} else {
			resolved[conflict.name] = conflict.new
		}
	}

	return resolved, nil
}

var Cmd = &bonzai.Cmd{
	Name:  "sync",
	Short: "synchronize install scripts from active namespace",
	Long: `
Synchronize install scripts from the active namespace into the local arara.yaml.
This will:
1. Find all executable files in the scripts/install directory
2. Add them to the local arara.yaml's install scripts section
3. Preserve existing script descriptions and configurations

Changes are applied atomically with automatic rollback on failure.
`,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(caller *bonzai.Cmd, args ...string) error {
		configPath := "arara.yaml"
		scriptsDir := "scripts/install"

		// Begin transaction
		tx, err := beginTransaction(configPath)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				tx.rollback()
			}
		}()

		// In tests, we might be in a directory with arara.yaml and no active namespace
		// Try to load config without namespace validation first
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Sync scripts and get conflicts
		newScripts, conflicts, err := syncScripts(cfg, scriptsDir)
		if err != nil {
			return err
		}

		// Resolve any conflicts interactively
		if len(conflicts) > 0 {
			fmt.Printf("\nFound %d script conflicts to resolve:\n", len(conflicts))
			resolved, err := resolveConflictsInteractive(conflicts)
			if err != nil {
				return err
			}
			for _, script := range resolved {
				newScripts = append(newScripts, script)
			}
		}

		// Check for concurrent modifications before writing
		if modified, err := tx.checkModified(); err != nil {
			return err
		} else if modified {
			if err := tx.rollback(); err != nil {
				return fmt.Errorf("failed to rollback after concurrent modification: %w", err)
			}
			return fmt.Errorf("config was modified during sync")
		}

		// Update config
		cfg.Scripts.Install = newScripts

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

		fmt.Printf("Synchronized %d install scripts\n", len(newScripts))
		return nil
	},
}

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

	_, err = io.Copy(destination, source)
	return err
}
