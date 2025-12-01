package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ConfigValidator defines the interface for validating service configurations
type ConfigValidator interface {
	// Validate validates a complete configuration object against its schema
	Validate(config interface{}) error

	// ValidateField validates a single field value against its validation rules
	ValidateField(fieldPath, value string, rules FieldValidationRules) error

	// AddValidationRule adds a custom validation rule
	AddValidationRule(name string, rule ValidationRule)

	// GetValidationErrors returns detailed validation errors for a configuration
	GetValidationErrors(config interface{}) []ConfigValidationError
}

// FieldValidationRules defines validation rules for a configuration field
type FieldValidationRules struct {
	// Required indicates if the field is mandatory
	Required bool

	// Type specifies the expected Go type for the field
	Type reflect.Type

	// Pattern specifies a regex pattern the field value must match
	Pattern *regexp.Regexp

	// MinLength specifies minimum string length (for string fields)
	MinLength *int

	// MaxLength specifies maximum string length (for string fields)
	MaxLength *int

	// AllowedValues specifies a list of allowed values
	AllowedValues []string

	// CustomRules specifies custom validation rules
	CustomRules []string

	// EnvironmentVariable specifies the environment variable name for this field
	EnvironmentVariable string

	// DefaultValue specifies the default value if not provided
	DefaultValue interface{}

	// ExampleValue provides an example of a valid value for error messages
	ExampleValue string

	// Description provides human-readable description for error messages
	Description string
}

// ValidationRule defines a custom validation rule function
type ValidationRule func(value interface{}, config interface{}) error

// ConfigValidationError represents a detailed configuration validation error
type ConfigValidationError struct {
	// FieldPath is the path to the invalid field (e.g., "database.connection.host")
	FieldPath string

	// FieldName is the simple field name
	FieldName string

	// Value is the invalid value that was provided
	Value interface{}

	// ExpectedType is the expected type for the field
	ExpectedType string

	// ValidationRule is the rule that failed
	ValidationRule string

	// Message is the human-readable error message
	Message string

	// Suggestions provides actionable suggestions to fix the error
	Suggestions []string

	// EnvironmentVariable is the env var name if this field is env-based
	EnvironmentVariable string

	// ExampleValue provides an example of a valid value
	ExampleValue string
}

// Error implements the error interface for ConfigValidationError
func (e *ConfigValidationError) Error() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Configuration validation failed for field '%s': %s",
		e.FieldPath, e.Message))

	if e.Value != nil {
		builder.WriteString(fmt.Sprintf(" (provided value: %v)", e.Value))
	}

	if e.ExpectedType != "" {
		builder.WriteString(fmt.Sprintf(" (expected type: %s)", e.ExpectedType))
	}

	if e.EnvironmentVariable != "" {
		builder.WriteString(fmt.Sprintf(" (environment variable: %s)", e.EnvironmentVariable))
	}

	if e.ExampleValue != "" {
		builder.WriteString(fmt.Sprintf(" (example: %s)", e.ExampleValue))
	}

	if len(e.Suggestions) > 0 {
		builder.WriteString("\nSuggestions:")
		for _, suggestion := range e.Suggestions {
			builder.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	}

	return builder.String()
}

// ConfigValidationErrors represents multiple configuration validation errors
type ConfigValidationErrors struct {
	Errors []ConfigValidationError
}

// Error implements the error interface for ConfigValidationErrors
func (e *ConfigValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no configuration validation errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Configuration validation failed with %d errors:\n", len(e.Errors)))

	for i, err := range e.Errors {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}

	return builder.String()
}

