package compat

import (
	"fmt"
	"sync"
)

// CustomValidator defines the interface for custom validators
type CustomValidator interface {
	// Name returns the unique name of the validator
	Name() string
	
	// Validate checks if the given value meets the validator's criteria
	Validate(value interface{}) bool
}

// customRegistry stores custom validators
var customRegistry = struct {
	sync.RWMutex
	validators map[string]CustomValidator
}{
	validators: map[string]CustomValidator{},
}

// RegisterCustomValidator registers a custom validator
func RegisterCustomValidator(validator CustomValidator) error {
	if validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	name := validator.Name()
	if name == "" {
		return fmt.Errorf("validator name cannot be empty")
	}

	customRegistry.Lock()
	defer customRegistry.Unlock()
	
	// Check if the validator with the same name already exists
	if _, exists := customRegistry.validators[name]; exists {
		return fmt.Errorf("validator with name '%s' already registered", name)
	}
	
	customRegistry.validators[name] = validator
	return nil
}

// CheckCustom validates custom compatibility requirements
func CheckCustom(customReqs []interface{}) bool {
	if len(customReqs) == 0 {
		return true // No custom requirements
	}

	for _, req := range customReqs {
		// Handle different custom requirement formats
		switch r := req.(type) {
		case map[string]interface{}:
			// Process map-based requirement
			if !checkMapReq(r) {
				return false
			}
		case string:
			// Process string-based requirement (validator name only)
			if !checkStringReq(r) {
				return false
			}
		default:
			// Unknown format, fail validation
			return false
		}
	}

	return true
}

// checkMapReq validates a map-based custom requirement
func checkMapReq(req map[string]interface{}) bool {
	// Extract the validator name
	nameVal, ok := req["name"]
	if !ok {
		return false // No validator name specified
	}
	
	name, ok := nameVal.(string)
	if !ok {
		return false // Name is not a string
	}
	
	// Check if the validator exists
	customRegistry.RLock()
	validator, ok := customRegistry.validators[name]
	customRegistry.RUnlock()
	
	if !ok {
		return false // Validator not found
	}
	
	// Extract the value to validate
	valueVal, ok := req["value"]
	if !ok {
		// If no value is provided, pass nil to the validator
		return validator.Validate(nil)
	}
	
	// Validate the value
	return validator.Validate(valueVal)
}

// checkStringReq validates a string-based custom requirement (validator name only)
func checkStringReq(name string) bool {
	// Check if the validator exists
	customRegistry.RLock()
	validator, ok := customRegistry.validators[name]
	customRegistry.RUnlock()
	
	if !ok {
		return false // Validator not found
	}
	
	// Validate with nil value
	return validator.Validate(nil)
}