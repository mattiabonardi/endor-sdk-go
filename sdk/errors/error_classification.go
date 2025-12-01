package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorClassification defines the classification system for framework errors
type ErrorClassification struct {
	// Category is the high-level category of the error
	Category ErrorCategory

	// Subcategory provides more specific classification
	Subcategory ErrorSubcategory

	// Severity indicates how critical the error is
	Severity ErrorSeverity

	// Recoverability indicates if the error can be recovered from
	Recoverability ErrorRecoverability

	// UserImpact describes the impact on end users
	UserImpact UserImpactLevel

	// Tags provide additional metadata for classification
	Tags []string
}

// ErrorCategory represents the main category of errors
type ErrorCategory string

const (
	CategoryDependencyInjection ErrorCategory = "DependencyInjection"
	CategoryConfiguration       ErrorCategory = "Configuration"
	CategoryValidation          ErrorCategory = "Validation"
	CategoryHTTP                ErrorCategory = "HTTP"
	CategoryDatabase            ErrorCategory = "Database"
	CategoryMiddleware          ErrorCategory = "Middleware"
	CategoryLifecycle           ErrorCategory = "Lifecycle"
	CategorySerialization       ErrorCategory = "Serialization"
	CategoryAuthentication      ErrorCategory = "Authentication"
	CategoryAuthorization       ErrorCategory = "Authorization"
	CategoryBusinessLogic       ErrorCategory = "BusinessLogic"
	CategoryInfrastructure      ErrorCategory = "Infrastructure"
	CategorySecurity            ErrorCategory = "Security"
	CategoryPerformance         ErrorCategory = "Performance"
	CategoryConcurrency         ErrorCategory = "Concurrency"
	CategoryExternal            ErrorCategory = "External"
	CategoryFramework           ErrorCategory = "Framework"
	CategoryUnknown             ErrorCategory = "Unknown"
)

// ErrorSubcategory provides more specific error classification
type ErrorSubcategory string

const (
	// Dependency Injection subcategories
	SubcategoryDIRegistration       ErrorSubcategory = "DI.Registration"
	SubcategoryDIResolution         ErrorSubcategory = "DI.Resolution"
	SubcategoryDICircularDependency ErrorSubcategory = "DI.CircularDependency"
	SubcategoryDIConstructor        ErrorSubcategory = "DI.Constructor"

	// Configuration subcategories
	SubcategoryConfigMissing     ErrorSubcategory = "Config.Missing"
	SubcategoryConfigInvalid     ErrorSubcategory = "Config.Invalid"
	SubcategoryConfigParsing     ErrorSubcategory = "Config.Parsing"
	SubcategoryConfigEnvironment ErrorSubcategory = "Config.Environment"
	SubcategoryConfigSchema      ErrorSubcategory = "Config.Schema"

	// Validation subcategories
	SubcategoryValidationRequired   ErrorSubcategory = "Validation.Required"
	SubcategoryValidationFormat     ErrorSubcategory = "Validation.Format"
	SubcategoryValidationRange      ErrorSubcategory = "Validation.Range"
	SubcategoryValidationConstraint ErrorSubcategory = "Validation.Constraint"
	SubcategoryValidationCustom     ErrorSubcategory = "Validation.Custom"

	// HTTP subcategories
	SubcategoryHTTPRouting       ErrorSubcategory = "HTTP.Routing"
	SubcategoryHTTPParsing       ErrorSubcategory = "HTTP.Parsing"
	SubcategoryHTTPSerialization ErrorSubcategory = "HTTP.Serialization"
	SubcategoryHTTPTimeout       ErrorSubcategory = "HTTP.Timeout"
	SubcategoryHTTPClientError   ErrorSubcategory = "HTTP.ClientError"
	SubcategoryHTTPServerError   ErrorSubcategory = "HTTP.ServerError"

	// Database subcategories
	SubcategoryDatabaseConnection  ErrorSubcategory = "Database.Connection"
	SubcategoryDatabaseQuery       ErrorSubcategory = "Database.Query"
	SubcategoryDatabaseTransaction ErrorSubcategory = "Database.Transaction"
	SubcategoryDatabaseConstraint  ErrorSubcategory = "Database.Constraint"
	SubcategoryDatabaseTimeout     ErrorSubcategory = "Database.Timeout"

	// Security subcategories
	SubcategorySecurityAuthentication ErrorSubcategory = "Security.Authentication"
	SubcategorySecurityAuthorization  ErrorSubcategory = "Security.Authorization"
	SubcategorySecurityEncryption     ErrorSubcategory = "Security.Encryption"
	SubcategorySecurityValidation     ErrorSubcategory = "Security.Validation"

	// Generic subcategory
	SubcategoryGeneric ErrorSubcategory = "Generic"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "Critical" // System failure, immediate action required
	SeverityHigh     ErrorSeverity = "High"     // Major functionality affected
	SeverityMedium   ErrorSeverity = "Medium"   // Some functionality affected
	SeverityLow      ErrorSeverity = "Low"      // Minor issues, degraded performance
	SeverityInfo     ErrorSeverity = "Info"     // Informational, no action needed
)

