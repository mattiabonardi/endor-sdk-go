package validation

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// EnvResolver handles environment variable resolution with debugging support
type EnvResolver struct {
	// enableDebugging controls debug output for environment variable resolution
	enableDebugging bool

	// resolutionLog stores debug information about environment variable resolution
	resolutionLog []EnvResolutionEntry

	// defaultValues stores default values for environment variables
	defaultValues map[string]interface{}
}

// EnvResolutionEntry represents a single environment variable resolution attempt
type EnvResolutionEntry struct {
	// VariableName is the environment variable name
	VariableName string

	// FieldPath is the configuration field path this variable maps to
	FieldPath string

	// RequestedType is the Go type requested for the variable
	RequestedType string

	// RawValue is the raw string value from the environment
	RawValue string

	// ResolvedValue is the converted value
	ResolvedValue interface{}

	// WasSet indicates if the environment variable was actually set
	WasSet bool

	// UsedDefault indicates if a default value was used
	UsedDefault bool

	// DefaultValue is the default value used (if any)
	DefaultValue interface{}

	// Error contains any error that occurred during resolution
	Error error
}

// NewEnvResolver creates a new environment variable resolver
func NewEnvResolver(enableDebugging bool) *EnvResolver {
	return &EnvResolver{
		enableDebugging: enableDebugging,
		resolutionLog:   make([]EnvResolutionEntry, 0),
		defaultValues:   make(map[string]interface{}),
	}
}

// SetDefault sets a default value for an environment variable
func (r *EnvResolver) SetDefault(varName string, defaultValue interface{}) {
	r.defaultValues[varName] = defaultValue
}

// ResolveEnvVar resolves an environment variable to the requested type
func (r *EnvResolver) ResolveEnvVar(varName, fieldPath string, targetType reflect.Type) (interface{}, error) {
	entry := EnvResolutionEntry{
		VariableName:  varName,
		FieldPath:     fieldPath,
		RequestedType: targetType.String(),
	}

	// Get raw value from environment
	rawValue, wasSet := os.LookupEnv(varName)
	entry.RawValue = rawValue
	entry.WasSet = wasSet

	var resolvedValue interface{}
	var err error

	if wasSet {
		// Convert the string value to the requested type
		resolvedValue, err = r.convertStringToType(rawValue, targetType)
		if err != nil {
			// On conversion error, resolved value should be nil
			resolvedValue = nil
		}
		entry.ResolvedValue = resolvedValue
		entry.Error = err
	} else {
		// Check for default value
		if defaultValue, hasDefault := r.defaultValues[varName]; hasDefault {
			resolvedValue = defaultValue
			entry.UsedDefault = true
			entry.DefaultValue = defaultValue
			entry.ResolvedValue = resolvedValue
		} else {
			err = fmt.Errorf("environment variable %s not set and no default provided", varName)
			entry.Error = err
		}
	}

	// Log the resolution attempt
	r.resolutionLog = append(r.resolutionLog, entry)

	// Print debug info if enabled
	if r.enableDebugging {
		r.printDebugInfo(entry)
	}

	return resolvedValue, err
}

// GetResolutionLog returns the complete environment variable resolution log
func (r *EnvResolver) GetResolutionLog() []EnvResolutionEntry {
	return r.resolutionLog
}

// ClearResolutionLog clears the resolution log
func (r *EnvResolver) ClearResolutionLog() {
	r.resolutionLog = make([]EnvResolutionEntry, 0)
}