// HasErrors returns true if there are validation errors
func (e *ConfigValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// AddError adds a validation error
func (e *ConfigValidationErrors) AddError(err ConfigValidationError) {
	e.Errors = append(e.Errors, err)
}

// defaultConfigValidator implements the ConfigValidator interface
type defaultConfigValidator struct {
	// customRules stores custom validation rules
	customRules map[string]ValidationRule

	// fieldRules stores validation rules for specific fields
	fieldRules map[string]FieldValidationRules

	// enableEnvDebugging controls environment variable resolution debugging
	enableEnvDebugging bool
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator(enableEnvDebugging bool) ConfigValidator {
	return &defaultConfigValidator{
		customRules:        make(map[string]ValidationRule),
		fieldRules:         make(map[string]FieldValidationRules),
		enableEnvDebugging: enableEnvDebugging,
	}
}

// Validate validates a complete configuration object
func (v *defaultConfigValidator) Validate(config interface{}) error {
	errors := v.GetValidationErrors(config)
	if len(errors) == 0 {
		return nil
	}

	return &ConfigValidationErrors{Errors: errors}
}

// ValidateField validates a single field value
func (v *defaultConfigValidator) ValidateField(fieldPath, value string, rules FieldValidationRules) error {
	validationErr := &ConfigValidationError{
		FieldPath:           fieldPath,
		FieldName:           getFieldNameFromPath(fieldPath),
		Value:               value,
		EnvironmentVariable: rules.EnvironmentVariable,
		ExpectedType:        getTypeString(rules.Type),
	}

	// Check if required field is missing
	if rules.Required && value == "" {
		validationErr.ValidationRule = "required"
		validationErr.Message = "field is required but not provided"
		validationErr.Suggestions = []string{
			"Provide a value for this field",
		}
		if rules.EnvironmentVariable != "" {
			validationErr.Suggestions = append(validationErr.Suggestions,
				fmt.Sprintf("Set environment variable %s", rules.EnvironmentVariable))
		}
		if rules.DefaultValue != nil {
			validationErr.Suggestions = append(validationErr.Suggestions,
				fmt.Sprintf("Or remove the field to use default value: %v", rules.DefaultValue))
		}
		return validationErr
	}

	// Skip validation for empty optional fields
	if value == "" && !rules.Required {
		return nil
	}

	// Validate pattern
	if rules.Pattern != nil && !rules.Pattern.MatchString(value) {
		validationErr.ValidationRule = "pattern"
		validationErr.Message = "value does not match required pattern"
		validationErr.Suggestions = []string{
			fmt.Sprintf("Value must match pattern: %s", rules.Pattern.String()),
		}
		if rules.ExampleValue != "" {
			validationErr.ExampleValue = rules.ExampleValue
		}
		return validationErr
	}

	// Validate length constraints
	if rules.MinLength != nil && len(value) < *rules.MinLength {
		validationErr.ValidationRule = "minLength"
		validationErr.Message = fmt.Sprintf("value is too short (minimum length: %d)", *rules.MinLength)
		validationErr.Suggestions = []string{
			fmt.Sprintf("Provide a value with at least %d characters", *rules.MinLength),
		}
		return validationErr
	}

	if rules.MaxLength != nil && len(value) > *rules.MaxLength {
		validationErr.ValidationRule = "maxLength"
		validationErr.Message = fmt.Sprintf("value is too long (maximum length: %d)", *rules.MaxLength)
		validationErr.Suggestions = []string{
			fmt.Sprintf("Provide a value with at most %d characters", *rules.MaxLength),
		}
		return validationErr
	}

	// Validate allowed values
	if len(rules.AllowedValues) > 0 {
		found := false
		for _, allowed := range rules.AllowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			validationErr.ValidationRule = "allowedValues"
			validationErr.Message = "value is not in the list of allowed values"
			validationErr.Suggestions = []string{
				fmt.Sprintf("Use one of: %s", strings.Join(rules.AllowedValues, ", ")),
			}
			return validationErr
		}
	}

	// Apply custom rules
	for _, ruleName := range rules.CustomRules {
		if rule, exists := v.customRules[ruleName]; exists {
			if err := rule(value, nil); err != nil {
				validationErr.ValidationRule = ruleName
				validationErr.Message = err.Error()
				validationErr.Suggestions = []string{
					"Check custom validation rule requirements",
				}
				return validationErr
			}
		}
	}

	return nil
}

// AddValidationRule adds a custom validation rule
func (v *defaultConfigValidator) AddValidationRule(name string, rule ValidationRule) {
	v.customRules[name] = rule
}

// GetValidationErrors returns detailed validation errors for a configuration
func (v *defaultConfigValidator) GetValidationErrors(config interface{}) []ConfigValidationError {
	var errors []ConfigValidationError

	// Use reflection to iterate through configuration fields
	configValue := reflect.ValueOf(config)
	configType := reflect.TypeOf(config)

	// Handle pointers
	if configType.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
		configType = configType.Elem()
	}

	// Only validate struct types
	if configType.Kind() != reflect.Struct {
		errors = append(errors, ConfigValidationError{
			FieldPath:      "root",
			ValidationRule: "type",
			Message:        "configuration must be a struct",
			Suggestions: []string{
				"Provide a struct type for configuration validation",
			},
		})
		return errors
	}

	// Validate each field
	errors = append(errors, v.validateStructFields(configValue, configType, "")...)

	return errors
}

