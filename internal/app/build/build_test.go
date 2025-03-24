package build

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestListCmd(t *testing.T) {
	// Capture stdout to verify output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run the list command
	err := listCmd.Do(listCmd)
	if err != nil {
		t.Fatalf("Failed to execute list command: %v", err)
	}
	
	// Reset stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	// Verify output contains expected build steps
	expectedSteps := []string{
		"backup",
		"link",
		"xmonad",
	}
	
	for _, step := range expectedSteps {
		if !strings.Contains(output, step) {
			t.Errorf("Expected output to contain build step '%s', but it didn't", step)
		}
	}
}

// Mock function to use for testing the install command without executing external commands
// We're not testing this now because it would require significant mocking of external commands
func mockExecCommand(command string, args ...string) *mockCmd {
	return &mockCmd{
		command: command,
		args:    args,
	}
}

type mockCmd struct {
	command string
	args    []string
	stdout  io.Writer
	stderr  io.Writer
}

func (c *mockCmd) Run() error {
	// Mock successful command execution
	switch c.command {
	case "arara":
		if len(c.args) >= 1 && c.args[0] == "setup" {
			if c.stdout != nil {
				io.WriteString(c.stdout, "Mock arara setup command executed successfully")
			}
		}
	case "git":
		if len(c.args) >= 1 && c.args[0] == "clone" {
			if c.stdout != nil {
				io.WriteString(c.stdout, "Mock git clone executed successfully")
			}
		}
	case "bash":
		if c.stdout != nil {
			io.WriteString(c.stdout, "Mock bash command executed successfully")
		}
	}
	return nil
}