// ErrorRecoverability indicates whether an error can be recovered from
type ErrorRecoverability string

const (
	RecoverabilityNone      ErrorRecoverability = "None"      // Cannot be recovered, requires restart
	RecoverabilityManual    ErrorRecoverability = "Manual"    // Requires manual intervention
	RecoverabilityAutomatic ErrorRecoverability = "Automatic" // Can be automatically retried
	RecoverabilityGraceful  ErrorRecoverability = "Graceful"  // Can gracefully degrade
	RecoverabilityTransient ErrorRecoverability = "Transient" // Temporary issue, will resolve
)

// UserImpactLevel describes the impact on end users
type UserImpactLevel string

const (
	UserImpactBlocking    UserImpactLevel = "Blocking"    // User cannot proceed
	UserImpactDegraded    UserImpactLevel = "Degraded"    // User experience is degraded
	UserImpactInvisible   UserImpactLevel = "Invisible"   // User is not affected
	UserImpactPerformance UserImpactLevel = "Performance" // Slower performance
)

// ClassifiedError represents an error with classification information
type ClassifiedError struct {
	// OriginalError is the underlying error
	OriginalError error

	// Classification contains the error classification
	Classification ErrorClassification

	// Timestamp when the error was classified
	Timestamp time.Time

	// Context contains additional context about the error
	Context map[string]interface{}

	// ErrorID is a unique identifier for this error instance
	ErrorID string

	// CorrelationID links related errors
	CorrelationID string
}

// Error implements the error interface
func (ce *ClassifiedError) Error() string {
	return fmt.Sprintf("[%s:%s] [%s] %s",
		ce.Classification.Category,
		ce.Classification.Subcategory,
		ce.Classification.Severity,
		ce.OriginalError.Error())
}

// Unwrap returns the original error for error unwrapping
func (ce *ClassifiedError) Unwrap() error {
	return ce.OriginalError
}

// IsCritical returns true if the error is critical severity
func (ce *ClassifiedError) IsCritical() bool {
	return ce.Classification.Severity == SeverityCritical
}

// IsRecoverable returns true if the error can be recovered from automatically
func (ce *ClassifiedError) IsRecoverable() bool {
	return ce.Classification.Recoverability == RecoverabilityAutomatic ||
		ce.Classification.Recoverability == RecoverabilityGraceful ||
		ce.Classification.Recoverability == RecoverabilityTransient
}

// BlocksUser returns true if the error blocks user operations
func (ce *ClassifiedError) BlocksUser() bool {
	return ce.Classification.UserImpact == UserImpactBlocking
}

// ErrorClassifier provides functionality to classify errors automatically
type ErrorClassifier struct {
	// Rules contains classification rules
	rules []ClassificationRule

	// DefaultClassification is used when no rules match
	defaultClassification ErrorClassification
}

// ClassificationRule defines a rule for classifying errors
type ClassificationRule struct {
	// Name is a human-readable name for the rule
	Name string

	// Condition determines if this rule applies to an error
	Condition func(error, *ContextualError) bool

	// Classification to apply if the condition matches
	Classification ErrorClassification

	// Priority determines rule evaluation order (higher = evaluated first)
	Priority int
}

