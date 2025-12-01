package di

import (
	"fmt"
	"reflect"
	"strings"
)

// DIError represents enhanced dependency injection errors with contextual debugging information
type DIError struct {
	// Type is the interface type that failed to resolve
	Type reflect.Type
	// Operation describes what operation failed (registration, resolution, etc.)
	Operation string
	// Message provides human-readable error details
	Message string
	// DependencyChain shows the full resolution path that led to this error
	DependencyChain []string
	// AvailableRegistrations lists all currently registered interface types
	AvailableRegistrations []string
	// SuggestedFixes provides actionable suggestions to resolve the error
	SuggestedFixes []string
	// Context provides additional error context for debugging
	Context map[string]interface{}
}

// Error implements the error interface with enhanced contextual information
func (e *DIError) Error() string {
	var builder strings.Builder

	typeName := "unknown"
	if e.Type != nil {
		typeName = e.Type.String()
	}

	// Primary error message
	builder.WriteString(fmt.Sprintf("DI %s failed for type %s: %s", e.Operation, typeName, e.Message))

	// Add dependency chain if available
	if len(e.DependencyChain) > 0 {
		builder.WriteString(fmt.Sprintf("\nDependency chain: %s", strings.Join(e.DependencyChain, " -> ")))
	}

	// Add available registrations if this is a resolution error
	if strings.Contains(e.Operation, "resolution") && len(e.AvailableRegistrations) > 0 {
		builder.WriteString(fmt.Sprintf("\nAvailable registrations: %s", strings.Join(e.AvailableRegistrations, ", ")))
	}

	// Add suggested fixes
	if len(e.SuggestedFixes) > 0 {
		builder.WriteString("\nSuggested fixes:")
		for _, fix := range e.SuggestedFixes {
			builder.WriteString(fmt.Sprintf("\n  - %s", fix))
		}
	}

	return builder.String()
}

// NewDIError creates a new enhanced DI error with contextual information
func NewDIError(typ reflect.Type, operation, message string) *DIError {
	return &DIError{
		Type:                   typ,
		Operation:              operation,
		Message:                message,
		DependencyChain:        make([]string, 0),
		AvailableRegistrations: make([]string, 0),
		SuggestedFixes:         make([]string, 0),
		Context:                make(map[string]interface{}),
	}
}

// WithDependencyChain adds dependency resolution chain information
func (e *DIError) WithDependencyChain(chain []string) *DIError {
	e.DependencyChain = chain
	return e
}

// WithAvailableRegistrations adds information about currently registered types
func (e *DIError) WithAvailableRegistrations(registrations []string) *DIError {
	e.AvailableRegistrations = registrations
	return e
}

// WithSuggestions adds actionable suggestions to resolve the error
func (e *DIError) WithSuggestions(fixes []string) *DIError {
	e.SuggestedFixes = fixes
	return e
}

// WithContext adds additional debugging context
func (e *DIError) WithContext(key string, value interface{}) *DIError {
	e.Context[key] = value
	return e
}

// SuggestionEngine provides intelligent suggestions for common DI registration mistakes
type SuggestionEngine struct {
	// registeredTypes tracks all registered interface types for lookup
	registeredTypes map[string]reflect.Type
	// commonPatterns maps common error patterns to suggestions
	commonPatterns map[string][]string
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() *SuggestionEngine {
	engine := &SuggestionEngine{
		registeredTypes: make(map[string]reflect.Type),
		commonPatterns:  make(map[string][]string),
	}

	// Initialize common error patterns and their suggested fixes
	engine.initializeCommonPatterns()

	return engine
}

// initializeCommonPatterns sets up common DI error patterns and their solutions
func (s *SuggestionEngine) initializeCommonPatterns() {
	s.commonPatterns = map[string][]string{
		"not registered": {
			"Register the interface using container.Register[MyInterface](implementation)",
			"Check if you're using the correct interface type (not the implementation type)",
			"Ensure the registration happens before attempting resolution",
		},
		"circular dependency": {
			"Break the circular dependency by introducing an intermediate interface",
			"Use factory registration to defer one of the dependencies",
			"Consider redesigning the dependency relationship",
		},
		"type mismatch": {
			"Ensure the implementation actually implements the requested interface",
			"Check for typos in interface method signatures",
			"Verify generic type parameters match exactly",
		},
		"nil implementation": {
			"Provide a non-nil implementation when registering",
			"Check if the factory function returns nil",
			"Ensure the implementation is properly initialized",
		},
		"scope conflict": {
			"Verify the scope (Singleton, Transient, Scoped) is appropriate",
			"Consider using Scoped dependencies for request-bound resources",
			"Check if dependency lifetimes are compatible",
		},
	}
}

// UpdateRegisteredTypes updates the engine's knowledge of registered types
func (s *SuggestionEngine) UpdateRegisteredTypes(types map[string]reflect.Type) {
	s.registeredTypes = types
}

// GenerateSuggestions creates contextual suggestions based on the error and available registrations
func (s *SuggestionEngine) GenerateSuggestions(err *DIError) []string {
	suggestions := make([]string, 0)

	// Add pattern-based suggestions
	for pattern, fixes := range s.commonPatterns {
		if strings.Contains(strings.ToLower(err.Message), pattern) {
			suggestions = append(suggestions, fixes...)
		}
	}

	// Add type-specific suggestions
	if err.Type != nil {
		typeName := err.Type.String()

		// Suggest similar registered types if this one isn't found
		if strings.Contains(err.Message, "not registered") {
			similar := s.findSimilarTypes(typeName)
			for _, similarType := range similar {
				suggestions = append(suggestions,
					fmt.Sprintf("Did you mean to use '%s' instead of '%s'?", similarType, typeName))
			}
		}

		// Suggest concrete registration example
		if err.Type.Kind() == reflect.Interface {
			suggestions = append(suggestions,
				fmt.Sprintf("Example registration: container.Register[%s](myImplementation)",
					getShortTypeName(typeName)))
		}
	}

	// Add general debugging suggestions
	suggestions = append(suggestions,
		"Use container.GetDependencyGraph() to inspect current registrations",
		"Enable debug mode for detailed dependency resolution tracing")

	return suggestions
}

// findSimilarTypes finds registered types similar to the requested type name
func (s *SuggestionEngine) findSimilarTypes(typeName string) []string {
	similar := make([]string, 0)

	targetLower := strings.ToLower(typeName)

	for registeredName := range s.registeredTypes {
		registeredLower := strings.ToLower(registeredName)

		// Look for partial matches or similar naming patterns
		if strings.Contains(registeredLower, targetLower) ||
			strings.Contains(targetLower, registeredLower) ||
			levenshteinDistance(targetLower, registeredLower) <= 5 {
			similar = append(similar, registeredName)
		}
	}

	return similar
}

// getShortTypeName extracts the short type name from a full type string
func getShortTypeName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := 1; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,
					matrix[i][j-1]+1,
					matrix[i-1][j-1]+1,
				)
			}
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
