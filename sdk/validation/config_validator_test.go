package validation

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration struct
type TestConfig struct {
	ServerPort int           `validate:"required" env:"SERVER_PORT" default:"8080"`
	Database   string        `validate:"required,pattern=^mongodb://" env:"DATABASE_URL" description:"MongoDB connection URL"`
	LogLevel   string        `validate:"oneof=DEBUG|INFO|WARN|ERROR" env:"LOG_LEVEL" default:"INFO"`
	MaxRetries *int          `validate:"min=1,max=10" env:"MAX_RETRIES" default:"3"`
	Features   FeatureConfig `description:"Feature flags configuration"`
}

type FeatureConfig struct {
	EnableAuth bool   `validate:"required" env:"ENABLE_AUTH" default:"true"`
	ApiKey     string `validate:"required,min=16" env:"API_KEY" description:"API key for external services"`
}

func TestConfigValidator_Validate(t *testing.T) {
	validator := NewConfigValidator(true)

	t.Run("valid configuration", func(t *testing.T) {
		maxRetries := 5
		config := TestConfig{
			ServerPort: 8080,
			Database:   "mongodb://localhost:27017/test",
			LogLevel:   "INFO",
			MaxRetries: &maxRetries,
			Features: FeatureConfig{
				EnableAuth: true,
				ApiKey:     "abcd1234567890ef",
			},
		}

		err := validator.Validate(config)
		assert.NoError(t, err)
	})

	t.Run("missing required field", func(t *testing.T) {
		config := TestConfig{
			// ServerPort missing (required)
			Database: "mongodb://localhost:27017/test",
			LogLevel: "INFO",
		}

		err := validator.Validate(config)
		require.Error(t, err)

		var configErrors *ConfigValidationErrors
		assert.ErrorAs(t, err, &configErrors)
		assert.True(t, configErrors.HasErrors())

		// Should have errors for missing ServerPort and missing nested ApiKey
		errors := configErrors.Errors
		assert.GreaterOrEqual(t, len(errors), 1)

		// Check for ServerPort error
		found := false
		for _, validationErr := range errors {
			if validationErr.FieldPath == "ServerPort" && validationErr.ValidationRule == "required" {
				found = true
				assert.Contains(t, validationErr.Message, "required")
				assert.NotEmpty(t, validationErr.Suggestions)
				break
			}
		}
		assert.True(t, found, "Should have validation error for missing ServerPort")
	})

	t.Run("invalid pattern", func(t *testing.T) {
		config := TestConfig{
			ServerPort: 8080,
			Database:   "invalid-url", // doesn't match mongodb:// pattern
			LogLevel:   "INFO",
		}

		err := validator.Validate(config)
		require.Error(t, err)

		var configErrors *ConfigValidationErrors
		assert.ErrorAs(t, err, &configErrors)

		errors := configErrors.Errors
		found := false
		for _, validationErr := range errors {
			if validationErr.FieldPath == "Database" && validationErr.ValidationRule == "pattern" {
				found = true
				assert.Contains(t, validationErr.Message, "pattern")
				assert.Contains(t, validationErr.Error(), "mongodb://")
				break
			}
		}
		assert.True(t, found, "Should have pattern validation error for Database")
	})

	t.Run("invalid enum value", func(t *testing.T) {
		config := TestConfig{
			ServerPort: 8080,
			Database:   "mongodb://localhost:27017/test",
			LogLevel:   "INVALID", // not in allowed values
		}

		err := validator.Validate(config)
		require.Error(t, err)

		var configErrors *ConfigValidationErrors
		assert.ErrorAs(t, err, &configErrors)

		errors := configErrors.Errors
		found := false
		for _, validationErr := range errors {
			if validationErr.FieldPath == "LogLevel" && validationErr.ValidationRule == "allowedValues" {
				found = true
				assert.Contains(t, validationErr.Message, "allowed values")
				// The error should contain the values separated by commas
				assert.Contains(t, validationErr.Error(), "DEBUG, INFO, WARN, ERROR")
				break
			}
		}
		assert.True(t, found, "Should have enum validation error for LogLevel")
	})
}

