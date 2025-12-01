package di

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test interface for DI error testing
type TestService interface {
	GetName() string
}

type TestServiceImpl struct {
	name string
}

func (t *TestServiceImpl) GetName() string {
	return t.name
}

// Another test interface for dependency chain testing
type TestDependency interface {
	GetValue() int
}

type TestDependencyImpl struct {
	value int
}

func (t *TestDependencyImpl) GetValue() int {
	return t.value
}

func TestDIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		diErr    *DIError
		contains []string
	}{
		{
			name: "basic error",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"resolution",
				"no registration found",
			),
			contains: []string{
				"DI resolution failed",
				"TestService",
				"no registration found",
			},
		},
		{
			name: "error with dependency chain",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"resolution",
				"circular dependency detected",
			).WithDependencyChain([]string{"ServiceA", "ServiceB", "ServiceA"}),
			contains: []string{
				"Dependency chain: ServiceA -> ServiceB -> ServiceA",
			},
		},
		{
			name: "error with available registrations",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"resolution",
				"no registration found",
			).WithAvailableRegistrations([]string{"ServiceA", "ServiceB", "ServiceC"}),
			contains: []string{
				"Available registrations: ServiceA, ServiceB, ServiceC",
			},
		},
		{
			name: "error with suggestions",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"registration",
				"type mismatch",
			).WithSuggestions([]string{"Check interface compliance", "Verify method signatures"}),
			contains: []string{
				"Suggested fixes:",
				"- Check interface compliance",
				"- Verify method signatures",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := tt.diErr.Error()
			for _, expected := range tt.contains {
				assert.Contains(t, errorMsg, expected, "Error message should contain: %s", expected)
			}
		})
	}
}

func TestSuggestionEngine_GenerateSuggestions(t *testing.T) {
	engine := NewSuggestionEngine()

	// Update with some registered types
	registeredTypes := map[string]reflect.Type{
		"TestServiceInterface":    reflect.TypeOf((*TestService)(nil)).Elem(),
		"TestDependencyInterface": reflect.TypeOf((*TestDependency)(nil)).Elem(),
	}
	engine.UpdateRegisteredTypes(registeredTypes)

	tests := []struct {
		name             string
		diErr            *DIError
		expectedContains []string
	}{
		{
			name: "not registered error",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"resolution",
				"not registered",
			),
			expectedContains: []string{
				"Register the interface using container.Register",
				"Example registration:",
			},
		},
		{
			name: "circular dependency error",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"resolution",
				"circular dependency detected",
			),
			expectedContains: []string{
				"Break the circular dependency",
				"Use factory registration",
			},
		},
		{
			name: "type mismatch error",
			diErr: NewDIError(
				reflect.TypeOf((*TestService)(nil)).Elem(),
				"registration",
				"type mismatch",
			),
			expectedContains: []string{
				"Ensure the implementation actually implements",
				"Check for typos in interface method signatures",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := engine.GenerateSuggestions(tt.diErr)
			require.NotEmpty(t, suggestions, "Should generate suggestions")

			suggestionsText := strings.Join(suggestions, " ")
			for _, expected := range tt.expectedContains {
				assert.Contains(t, suggestionsText, expected,
					"Suggestions should contain: %s\nActual suggestions: %v", expected, suggestions)
			}
		})
	}
}

