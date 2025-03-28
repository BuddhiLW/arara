package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/arara/internal/pkg/config"
	"github.com/stretchr/testify/suite"
)

type BackupTestSuite struct {
	suite.Suite
	// tmpDir will be our temporary working directory
	tmpDir string
	// testFiles holds full paths to our dummy files and their expected contents
	testFiles map[string]string
	// originalHome holds the original HOME environment variable to restore later
	originalHome string
	// originalWd holds the original working directory to restore later
	originalWd string
}

func (s *BackupTestSuite) SetupTest() {
	var err error
	// Create temporary directory for our suite
	s.tmpDir, err = os.MkdirTemp("", "backup-suite")
	s.Require().NoError(err, "Failed to create temporary directory")

	// Create needed directories
	err = os.MkdirAll(filepath.Join(s.tmpDir, "config"), 0755)
	s.Require().NoError(err, "Failed to create config dir")
	err = os.MkdirAll(filepath.Join(s.tmpDir, "local"), 0755)
	s.Require().NoError(err, "Failed to create local dir")

	// Create dummy test files
	s.testFiles = map[string]string{
		filepath.Join(s.tmpDir, "config", "test.conf"): "test config content",
		filepath.Join(s.tmpDir, "local", "data.txt"):   "test data content",
	}
	for path, content := range s.testFiles {
		err = os.WriteFile(path, []byte(content), 0644)
		s.Require().NoError(err, "Failed to create test file %s", path)
	}

	// Save current HOME and working directory
	s.originalHome = os.Getenv("HOME")
	s.originalWd, err = os.Getwd()
	s.Require().NoError(err, "Failed to get current working directory")

	// Set environment variables so that the backup command uses our temporary directory
	os.Setenv("HOME", s.tmpDir)
	os.Setenv("TEST_MODE", "1")

	// Change working directory so that config.LoadConfig("arara.yaml") finds our file
	err = os.Chdir(s.tmpDir)
	s.Require().NoError(err, "Failed to change working directory")
}

func (s *BackupTestSuite) TearDownTest() {
	// Restore working directory and HOME variable, and remove the tmpDir
	_ = os.Chdir(s.originalWd)
	os.Setenv("HOME", s.originalHome)
	os.Unsetenv("TEST_MODE")
	_ = os.RemoveAll(s.tmpDir)
}

// createTestConfig writes a DotfilesConfig to an arara.yaml file in s.tmpDir.
// The caller passes the list of directories to back up.
func (s *BackupTestSuite) createTestConfig(dirs []string) {
	cfg := &config.DotfilesConfig{
		Name:        "test",
		Description: "Test config",
		Setup: struct {
			BackupDirs  []string      `yaml:"backup_dirs"`
			CoreLinks   []config.Link `yaml:"core_links"`
			ConfigLinks []config.Link `yaml:"config_links"`
		}{
			BackupDirs: dirs,
		},
	}
	configPath := filepath.Join(s.tmpDir, "arara.yaml")
	data, err := cfg.Marshal()
	s.Require().NoError(err, "Failed to marshal config")
	err = os.WriteFile(configPath, data, 0644)
	s.Require().NoError(err, "Failed to write config file")
}

// TestDirectoryCreation checks that the backup command creates a backup directory.
func (s *BackupTestSuite) TestDirectoryCreation() {
	s.createTestConfig([]string{
		filepath.Join(s.tmpDir, "config"),
	})

	err := Cmd.Do(Cmd)
	s.Require().NoError(err, "Backup command failed")

	entries, err := os.ReadDir(s.tmpDir)
	s.Require().NoError(err, "Failed to read temporary directory")
	var found bool
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "dotbk-") {
			found = true
			break
		}
	}
	s.True(found, "Backup directory was not created")
}

// TestFileContent verifies the contents of backed-up files.
func (s *BackupTestSuite) TestFileContent() {
	s.createTestConfig([]string{
		filepath.Join(s.tmpDir, "config"),
		filepath.Join(s.tmpDir, "local"),
	})

	err := Cmd.Do(Cmd)
	s.Require().NoError(err, "Backup command failed")

	entries, err := os.ReadDir(s.tmpDir)
	s.Require().NoError(err)
	var backupDir string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "dotbk-") {
			backupDir = filepath.Join(s.tmpDir, entry.Name())
			break
		}
	}
	s.Require().NotEmpty(backupDir, "Backup directory not found")

	// Check contents of each expected file in the backed-up directories.
	for originalPath, expectedContent := range s.testFiles {
		backupPath := filepath.Join(backupDir, filepath.Base(filepath.Dir(originalPath)), filepath.Base(originalPath))
		content, err := os.ReadFile(backupPath)
		s.Require().NoError(err, "Failed to read backup file %s", backupPath)
		s.Equal(expectedContent, string(content), "File content mismatch for %s", filepath.Base(originalPath))
	}
}

// TestOriginalRemoval verifies that after backup the original directories are removed.
func (s *BackupTestSuite) TestOriginalRemoval() {
	dirsToBackup := []string{
		filepath.Join(s.tmpDir, "config"),
		filepath.Join(s.tmpDir, "local"),
	}
	s.createTestConfig(dirsToBackup)

	err := Cmd.Do(Cmd)
	s.Require().NoError(err, "Backup command failed")

	for _, dir := range dirsToBackup {
		_, err = os.Stat(dir)
		s.True(os.IsNotExist(err), "Original directory still exists: %s", dir)
	}
}

// TestCopyDir_Basic tests the copyDir helper function with a single file.
func (s *BackupTestSuite) TestCopyDir_Basic() {
	srcDir, err := os.MkdirTemp("", "copytest-src")
	s.Require().NoError(err)
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "copytest-dst")
	s.Require().NoError(err)
	defer os.RemoveAll(dstDir)

	// Create a simple test file in srcDir.
	testContent := "test content"
	testFile := filepath.Join(srcDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	s.Require().NoError(err)

	err = copyDir(srcDir, dstDir)
	s.Require().NoError(err)

	copiedContent, err := os.ReadFile(filepath.Join(dstDir, "test.txt"))
	s.Require().NoError(err)
	s.Equal(testContent, string(copiedContent), "Copied file content mismatch")
}

// TestCopyDir_NestedStructure tests copyDir with a nested directory structure.
func (s *BackupTestSuite) TestCopyDir_NestedStructure() {
	srcDir, err := os.MkdirTemp("", "copytest-src")
	s.Require().NoError(err)
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "copytest-dst")
	s.Require().NoError(err)
	defer os.RemoveAll(dstDir)

	// Establish nested files and directories in srcDir.
	testFiles := map[string]string{
		"file1.txt":            "content1",
		"subdir/file2.txt":     "content2",
		"subdir/sub/file3.txt": "content3",
	}
	for relPath, content := range testFiles {
		fullPath := filepath.Join(srcDir, relPath)
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		s.Require().NoError(err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		s.Require().NoError(err)
	}

	err = copyDir(srcDir, dstDir)
	s.Require().NoError(err)

	for relPath, expectedContent := range testFiles {
		fullPath := filepath.Join(dstDir, relPath)
		data, err := os.ReadFile(fullPath)
		s.Require().NoError(err, "Failed to read copied file %s", fullPath)
		s.Equal(expectedContent, string(data), "Copied file %s content mismatch", relPath)
	}
}

func TestBackupTestSuite(t *testing.T) {
	suite.Run(t, new(BackupTestSuite))
}
