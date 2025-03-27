package compat

import (
	"os"
	"runtime"
	"testing"
)

// TestOSValidator tests the OS validator
func TestOSValidator(t *testing.T) {
	osValidator, ok := getValidator("os")
	if !ok {
		t.Fatal("OS validator not found")
	}

	// Empty value should always return true
	if !osValidator("") {
		t.Error("OS validator should return true for empty value")
	}

	// Current OS should match
	if runtime.GOOS == "linux" {
		// On Linux systems, test with actual OS ID
		osInfo, err := getOSInfo()
		if err == nil {
			if !osValidator(osInfo["ID"]) {
				t.Errorf("OS validator should return true for current OS ID: %s", osInfo["ID"])
			}
		}
	} else if runtime.GOOS == "darwin" {
		// On macOS, test with "darwin"
		if !osValidator("darwin") {
			t.Error("OS validator should return true for 'darwin' on macOS")
		}
	}

	// Non-existent OS should not match
	if osValidator("nonexistent-os") {
		t.Error("OS validator should return false for non-existent OS")
	}
}

// TestArchValidator tests the architecture validator
func TestArchValidator(t *testing.T) {
	archValidator, ok := getValidator("arch")
	if !ok {
		t.Fatal("Architecture validator not found")
	}

	// Empty value should always return true
	if !archValidator("") {
		t.Error("Architecture validator should return true for empty value")
	}

	// Current architecture should match
	if !archValidator(runtime.GOARCH) {
		t.Errorf("Architecture validator should return true for current architecture: %s", runtime.GOARCH)
	}

	// Non-existent architecture should not match
	if archValidator("nonexistent-arch") {
		t.Error("Architecture validator should return false for non-existent architecture")
	}
}

// TestShellValidator tests the shell validator
func TestShellValidator(t *testing.T) {
	shellValidator, ok := getValidator("shell")
	if !ok {
		t.Fatal("Shell validator not found")
	}

	// Empty value should always return true
	if !shellValidator("") {
		t.Error("Shell validator should return true for empty value")
	}

	// Get the current shell
	currentShell := os.Getenv("SHELL")
	if currentShell != "" {
		// Extract the shell name
		var shellName string
		for i := len(currentShell) - 1; i >= 0; i-- {
			if currentShell[i] == '/' {
				shellName = currentShell[i+1:]
				break
			}
		}

		if shellName != "" && !shellValidator(shellName) {
			t.Errorf("Shell validator should return true for current shell: %s", shellName)
		}
	}

	// Non-existent shell should not match
	if shellValidator("nonexistent-shell") {
		t.Error("Shell validator should return false for non-existent shell")
	}
}

// mockCustomValidator is a mock custom validator for testing
type mockCustomValidator struct {
	name  string
	value interface{}
}

// Name returns the validator name
func (m *mockCustomValidator) Name() string {
	return m.name
}

// Validate performs validation
func (m *mockCustomValidator) Validate(value interface{}) bool {
	if value == nil {
		return true
	}

	// Check if the provided value matches the expected value
	return value == m.value
}

// TestCustomValidator tests the custom validator system
func TestCustomValidator(t *testing.T) {
	// Register a mock validator
	mockValidator := &mockCustomValidator{
		name:  "test-validator",
		value: "expected-value",
	}

	err := RegisterCustomValidator(mockValidator)
	if err != nil {
		t.Fatalf("Failed to register custom validator: %v", err)
	}

	// Test with string requirement
	stringReq := "test-validator"
	if !checkStringReq(stringReq) {
		t.Errorf("Custom validator should return true for string requirement: %s", stringReq)
	}

	// Test with map requirement (correct value)
	mapReq := map[string]interface{}{
		"name":  "test-validator",
		"value": "expected-value",
	}
	if !checkMapReq(mapReq) {
		t.Errorf("Custom validator should return true for map requirement with correct value")
	}

	// Test with map requirement (incorrect value)
	incorrectMapReq := map[string]interface{}{
		"name":  "test-validator",
		"value": "incorrect-value",
	}
	if checkMapReq(incorrectMapReq) {
		t.Errorf("Custom validator should return false for map requirement with incorrect value")
	}

	// Test with non-existent validator
	nonExistentReq := "non-existent-validator"
	if checkStringReq(nonExistentReq) {
		t.Errorf("Custom validator should return false for non-existent validator")
	}
}

// TestCheck tests the main Check function
func TestCheck(t *testing.T) {
	// Test with empty CompatSpec (should always pass)
	emptySpec := CompatSpec{}
	if !Check(emptySpec) {
		t.Error("Check should return true for empty CompatSpec")
	}

	// Test with current OS
	// Note: We need to use OS ID from /etc/os-release, not runtime.GOOS
	var currentOS string
	if runtime.GOOS == "linux" {
		osInfo, err := getOSInfo()
		if err == nil {
			currentOS = osInfo["ID"]
		} else {
			currentOS = runtime.GOOS
		}
	} else {
		currentOS = runtime.GOOS
	}
	
	currentOSSpec := CompatSpec{
		OS: currentOS,
	}
	if !Check(currentOSSpec) {
		t.Errorf("Check should return true for current OS: %s", currentOS)
	}

	// Test with current architecture
	currentArchSpec := CompatSpec{
		Arch: runtime.GOARCH,
	}
	if !Check(currentArchSpec) {
		t.Errorf("Check should return true for current architecture: %s", runtime.GOARCH)
	}

	// Test with non-existent OS (should fail)
	nonExistentOSSpec := CompatSpec{
		OS: "nonexistent-os",
	}
	if Check(nonExistentOSSpec) {
		t.Error("Check should return false for non-existent OS")
	}

	// First make sure we have a clean registry
	customRegistry.Lock()
	customRegistry.validators = map[string]CustomValidator{}
	customRegistry.Unlock()

	// Register the test validator
	testValidator := &mockCustomValidator{
		name:  "test-validator",
		value: "test-value",
	}
	if err := RegisterCustomValidator(testValidator); err != nil {
		t.Fatalf("Failed to register custom validator: %v", err)
	}

	// Test with custom validator (should pass)
	customSpec := CompatSpec{
		Custom: []any{
			map[string]interface{}{
				"name":  "test-validator",
				"value": "test-value",
			},
		},
	}
	if !Check(customSpec) {
		t.Error("Check should return true for valid custom validator")
	}

	// Test with invalid custom validator (should fail)
	invalidCustomSpec := CompatSpec{
		Custom: []any{
			map[string]interface{}{
				"name":  "test-validator",
				"value": "invalid-value",
			},
		},
	}
	if Check(invalidCustomSpec) {
		t.Error("Check should return false for invalid custom validator")
	}
}