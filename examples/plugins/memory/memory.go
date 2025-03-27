package memory

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/BuddhiLW/arara/internal/app/compat"
)

// MinMemValidator checks if the system has enough memory
type MinMemValidator struct{}

// Name returns the validator name
func (v *MinMemValidator) Name() string {
	return "min-memory"
}

// Validate checks if the system has enough memory
func (v *MinMemValidator) Validate(value interface{}) bool {
	// If no value is provided, any amount of memory is acceptable
	if value == nil {
		return true
	}

	// Convert value to float64 (MB)
	var requiredMB float64
	switch val := value.(type) {
	case float64:
		requiredMB = val
	case int:
		requiredMB = float64(val)
	case string:
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}
		requiredMB = parsed
	default:
		return false
	}

	// Get available memory
	availableMB, err := getSystemMemoryMB()
	if err != nil {
		return false
	}

	return availableMB >= requiredMB
}

// getSystemMemoryMB returns the available memory in MB
func getSystemMemoryMB() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxMemory()
	case "darwin":
		return getDarwinMemory()
	default:
		// Fallback to Go runtime memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Sys) / 1024 / 1024, nil
	}
}

// getLinuxMemory gets the available memory on Linux systems
func getLinuxMemory() (float64, error) {
	cmd := exec.Command("free", "-m")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output from free command")
	}

	memLine := lines[1]
	fields := strings.Fields(memLine)
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected format in free command output")
	}

	totalMem, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, err
	}

	return totalMem, nil
}

// getDarwinMemory gets the available memory on macOS systems
func getDarwinMemory() (float64, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	memBytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, err
	}

	// Convert bytes to MB
	return float64(memBytes) / 1024 / 1024, nil
}

// Init registers the validator
func init() {
	err := compat.RegisterCustomValidator(&MinMemValidator{})
	if err != nil {
		// Just log the error for now
		fmt.Printf("Failed to register min-memory validator: %v\n", err)
	}
}