package di

import (
	"fmt"
	"reflect"
)

// DependencyError represents errors that occur during dependency operations
type DependencyError struct {
	// Type is the interface type that failed to resolve
	Type reflect.Type
	// Operation describes what operation failed (registration, resolution, etc.)
	Operation string
	// Message provides human-readable error details
	Message string
	// Context provides additional error context for debugging
	Context map[string]interface{}
}

// Error implements the error interface
func (e *DependencyError) Error() string {
	typeName := "unknown"
	if e.Type != nil {
		typeName = e.Type.String()
	}
	return fmt.Sprintf("dependency %s failed for type %s: %s", e.Operation, typeName, e.Message)
}

// CircularDependencyError represents a circular dependency detection error
type CircularDependencyError struct {
	// Path shows the dependency resolution path that led to the cycle
	Path []string
	// Type is the interface type where the cycle was detected
	Type reflect.Type
}

// Error implements the error interface
func (e *CircularDependencyError) Error() string {
	typeName := "unknown"
	if e.Type != nil {
		typeName = e.Type.String()
	}
	return fmt.Sprintf("circular dependency detected for type %s in path: %v", typeName, e.Path)
}

// ValidationError represents a dependency graph validation error
type ValidationError struct {
	// Errors contains all validation errors found
	Errors []error
	// Graph provides the dependency graph for debugging
	Graph map[string][]string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("dependency validation failed: %s", e.Errors[0].Error())
	}
	return fmt.Sprintf("dependency validation failed with %d errors: %v", len(e.Errors), e.Errors)
}

// NewDependencyError creates a new dependency error with context
func NewDependencyError(typ reflect.Type, operation, message string, context map[string]interface{}) *DependencyError {
	return &DependencyError{
		Type:      typ,
		Operation: operation,
		Message:   message,
		Context:   context,
	}
}

// NewCircularDependencyError creates a new circular dependency error
func NewCircularDependencyError(typ reflect.Type, path []string) *CircularDependencyError {
	return &CircularDependencyError{
		Type: typ,
		Path: path,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(errors []error, graph map[string][]string) *ValidationError {
	return &ValidationError{
		Errors: errors,
		Graph:  graph,
	}
}