// validateStructFields recursively validates struct fields
func (v *defaultConfigValidator) validateStructFields(value reflect.Value, typ reflect.Type, prefix string) []ConfigValidationError {
	var errors []ConfigValidationError

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldPath := field.Name
		if prefix != "" {
			fieldPath = prefix + "." + field.Name
		}

		// Extract validation rules from struct tags
		rules := v.extractValidationRules(field)

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			errors = append(errors, v.validateStructFields(fieldValue, field.Type, fieldPath)...)
			continue
		}

		// Handle pointers to structs
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			if !fieldValue.IsNil() {
				errors = append(errors, v.validateStructFields(fieldValue.Elem(), field.Type.Elem(), fieldPath)...)
			}
			continue
		}

		// Convert field value to string for validation
		fieldStringValue := v.getFieldValueAsString(fieldValue)

		// Check if this is a zero value for the field type (which indicates "not set")
		isZeroValue := fieldValue.IsZero()

		// For required validation, we need to check the zero value, not the string representation
		if isZeroValue {
			fieldStringValue = "" // Treat zero values as empty strings for validation purposes
		}

		// Validate the field
		if err := v.ValidateField(fieldPath, fieldStringValue, rules); err != nil {
			if configErr, ok := err.(*ConfigValidationError); ok {
				errors = append(errors, *configErr)
			}
		}
	}

	return errors
}

// extractValidationRules extracts validation rules from struct field tags
func (v *defaultConfigValidator) extractValidationRules(field reflect.StructField) FieldValidationRules {
	rules := FieldValidationRules{
		Type: field.Type,
	}

	// Extract from 'validate' tag
	if validateTag := field.Tag.Get("validate"); validateTag != "" {
		v.parseValidateTag(validateTag, &rules)
	}

	// Extract environment variable from 'env' tag
	if envTag := field.Tag.Get("env"); envTag != "" {
		rules.EnvironmentVariable = envTag
	}

	// Extract default value from 'default' tag
	if defaultTag := field.Tag.Get("default"); defaultTag != "" {
		rules.DefaultValue = defaultTag
	}

	// Extract description from 'description' tag
	if descTag := field.Tag.Get("description"); descTag != "" {
		rules.Description = descTag
	}

	return rules
}

// parseValidateTag parses validation rules from the 'validate' struct tag
func (v *defaultConfigValidator) parseValidateTag(tag string, rules *FieldValidationRules) {
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			rules.Required = true
		} else if strings.HasPrefix(part, "pattern=") {
			pattern := strings.TrimPrefix(part, "pattern=")
			if compiled, err := regexp.Compile(pattern); err == nil {
				rules.Pattern = compiled
			}
		} else if strings.HasPrefix(part, "min=") {
			if length := parseIntValue(strings.TrimPrefix(part, "min=")); length != nil {
				rules.MinLength = length
			}
		} else if strings.HasPrefix(part, "max=") {
			if length := parseIntValue(strings.TrimPrefix(part, "max=")); length != nil {
				rules.MaxLength = length
			}
		} else if strings.HasPrefix(part, "oneof=") {
			values := strings.TrimPrefix(part, "oneof=")
			rules.AllowedValues = strings.Split(values, "|")
		} else {
			// Assume it's a custom rule name
			rules.CustomRules = append(rules.CustomRules, part)
		}
	}
}

// Helper functions

func getFieldNameFromPath(path string) string {
	parts := strings.Split(path, ".")
	return parts[len(parts)-1]
}

func getTypeString(typ reflect.Type) string {
	if typ == nil {
		return "unknown"
	}
	return typ.String()
}

func (v *defaultConfigValidator) getFieldValueAsString(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}

	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", value.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", value.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", value.Bool())
	case reflect.Ptr:
		if value.IsNil() {
			return ""
		}
		return v.getFieldValueAsString(value.Elem())
	default:
		return fmt.Sprintf("%v", value.Interface())
	}
}

func parseIntValue(s string) *int {
	var val int
	if n, err := fmt.Sscanf(s, "%d", &val); n == 1 && err == nil {
		return &val
	}
	return nil
}