// NewErrorClassifier creates a new error classifier with default rules
func NewErrorClassifier() *ErrorClassifier {
	classifier := &ErrorClassifier{
		rules: make([]ClassificationRule, 0),
		defaultClassification: ErrorClassification{
			Category:       CategoryUnknown,
			Subcategory:    SubcategoryGeneric,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactDegraded,
			Tags:           []string{"unclassified"},
		},
	}

	// Add default classification rules
	classifier.addDefaultRules()

	return classifier
}

// addDefaultRules adds the built-in classification rules
func (ec *ErrorClassifier) addDefaultRules() {
	// Dependency Injection errors
	ec.AddRule(ClassificationRule{
		Name:     "DI Registration Error",
		Priority: 100,
		Condition: func(err error, ctxErr *ContextualError) bool {
			if ctxErr != nil && ctxErr.OperationContext.OperationType == "DependencyInjection" {
				return strings.Contains(strings.ToLower(err.Error()), "register") ||
					strings.Contains(strings.ToLower(err.Error()), "registration")
			}
			return strings.Contains(strings.ToLower(err.Error()), "dependency") &&
				strings.Contains(strings.ToLower(err.Error()), "register")
		},
		Classification: ErrorClassification{
			Category:       CategoryDependencyInjection,
			Subcategory:    SubcategoryDIRegistration,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"dependency-injection", "registration"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "DI Resolution Error",
		Priority: 100,
		Condition: func(err error, ctxErr *ContextualError) bool {
			if ctxErr != nil && ctxErr.OperationContext.OperationType == "DependencyInjection" {
				return strings.Contains(strings.ToLower(err.Error()), "resolve") ||
					strings.Contains(strings.ToLower(err.Error()), "not found")
			}
			return strings.Contains(strings.ToLower(err.Error()), "dependency") &&
				(strings.Contains(strings.ToLower(err.Error()), "resolve") ||
					strings.Contains(strings.ToLower(err.Error()), "not found"))
		},
		Classification: ErrorClassification{
			Category:       CategoryDependencyInjection,
			Subcategory:    SubcategoryDIResolution,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"dependency-injection", "resolution"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "Circular Dependency Error",
		Priority: 110,
		Condition: func(err error, ctxErr *ContextualError) bool {
			return strings.Contains(strings.ToLower(err.Error()), "circular")
		},
		Classification: ErrorClassification{
			Category:       CategoryDependencyInjection,
			Subcategory:    SubcategoryDICircularDependency,
			Severity:       SeverityCritical,
			Recoverability: RecoverabilityNone,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"dependency-injection", "circular", "design-issue"},
		},
	})

	// Configuration errors
	ec.AddRule(ClassificationRule{
		Name:     "Missing Configuration",
		Priority: 90,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return strings.Contains(errMsg, "config") &&
				(strings.Contains(errMsg, "missing") || strings.Contains(errMsg, "not found") ||
					strings.Contains(errMsg, "required"))
		},
		Classification: ErrorClassification{
			Category:       CategoryConfiguration,
			Subcategory:    SubcategoryConfigMissing,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"configuration", "missing"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "Invalid Configuration",
		Priority: 90,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return strings.Contains(errMsg, "config") &&
				(strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "malformed"))
		},
		Classification: ErrorClassification{
			Category:       CategoryConfiguration,
			Subcategory:    SubcategoryConfigInvalid,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"configuration", "invalid"},
		},
	})

	// Validation errors
	ec.AddRule(ClassificationRule{
		Name:     "Validation Error",
		Priority: 80,
		Condition: func(err error, ctxErr *ContextualError) bool {
			if ctxErr != nil && ctxErr.OperationContext.OperationType == "Validation" {
				return true
			}
			return strings.Contains(strings.ToLower(err.Error()), "validat")
		},
		Classification: ErrorClassification{
			Category:       CategoryValidation,
			Subcategory:    SubcategoryValidationConstraint,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityAutomatic,
			UserImpact:     UserImpactDegraded,
			Tags:           []string{"validation", "input"},
		},
	})

	// HTTP errors
	ec.AddRule(ClassificationRule{
		Name:     "HTTP 4xx Client Error",
		Priority: 70,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := err.Error()
			return strings.Contains(errMsg, "400") || strings.Contains(errMsg, "404") ||
				strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") ||
				strings.Contains(errMsg, "bad request") || strings.Contains(errMsg, "not found") ||
				strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden")
		},
		Classification: ErrorClassification{
			Category:       CategoryHTTP,
			Subcategory:    SubcategoryHTTPClientError,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityAutomatic,
			UserImpact:     UserImpactDegraded,
			Tags:           []string{"http", "client-error"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "HTTP 5xx Server Error",
		Priority: 70,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := err.Error()
			return strings.Contains(errMsg, "500") || strings.Contains(errMsg, "502") ||
				strings.Contains(errMsg, "503") || strings.Contains(errMsg, "504") ||
				strings.Contains(errMsg, "internal server error") ||
				strings.Contains(errMsg, "service unavailable")
		},
		Classification: ErrorClassification{
			Category:       CategoryHTTP,
			Subcategory:    SubcategoryHTTPServerError,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityTransient,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"http", "server-error"},
		},
	})

	// Database errors
	ec.AddRule(ClassificationRule{
		Name:     "Database Connection Error",
		Priority: 95,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return strings.Contains(errMsg, "connection") &&
				(strings.Contains(errMsg, "database") || strings.Contains(errMsg, "mongo") ||
					strings.Contains(errMsg, "refused") || strings.Contains(errMsg, "timeout"))
		},
		Classification: ErrorClassification{
			Category:       CategoryDatabase,
			Subcategory:    SubcategoryDatabaseConnection,
			Severity:       SeverityCritical,
			Recoverability: RecoverabilityTransient,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"database", "connection", "infrastructure"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "Database Query Error",
		Priority: 85,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return (strings.Contains(errMsg, "database") || strings.Contains(errMsg, "mongo") ||
				strings.Contains(errMsg, "query") || strings.Contains(errMsg, "sql")) &&
				(strings.Contains(errMsg, "syntax") || strings.Contains(errMsg, "invalid"))
		},
		Classification: ErrorClassification{
			Category:       CategoryDatabase,
			Subcategory:    SubcategoryDatabaseQuery,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactDegraded,
			Tags:           []string{"database", "query", "syntax"},
		},
	})

	// Security errors
	ec.AddRule(ClassificationRule{
		Name:     "Authentication Error",
		Priority: 95,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return strings.Contains(errMsg, "auth") &&
				(strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "failed") ||
					strings.Contains(errMsg, "unauthorized"))
		},
		Classification: ErrorClassification{
			Category:       CategoryAuthentication,
			Subcategory:    SubcategorySecurityAuthentication,
			Severity:       SeverityHigh,
			Recoverability: RecoverabilityAutomatic,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"security", "authentication"},
		},
	})

	ec.AddRule(ClassificationRule{
		Name:     "Authorization Error",
		Priority: 95,
		Condition: func(err error, ctxErr *ContextualError) bool {
			errMsg := strings.ToLower(err.Error())
			return strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "access denied") ||
				(strings.Contains(errMsg, "permission") && strings.Contains(errMsg, "denied"))
		},
		Classification: ErrorClassification{
			Category:       CategoryAuthorization,
			Subcategory:    SubcategorySecurityAuthorization,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityManual,
			UserImpact:     UserImpactBlocking,
			Tags:           []string{"security", "authorization", "permissions"},
		},
	})

	// Performance errors
	ec.AddRule(ClassificationRule{
		Name:     "Timeout Error",
		Priority: 75,
		Condition: func(err error, ctxErr *ContextualError) bool {
			return strings.Contains(strings.ToLower(err.Error()), "timeout")
		},
		Classification: ErrorClassification{
			Category:       CategoryPerformance,
			Subcategory:    SubcategoryHTTPTimeout,
			Severity:       SeverityMedium,
			Recoverability: RecoverabilityTransient,
			UserImpact:     UserImpactPerformance,
			Tags:           []string{"performance", "timeout"},
		},
	})
}