func TestConfigValidator_ValidateField(t *testing.T) {
	validator := NewConfigValidator(false)

	tests := []struct {
		name      string
		fieldPath string
		value     string
		rules     FieldValidationRules
		wantError bool
		errorRule string
	}{
		{
			name:      "valid required field",
			fieldPath: "database.url",
			value:     "mongodb://localhost:27017",
			rules:     FieldValidationRules{Required: true},
			wantError: false,
		},
		{
			name:      "missing required field",
			fieldPath: "database.url",
			value:     "",
			rules:     FieldValidationRules{Required: true},
			wantError: true,
			errorRule: "required",
		},
		{
			name:      "valid pattern match",
			fieldPath: "email",
			value:     "user@example.com",
			rules:     FieldValidationRules{Pattern: regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)},
			wantError: false,
		},
		{
			name:      "invalid pattern match",
			fieldPath: "email",
			value:     "invalid-email",
			rules:     FieldValidationRules{Pattern: regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)},
			wantError: true,
			errorRule: "pattern",
		},
		{
			name:      "value too short",
			fieldPath: "password",
			value:     "123",
			rules:     FieldValidationRules{MinLength: intPtr(8)},
			wantError: true,
			errorRule: "minLength",
		},
		{
			name:      "value too long",
			fieldPath: "username",
			value:     "verylongusernamethatexceedsmaximum",
			rules:     FieldValidationRules{MaxLength: intPtr(20)},
			wantError: true,
			errorRule: "maxLength",
		},
		{
			name:      "valid allowed value",
			fieldPath: "log.level",
			value:     "INFO",
			rules:     FieldValidationRules{AllowedValues: []string{"DEBUG", "INFO", "WARN", "ERROR"}},
			wantError: false,
		},
		{
			name:      "invalid allowed value",
			fieldPath: "log.level",
			value:     "INVALID",
			rules:     FieldValidationRules{AllowedValues: []string{"DEBUG", "INFO", "WARN", "ERROR"}},
			wantError: true,
			errorRule: "allowedValues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.fieldPath, tt.value, tt.rules)

			if tt.wantError {
				require.Error(t, err)
				var configErr *ConfigValidationError
				assert.ErrorAs(t, err, &configErr)
				assert.Equal(t, tt.errorRule, configErr.ValidationRule)
				assert.Equal(t, tt.fieldPath, configErr.FieldPath)
				assert.NotEmpty(t, configErr.Suggestions)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_CustomRules(t *testing.T) {
	validator := NewConfigValidator(false)

	// Add custom validation rule
	validator.AddValidationRule("positive_number", func(value interface{}, config interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}

		// Parse as integer and check if positive
		if num, err := strconv.Atoi(str); err != nil {
			return fmt.Errorf("value must be a valid integer")
		} else if num <= 0 {
			return fmt.Errorf("value must be a positive number")
		}

		return nil
	})

	t.Run("valid custom rule", func(t *testing.T) {
		rules := FieldValidationRules{
			CustomRules: []string{"positive_number"},
		}

		err := validator.ValidateField("count", "42", rules)
		assert.NoError(t, err)
	})

	t.Run("invalid custom rule", func(t *testing.T) {
		rules := FieldValidationRules{
			CustomRules: []string{"positive_number"},
		}

		err := validator.ValidateField("count", "-5", rules)
		require.Error(t, err)

		var configErr *ConfigValidationError
		assert.ErrorAs(t, err, &configErr)
		assert.Equal(t, "positive_number", configErr.ValidationRule)
		assert.Contains(t, configErr.Message, "positive number")
	})
}

func TestEnvResolver_ResolveEnvVar(t *testing.T) {
	resolver := NewEnvResolver(false) // Disable debugging for tests

	t.Run("resolve existing string env var", func(t *testing.T) {
		os.Setenv("TEST_STRING", "hello-world")
		defer os.Unsetenv("TEST_STRING")

		value, err := resolver.ResolveEnvVar("TEST_STRING", "config.message", reflect.TypeOf(""))
		assert.NoError(t, err)
		assert.Equal(t, "hello-world", value)
	})

	t.Run("resolve existing int env var", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		value, err := resolver.ResolveEnvVar("TEST_INT", "config.port", reflect.TypeOf(0))
		assert.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("resolve existing bool env var", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "true")
		defer os.Unsetenv("TEST_BOOL")

		value, err := resolver.ResolveEnvVar("TEST_BOOL", "config.enabled", reflect.TypeOf(false))
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("use default value for missing env var", func(t *testing.T) {
		resolver.SetDefault("MISSING_VAR", "default-value")

		value, err := resolver.ResolveEnvVar("MISSING_VAR", "config.missing", reflect.TypeOf(""))
		assert.NoError(t, err)
		assert.Equal(t, "default-value", value)
	})

	t.Run("error for missing env var without default", func(t *testing.T) {
		value, err := resolver.ResolveEnvVar("COMPLETELY_MISSING", "config.missing", reflect.TypeOf(""))
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "not set and no default provided")
	})

	t.Run("error for invalid int conversion", func(t *testing.T) {
		os.Setenv("INVALID_INT", "not-a-number")
		defer os.Unsetenv("INVALID_INT")

		value, err := resolver.ResolveEnvVar("INVALID_INT", "config.port", reflect.TypeOf(0))
		assert.Error(t, err)
		assert.Nil(t, value)
	})
}

func TestEnvResolver_DebuggingAndLogging(t *testing.T) {
	resolver := NewEnvResolver(true)

	t.Run("resolution log tracking", func(t *testing.T) {
		os.Setenv("LOG_TEST", "debug")
		defer os.Unsetenv("LOG_TEST")

		resolver.SetDefault("DEFAULT_TEST", "default-val")

		// Resolve some variables
		resolver.ResolveEnvVar("LOG_TEST", "config.logLevel", reflect.TypeOf(""))
		resolver.ResolveEnvVar("DEFAULT_TEST", "config.defaultField", reflect.TypeOf(""))
		resolver.ResolveEnvVar("MISSING_TEST", "config.missing", reflect.TypeOf(""))

		log := resolver.GetResolutionLog()
		assert.Len(t, log, 3)

		// Check first entry (existing env var)
		assert.Equal(t, "LOG_TEST", log[0].VariableName)
		assert.True(t, log[0].WasSet)
		assert.Equal(t, "debug", log[0].RawValue)
		assert.NoError(t, log[0].Error)

		// Check second entry (default value)
		assert.Equal(t, "DEFAULT_TEST", log[1].VariableName)
		assert.False(t, log[1].WasSet)
		assert.True(t, log[1].UsedDefault)
		assert.Equal(t, "default-val", log[1].DefaultValue)
		assert.NoError(t, log[1].Error)

		// Check third entry (missing with error)
		assert.Equal(t, "MISSING_TEST", log[2].VariableName)
		assert.False(t, log[2].WasSet)
		assert.False(t, log[2].UsedDefault)
		assert.Error(t, log[2].Error)
	})

	t.Run("debug report generation", func(t *testing.T) {
		resolver.ClearResolutionLog()

		os.Setenv("REPORT_TEST", "test-value")
		defer os.Unsetenv("REPORT_TEST")

		resolver.ResolveEnvVar("REPORT_TEST", "config.test", reflect.TypeOf(""))

		report := resolver.GetDebugReport()
		assert.Contains(t, report, "Environment Variable Resolution Debug Report")
		assert.Contains(t, report, "REPORT_TEST")
		assert.Contains(t, report, "config.test")
		assert.Contains(t, report, "test-value")
		assert.Contains(t, report, "Summary:")
	})
}

// Helper functions

func intPtr(i int) *int {
	return &i
}