func TestSuggestionEngine_FindSimilarTypes(t *testing.T) {
	engine := NewSuggestionEngine()

	registeredTypes := map[string]reflect.Type{
		"UserService":    reflect.TypeOf((*TestService)(nil)).Elem(),
		"UserRepository": reflect.TypeOf((*TestDependency)(nil)).Elem(),
		"AuthService":    reflect.TypeOf((*TestService)(nil)).Elem(),
		"TestService":    reflect.TypeOf((*TestService)(nil)).Elem(),
	}
	engine.UpdateRegisteredTypes(registeredTypes)

	tests := []struct {
		name            string
		typeName        string
		expectedSimilar []string
	}{
		{
			name:            "exact partial match",
			typeName:        "User",
			expectedSimilar: []string{"UserService", "UserRepository"},
		},
		{
			name:            "similar service names",
			typeName:        "TestSvc",
			expectedSimilar: []string{"TestService"},
		},
		{
			name:            "no similar types",
			typeName:        "CompletelyDifferent",
			expectedSimilar: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similar := engine.findSimilarTypes(tt.typeName)

			if len(tt.expectedSimilar) == 0 {
				assert.Empty(t, similar, "Should not find similar types")
			} else {
				for _, expected := range tt.expectedSimilar {
					assert.Contains(t, similar, expected,
						"Should find similar type: %s\nActual similar: %v", expected, similar)
				}
			}
		})
	}
}

func TestEnhancedDIContainer_ErrorReporting(t *testing.T) {
	container := NewContainer()

	t.Run("registration error with suggestions", func(t *testing.T) {
		// Try to register with wrong type (concrete type instead of interface)
		err := container.RegisterType(reflect.TypeOf(&TestServiceImpl{}), &TestServiceImpl{name: "test"}, Singleton)

		require.Error(t, err)

		var diErr *DIError
		assert.ErrorAs(t, err, &diErr, "Should be a DIError")
		assert.Equal(t, "registration", diErr.Operation)
		assert.Contains(t, diErr.Error(), "type is not an interface")
		assert.Contains(t, diErr.Error(), "Suggested fixes:")
		assert.NotEmpty(t, diErr.SuggestedFixes)
	})

	t.Run("resolution error with available registrations", func(t *testing.T) {
		// Register some types first
		container.Reset()
		testService := &TestServiceImpl{name: "test"}
		container.RegisterType(reflect.TypeOf((*TestService)(nil)).Elem(), testService, Singleton)

		// Try to resolve unregistered type
		_, err := container.ResolveType(reflect.TypeOf((*TestDependency)(nil)).Elem())

		require.Error(t, err)

		var diErr *DIError
		assert.ErrorAs(t, err, &diErr, "Should be a DIError")
		assert.Equal(t, "resolution", diErr.Operation)
		assert.Contains(t, diErr.Error(), "no registration found")
		assert.Contains(t, diErr.Error(), "Available registrations:")
		assert.NotEmpty(t, diErr.AvailableRegistrations)
	})

	t.Run("circular dependency error with chain", func(t *testing.T) {
		container.Reset()

		// Create circular dependency with factories
		factory1 := func(c Container) (TestService, error) {
			// This will cause circular dependency
			_, err := Resolve[TestDependency](c)
			if err != nil {
				return nil, err
			}
			return &TestServiceImpl{name: "service1"}, nil
		}

		factory2 := func(c Container) (TestDependency, error) {
			// This will cause circular dependency
			_, err := Resolve[TestService](c)
			if err != nil {
				return nil, err
			}
			return &TestDependencyImpl{value: 42}, nil
		}

		RegisterFactory[TestService](container, factory1, Singleton)
		RegisterFactory[TestDependency](container, factory2, Singleton)

		// Try to resolve - should detect circular dependency
		_, err := Resolve[TestService](container)

		require.Error(t, err)

		var diErr *DIError
		if assert.ErrorAs(t, err, &diErr, "Should be a DIError") {
			// Since the error happens in factory resolution, it may be a factory error rather than pure circular dependency
			// Let's check if it contains information about the dependency chain
			errorMsg := diErr.Error()
			assert.Contains(t, errorMsg, "Dependency chain:", "Should show dependency chain")
			assert.NotEmpty(t, diErr.DependencyChain, "Should have dependency chain")
		}
	})
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected int
	}{
		{"", "", 0},
		{"hello", "", 5},
		{"", "world", 5},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"UserService", "UserSvc", 4},
		{"TestService", "TestSrv", 4},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			actual := levenshteinDistance(tt.s1, tt.s2)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