// AddRule adds a new classification rule
func (ec *ErrorClassifier) AddRule(rule ClassificationRule) {
	ec.rules = append(ec.rules, rule)

	// Sort rules by priority (highest first)
	for i := len(ec.rules) - 1; i > 0; i-- {
		if ec.rules[i].Priority > ec.rules[i-1].Priority {
			ec.rules[i], ec.rules[i-1] = ec.rules[i-1], ec.rules[i]
		} else {
			break
		}
	}
}

// Classify classifies an error and returns a ClassifiedError
func (ec *ErrorClassifier) Classify(err error, contextualErr *ContextualError) *ClassifiedError {
	classification := ec.defaultClassification

	// Try to match against classification rules
	for _, rule := range ec.rules {
		if rule.Condition(err, contextualErr) {
			classification = rule.Classification
			break
		}
	}

	// Enhance classification based on contextual information
	if contextualErr != nil {
		classification = ec.enhanceWithContext(classification, contextualErr)
	}

	return &ClassifiedError{
		OriginalError:  err,
		Classification: classification,
		Timestamp:      time.Now(),
		Context:        extractErrorContext(err, contextualErr),
		ErrorID:        generateErrorID(),
		CorrelationID:  extractCorrelationID(contextualErr),
	}
}

// enhanceWithContext enhances classification based on contextual information
func (ec *ErrorClassifier) enhanceWithContext(classification ErrorClassification, ctxErr *ContextualError) ErrorClassification {
	// Enhance category based on operation type
	switch ctxErr.OperationContext.OperationType {
	case OperationTypeHTTP:
		if classification.Category == CategoryUnknown {
			classification.Category = CategoryHTTP
		}
	case OperationTypeDatabase:
		if classification.Category == CategoryUnknown {
			classification.Category = CategoryDatabase
		}
	case OperationTypeValidation:
		if classification.Category == CategoryUnknown {
			classification.Category = CategoryValidation
		}
	case OperationTypeMiddleware:
		if classification.Category == CategoryUnknown {
			classification.Category = CategoryMiddleware
		}
	}

	// Enhance severity based on operation phase
	switch ctxErr.OperationContext.Phase {
	case PhaseInitialization:
		// Initialization errors are generally more critical
		if classification.Severity == SeverityMedium {
			classification.Severity = SeverityHigh
		}
	case PhaseCleanup:
		// Cleanup errors are generally less critical
		if classification.Severity == SeverityHigh {
			classification.Severity = SeverityMedium
		}
	}

	// Add contextual tags
	if ctxErr.ServiceContext.ServiceType != "" {
		classification.Tags = append(classification.Tags,
			fmt.Sprintf("service-type-%s", strings.ToLower(string(ctxErr.ServiceContext.ServiceType))))
	}

	if ctxErr.ServiceContext.ResourceType != "" {
		classification.Tags = append(classification.Tags,
			fmt.Sprintf("resource-%s", strings.ToLower(ctxErr.ServiceContext.ResourceType)))
	}

	return classification
} // ErrorReporter provides utilities for reporting and handling classified errors
type ErrorReporter struct {
	classifier *ErrorClassifier
	handlers   map[ErrorSeverity][]ErrorHandler
}