// GetDebugReport returns a formatted debug report of environment variable resolution
func (r *EnvResolver) GetDebugReport() string {
	var builder strings.Builder

	builder.WriteString("Environment Variable Resolution Debug Report\n")
	builder.WriteString("============================================\n\n")

	if len(r.resolutionLog) == 0 {
		builder.WriteString("No environment variables resolved.\n")
		return builder.String()
	}

	for i, entry := range r.resolutionLog {
		builder.WriteString(fmt.Sprintf("%d. Variable: %s\n", i+1, entry.VariableName))
		builder.WriteString(fmt.Sprintf("   Field Path: %s\n", entry.FieldPath))
		builder.WriteString(fmt.Sprintf("   Requested Type: %s\n", entry.RequestedType))
		builder.WriteString(fmt.Sprintf("   Was Set: %t\n", entry.WasSet))

		if entry.WasSet {
			builder.WriteString(fmt.Sprintf("   Raw Value: %q\n", entry.RawValue))
			builder.WriteString(fmt.Sprintf("   Resolved Value: %v\n", entry.ResolvedValue))
		} else {
			builder.WriteString("   Raw Value: (not set)\n")
		}

		if entry.UsedDefault {
			builder.WriteString(fmt.Sprintf("   Used Default: %v\n", entry.DefaultValue))
		}

		if entry.Error != nil {
			builder.WriteString(fmt.Sprintf("   Error: %s\n", entry.Error.Error()))
		}

		builder.WriteString("\n")
	}

	// Add summary
	successful := 0
	failed := 0
	usedDefaults := 0

	for _, entry := range r.resolutionLog {
		if entry.Error == nil {
			successful++
		} else {
			failed++
		}
		if entry.UsedDefault {
			usedDefaults++
		}
	}

	builder.WriteString(fmt.Sprintf("Summary: %d total, %d successful, %d failed, %d used defaults\n",
		len(r.resolutionLog), successful, failed, usedDefaults))

	return builder.String()
}

// printDebugInfo prints debug information for a single resolution entry
func (r *EnvResolver) printDebugInfo(entry EnvResolutionEntry) {
	fmt.Printf("[ENV-DEBUG] %s -> %s (%s): ", entry.VariableName, entry.FieldPath, entry.RequestedType)

	if entry.WasSet {
		fmt.Printf("SET=%q", entry.RawValue)
		if entry.Error == nil {
			fmt.Printf(" RESOLVED=%v", entry.ResolvedValue)
		} else {
			fmt.Printf(" ERROR=%s", entry.Error.Error())
		}
	} else if entry.UsedDefault {
		fmt.Printf("NOT_SET DEFAULT=%v", entry.DefaultValue)
	} else {
		fmt.Printf("NOT_SET NO_DEFAULT ERROR=%s", entry.Error.Error())
	}

	fmt.Println()
}

// convertStringToType converts a string value to the target Go type
func (r *EnvResolver) convertStringToType(value string, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.String:
		return value, nil

	case reflect.Int:
		return strconv.Atoi(value)

	case reflect.Int8:
		val, err := strconv.ParseInt(value, 10, 8)
		return int8(val), err

	case reflect.Int16:
		val, err := strconv.ParseInt(value, 10, 16)
		return int16(val), err

	case reflect.Int32:
		val, err := strconv.ParseInt(value, 10, 32)
		return int32(val), err

	case reflect.Int64:
		return strconv.ParseInt(value, 10, 64)

	case reflect.Uint:
		val, err := strconv.ParseUint(value, 10, 0)
		return uint(val), err

	case reflect.Uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		return uint8(val), err

	case reflect.Uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		return uint16(val), err

	case reflect.Uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		return uint32(val), err

	case reflect.Uint64:
		return strconv.ParseUint(value, 10, 64)

	case reflect.Float32:
		val, err := strconv.ParseFloat(value, 32)
		return float32(val), err

	case reflect.Float64:
		return strconv.ParseFloat(value, 64)

	case reflect.Bool:
		return strconv.ParseBool(value)

	case reflect.Slice:
		// Handle string slices (comma-separated values)
		if targetType.Elem().Kind() == reflect.String {
			if value == "" {
				return []string{}, nil
			}
			return strings.Split(value, ","), nil
		}
		return nil, fmt.Errorf("unsupported slice type: %s", targetType.String())

	case reflect.Ptr:
		// Handle pointers by converting to the underlying type
		if value == "" {
			return nil, nil
		}
		innerValue, err := r.convertStringToType(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		// Create a new pointer to the converted value
		ptrValue := reflect.New(targetType.Elem())
		ptrValue.Elem().Set(reflect.ValueOf(innerValue))
		return ptrValue.Interface(), nil

	default:
		return nil, fmt.Errorf("unsupported type for environment variable conversion: %s", targetType.String())
	}
}