// ErrorHandler is a function that handles classified errors
type ErrorHandler func(*ClassifiedError) error

// NewErrorReporter creates a new error reporter
func NewErrorReporter() *ErrorReporter {
	return &ErrorReporter{
		classifier: NewErrorClassifier(),
		handlers:   make(map[ErrorSeverity][]ErrorHandler),
	}
}

// RegisterHandler registers an error handler for a specific severity level
func (er *ErrorReporter) RegisterHandler(severity ErrorSeverity, handler ErrorHandler) {
	if er.handlers[severity] == nil {
		er.handlers[severity] = make([]ErrorHandler, 0)
	}
	er.handlers[severity] = append(er.handlers[severity], handler)
}

// ReportError classifies and reports an error
func (er *ErrorReporter) ReportError(err error, contextualErr *ContextualError) *ClassifiedError {
	// Classify the error
	classifiedErr := er.classifier.Classify(err, contextualErr)

	// Execute handlers for this severity level
	if handlers, exists := er.handlers[classifiedErr.Classification.Severity]; exists {
		for _, handler := range handlers {
			if handlerErr := handler(classifiedErr); handlerErr != nil {
				// Log handler error but don't fail the original reporting
				fmt.Printf("Error handler failed: %v\n", handlerErr)
			}
		}
	}

	return classifiedErr
}

// Helper functions

// extractErrorContext extracts relevant context from an error
func extractErrorContext(err error, ctxErr *ContextualError) map[string]interface{} {
	context := make(map[string]interface{})

	context["error_type"] = fmt.Sprintf("%T", err)
	context["error_message"] = err.Error()

	if ctxErr != nil {
		context["service_name"] = ctxErr.ServiceContext.ServiceName
		context["service_type"] = string(ctxErr.ServiceContext.ServiceType)
		context["operation"] = ctxErr.OperationContext.Operation
		context["operation_type"] = string(ctxErr.OperationContext.OperationType)
		context["phase"] = string(ctxErr.OperationContext.Phase)
		context["request_id"] = ctxErr.RequestID

		if ctxErr.OperationContext.HTTPMethod != "" {
			context["http_method"] = ctxErr.OperationContext.HTTPMethod
			context["http_path"] = ctxErr.OperationContext.HTTPPath
		}

		if ctxErr.OperationContext.Duration > 0 {
			context["duration_ms"] = ctxErr.OperationContext.Duration.Milliseconds()
		}
	}

	return context
}

// extractCorrelationID extracts correlation ID from contextual error
func extractCorrelationID(ctxErr *ContextualError) string {
	if ctxErr != nil && ctxErr.RequestID != "" {
		return ctxErr.RequestID
	}
	return ""
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	// In a real implementation, this would use a proper UUID library
	// For now, use timestamp-based ID
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// GetErrorStatistics analyzes a collection of classified errors and returns statistics
func GetErrorStatistics(errors []*ClassifiedError) ErrorStatistics {
	stats := ErrorStatistics{
		TotalErrors:      len(errors),
		ByCategory:       make(map[ErrorCategory]int),
		BySubcategory:    make(map[ErrorSubcategory]int),
		BySeverity:       make(map[ErrorSeverity]int),
		ByRecoverability: make(map[ErrorRecoverability]int),
		ByUserImpact:     make(map[UserImpactLevel]int),
		MostFrequentTags: make(map[string]int),
		TimeRange:        TimeRange{},
	}

	if len(errors) == 0 {
		return stats
	}

	// Initialize time range
	stats.TimeRange.Start = errors[0].Timestamp
	stats.TimeRange.End = errors[0].Timestamp

	// Analyze each error
	for _, err := range errors {
		// Update counts
		stats.ByCategory[err.Classification.Category]++
		stats.BySubcategory[err.Classification.Subcategory]++
		stats.BySeverity[err.Classification.Severity]++
		stats.ByRecoverability[err.Classification.Recoverability]++
		stats.ByUserImpact[err.Classification.UserImpact]++

		// Count tags
		for _, tag := range err.Classification.Tags {
			stats.MostFrequentTags[tag]++
		}

		// Update time range
		if err.Timestamp.Before(stats.TimeRange.Start) {
			stats.TimeRange.Start = err.Timestamp
		}
		if err.Timestamp.After(stats.TimeRange.End) {
			stats.TimeRange.End = err.Timestamp
		}

		// Count critical and recoverable errors
		if err.IsCritical() {
			stats.CriticalErrorCount++
		}
		if err.IsRecoverable() {
			stats.RecoverableErrorCount++
		}
		if err.BlocksUser() {
			stats.UserBlockingErrorCount++
		}
	}

	return stats
}

// ErrorStatistics contains statistics about a collection of errors
type ErrorStatistics struct {
	TotalErrors            int
	CriticalErrorCount     int
	RecoverableErrorCount  int
	UserBlockingErrorCount int
	ByCategory             map[ErrorCategory]int
	BySubcategory          map[ErrorSubcategory]int
	BySeverity             map[ErrorSeverity]int
	ByRecoverability       map[ErrorRecoverability]int
	ByUserImpact           map[UserImpactLevel]int
	MostFrequentTags       map[string]int
	TimeRange              TimeRange
}

// TimeRange represents a time range for error analysis
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Global error reporter instance
var globalErrorReporter *ErrorReporter

// GetGlobalErrorReporter returns the global error reporter instance
func GetGlobalErrorReporter() *ErrorReporter {
	if globalErrorReporter == nil {
		globalErrorReporter = NewErrorReporter()
	}
	return globalErrorReporter
}

// ClassifyAndReport is a convenience function to classify and report an error
func ClassifyAndReport(err error, contextualErr *ContextualError) *ClassifiedError {
	return GetGlobalErrorReporter().ReportError(err, contextualErr)
}
