package errors

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorCategory_Constants(t *testing.T) {
	// Test that all error categories are properly defined
	categories := []ErrorCategory{
		CategoryDependencyInjection,
		CategoryConfiguration,
		CategoryValidation,
		CategoryHTTP,
		CategoryDatabase,
		CategoryMiddleware,
		CategoryLifecycle,
		CategorySerialization,
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryBusinessLogic,
		CategoryInfrastructure,
		CategorySecurity,
		CategoryPerformance,
		CategoryConcurrency,
		CategoryExternal,
		CategoryFramework,
		CategoryUnknown,
	}

	for _, category := range categories {
		assert.NotEmpty(t, string(category), "Category should not be empty")
	}
}

func TestErrorSeverity_Constants(t *testing.T) {
	// Test that all severity levels are properly defined
	severities := []ErrorSeverity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}

	for _, severity := range severities {
		assert.NotEmpty(t, string(severity), "Severity should not be empty")
	}
}

func TestClassifiedError_Error(t *testing.T) {
	originalErr := fmt.Errorf("test error message")
	classifiedErr := &ClassifiedError{
		OriginalError: originalErr,
		Classification: ErrorClassification{
			Category:    CategoryDependencyInjection,
			Subcategory: SubcategoryDIResolution,
			Severity:    SeverityHigh,
		},
		Timestamp: time.Now(),
	}

	expected := "[DependencyInjection:DI.Resolution] [High] test error message"
	assert.Equal(t, expected, classifiedErr.Error())
}

func TestClassifiedError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	classifiedErr := &ClassifiedError{
		OriginalError: originalErr,
		Classification: ErrorClassification{
			Category:    CategoryValidation,
			Subcategory: SubcategoryValidationRequired,
			Severity:    SeverityMedium,
		},
	}

	assert.Equal(t, originalErr, classifiedErr.Unwrap())
}

func TestClassifiedError_IsCritical(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		expected bool
	}{
		{
			name:     "critical severity",
			severity: SeverityCritical,
			expected: true,
		},
		{
			name:     "high severity",
			severity: SeverityHigh,
			expected: false,
		},
		{
			name:     "medium severity",
			severity: SeverityMedium,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifiedErr := &ClassifiedError{
				Classification: ErrorClassification{
					Severity: tt.severity,
				},
			}
			assert.Equal(t, tt.expected, classifiedErr.IsCritical())
		})
	}
}

func TestClassifiedError_IsRecoverable(t *testing.T) {
	tests := []struct {
		name           string
		recoverability ErrorRecoverability
		expected       bool
	}{
		{
			name:           "automatic recovery",
			recoverability: RecoverabilityAutomatic,
			expected:       true,
		},
		{
			name:           "graceful recovery",
			recoverability: RecoverabilityGraceful,
			expected:       true,
		},
		{
			name:           "transient recovery",
			recoverability: RecoverabilityTransient,
			expected:       true,
		},
		{
			name:           "manual recovery",
			recoverability: RecoverabilityManual,
			expected:       false,
		},
		{
			name:           "no recovery",
			recoverability: RecoverabilityNone,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifiedErr := &ClassifiedError{
				Classification: ErrorClassification{
					Recoverability: tt.recoverability,
				},
			}
			assert.Equal(t, tt.expected, classifiedErr.IsRecoverable())
		})
	}
}

func TestClassifiedError_BlocksUser(t *testing.T) {
	tests := []struct {
		name       string
		userImpact UserImpactLevel
		expected   bool
	}{
		{
			name:       "blocking impact",
			userImpact: UserImpactBlocking,
			expected:   true,
		},
		{
			name:       "degraded impact",
			userImpact: UserImpactDegraded,
			expected:   false,
		},
		{
			name:       "invisible impact",
			userImpact: UserImpactInvisible,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifiedErr := &ClassifiedError{
				Classification: ErrorClassification{
					UserImpact: tt.userImpact,
				},
			}
			assert.Equal(t, tt.expected, classifiedErr.BlocksUser())
		})
	}
}

func TestNewErrorClassifier(t *testing.T) {
	classifier := NewErrorClassifier()

	assert.NotNil(t, classifier)
	assert.NotEmpty(t, classifier.rules)
	assert.Equal(t, CategoryUnknown, classifier.defaultClassification.Category)

	// Verify rules are sorted by priority
	for i := 1; i < len(classifier.rules); i++ {
		assert.GreaterOrEqual(t, classifier.rules[i-1].Priority, classifier.rules[i].Priority,
			"Rules should be sorted by priority (highest first)")
	}
}

func TestErrorClassifier_Classify_DependencyInjection(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name             string
		errorMsg         string
		contextualErr    *ContextualError
		expectedCategory ErrorCategory
		expectedSubcat   ErrorSubcategory
		expectedSeverity ErrorSeverity
	}{
		{
			name:             "DI registration error",
			errorMsg:         "failed to register dependency UserService",
			expectedCategory: CategoryDependencyInjection,
			expectedSubcat:   SubcategoryDIRegistration,
			expectedSeverity: SeverityHigh,
		},
		{
			name:             "DI resolution error",
			errorMsg:         "dependency not found: UserRepository",
			expectedCategory: CategoryDependencyInjection,
			expectedSubcat:   SubcategoryDIResolution,
			expectedSeverity: SeverityHigh,
		},
		{
			name:             "circular dependency error",
			errorMsg:         "circular dependency detected: A -> B -> A",
			expectedCategory: CategoryDependencyInjection,
			expectedSubcat:   SubcategoryDICircularDependency,
			expectedSeverity: SeverityCritical,
		},
		{
			name:     "DI context with operation type",
			errorMsg: "resolve failed",
			contextualErr: &ContextualError{
				OperationContext: OperationErrorContext{
					OperationType: "DependencyInjection",
				},
			},
			expectedCategory: CategoryDependencyInjection,
			expectedSubcat:   SubcategoryDIResolution,
			expectedSeverity: SeverityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, tt.contextualErr)

			assert.Equal(t, tt.expectedCategory, classified.Classification.Category)
			assert.Equal(t, tt.expectedSubcat, classified.Classification.Subcategory)
			assert.Equal(t, tt.expectedSeverity, classified.Classification.Severity)
			assert.NotZero(t, classified.Timestamp)
			assert.NotEmpty(t, classified.ErrorID)
		})
	}
}

func TestErrorClassifier_Classify_Configuration(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name             string
		errorMsg         string
		expectedCategory ErrorCategory
		expectedSubcat   ErrorSubcategory
	}{
		{
			name:             "missing config error",
			errorMsg:         "config file not found: app.yaml",
			expectedCategory: CategoryConfiguration,
			expectedSubcat:   SubcategoryConfigMissing,
		},
		{
			name:             "invalid config error",
			errorMsg:         "invalid config format",
			expectedCategory: CategoryConfiguration,
			expectedSubcat:   SubcategoryConfigInvalid,
		},
		{
			name:             "required config missing",
			errorMsg:         "required config value missing: DATABASE_URL",
			expectedCategory: CategoryConfiguration,
			expectedSubcat:   SubcategoryConfigMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, nil)

			assert.Equal(t, tt.expectedCategory, classified.Classification.Category)
			assert.Equal(t, tt.expectedSubcat, classified.Classification.Subcategory)
		})
	}
}

func TestErrorClassifier_Classify_HTTP(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name             string
		errorMsg         string
		expectedCategory ErrorCategory
		expectedSubcat   ErrorSubcategory
		expectedSeverity ErrorSeverity
	}{
		{
			name:             "404 not found",
			errorMsg:         "404 not found",
			expectedCategory: CategoryHTTP,
			expectedSubcat:   SubcategoryHTTPClientError,
			expectedSeverity: SeverityMedium,
		},
		{
			name:             "400 bad request",
			errorMsg:         "400 bad request",
			expectedCategory: CategoryHTTP,
			expectedSubcat:   SubcategoryHTTPClientError,
			expectedSeverity: SeverityMedium,
		},
		{
			name:             "500 internal server error",
			errorMsg:         "500 internal server error",
			expectedCategory: CategoryHTTP,
			expectedSubcat:   SubcategoryHTTPServerError,
			expectedSeverity: SeverityHigh,
		},
		{
			name:             "503 service unavailable",
			errorMsg:         "503 service unavailable",
			expectedCategory: CategoryHTTP,
			expectedSubcat:   SubcategoryHTTPServerError,
			expectedSeverity: SeverityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, nil)

			assert.Equal(t, tt.expectedCategory, classified.Classification.Category)
			assert.Equal(t, tt.expectedSubcat, classified.Classification.Subcategory)
			assert.Equal(t, tt.expectedSeverity, classified.Classification.Severity)
		})
	}
}

func TestErrorClassifier_Classify_Database(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name             string
		errorMsg         string
		expectedCategory ErrorCategory
		expectedSubcat   ErrorSubcategory
		expectedSeverity ErrorSeverity
	}{
		{
			name:             "connection refused",
			errorMsg:         "database connection refused",
			expectedCategory: CategoryDatabase,
			expectedSubcat:   SubcategoryDatabaseConnection,
			expectedSeverity: SeverityCritical,
		},
		{
			name:             "mongo connection timeout",
			errorMsg:         "mongo connection timeout",
			expectedCategory: CategoryDatabase,
			expectedSubcat:   SubcategoryDatabaseConnection,
			expectedSeverity: SeverityCritical,
		},
		{
			name:             "query syntax error",
			errorMsg:         "SQL syntax error in query",
			expectedCategory: CategoryDatabase,
			expectedSubcat:   SubcategoryDatabaseQuery,
			expectedSeverity: SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, nil)

			assert.Equal(t, tt.expectedCategory, classified.Classification.Category)
			assert.Equal(t, tt.expectedSubcat, classified.Classification.Subcategory)
			assert.Equal(t, tt.expectedSeverity, classified.Classification.Severity)
		})
	}
}

func TestErrorClassifier_Classify_Security(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name             string
		errorMsg         string
		expectedCategory ErrorCategory
		expectedSubcat   ErrorSubcategory
	}{
		{
			name:             "authentication failed",
			errorMsg:         "authentication failed: invalid credentials",
			expectedCategory: CategoryAuthentication,
			expectedSubcat:   SubcategorySecurityAuthentication,
		},
		{
			name:             "unauthorized access",
			errorMsg:         "unauthorized access to resource",
			expectedCategory: CategoryAuthentication,
			expectedSubcat:   SubcategorySecurityAuthentication,
		},
		{
			name:             "access forbidden",
			errorMsg:         "access forbidden: insufficient permissions",
			expectedCategory: CategoryAuthorization,
			expectedSubcat:   SubcategorySecurityAuthorization,
		},
		{
			name:             "permission denied",
			errorMsg:         "permission denied for user",
			expectedCategory: CategoryAuthorization,
			expectedSubcat:   SubcategorySecurityAuthorization,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, nil)

			assert.Equal(t, tt.expectedCategory, classified.Classification.Category)
			assert.Equal(t, tt.expectedSubcat, classified.Classification.Subcategory)
		})
	}
}

func TestErrorClassifier_EnhanceWithContext(t *testing.T) {
	classifier := NewErrorClassifier()

	contextualErr := &ContextualError{
		ServiceContext: ServiceErrorContext{
			ServiceName:  "UserService",
			ServiceType:  ServiceTypeEndor,
			ResourceType: "User",
		},
		OperationContext: OperationErrorContext{
			Operation:     "CreateUser",
			OperationType: OperationTypeHTTP,
			Phase:         PhaseInitialization,
		},
	}

	err := fmt.Errorf("unknown error")
	classified := classifier.Classify(err, contextualErr)

	// Should enhance unknown category to HTTP based on operation type
	assert.Equal(t, CategoryHTTP, classified.Classification.Category)

	// Should include contextual tags
	assert.Contains(t, classified.Classification.Tags, "service-type-endorservice")
	assert.Contains(t, classified.Classification.Tags, "resource-user")

	// Should include context information
	assert.Equal(t, "UserService", classified.Context["service_name"])
	assert.Equal(t, "CreateUser", classified.Context["operation"])
	assert.Equal(t, "HTTP", classified.Context["operation_type"])
}

func TestErrorClassifier_AddRule(t *testing.T) {
	classifier := NewErrorClassifier()
	originalRuleCount := len(classifier.rules)

	// Add a custom rule with high priority
	customRule := ClassificationRule{
		Name:     "Custom High Priority Rule",
		Priority: 200, // Higher than default rules
		Condition: func(err error, ctxErr *ContextualError) bool {
			return err.Error() == "custom error"
		},
		Classification: ErrorClassification{
			Category:    CategoryBusinessLogic,
			Subcategory: SubcategoryGeneric,
			Severity:    SeverityLow,
		},
	}

	classifier.AddRule(customRule)

	// Should have one more rule
	assert.Equal(t, originalRuleCount+1, len(classifier.rules))

	// Should be first rule due to high priority
	assert.Equal(t, "Custom High Priority Rule", classifier.rules[0].Name)
	assert.Equal(t, 200, classifier.rules[0].Priority)

	// Test the custom rule works
	err := fmt.Errorf("custom error")
	classified := classifier.Classify(err, nil)
	assert.Equal(t, CategoryBusinessLogic, classified.Classification.Category)
}

func TestNewErrorReporter(t *testing.T) {
	reporter := NewErrorReporter()

	assert.NotNil(t, reporter)
	assert.NotNil(t, reporter.classifier)
	assert.NotNil(t, reporter.handlers)
}

func TestErrorReporter_RegisterHandler(t *testing.T) {
	reporter := NewErrorReporter()

	handlerCalled := false
	handler := func(err *ClassifiedError) error {
		handlerCalled = true
		return nil
	}

	reporter.RegisterHandler(SeverityHigh, handler)

	// Test that handler is registered
	assert.Contains(t, reporter.handlers, SeverityHigh)
	assert.Len(t, reporter.handlers[SeverityHigh], 1)

	// Test handler is called
	err := fmt.Errorf("test error")
	reporter.ReportError(err, nil)

	// Handler might not be called if the error isn't classified as high severity
	// Let's force a high severity error
	contextualErr := &ContextualError{
		OperationContext: OperationErrorContext{
			OperationType: "DependencyInjection",
		},
	}
	dependencyErr := fmt.Errorf("dependency not found")
	reporter.ReportError(dependencyErr, contextualErr)

	assert.True(t, handlerCalled, "Handler should have been called for high severity error")
}

func TestErrorReporter_ReportError(t *testing.T) {
	reporter := NewErrorReporter()

	err := fmt.Errorf("test error")
	classified := reporter.ReportError(err, nil)

	assert.NotNil(t, classified)
	assert.Equal(t, err, classified.OriginalError)
	assert.NotZero(t, classified.Timestamp)
	assert.NotEmpty(t, classified.ErrorID)
}

func TestGetErrorStatistics(t *testing.T) {
	// Create test errors
	errors := []*ClassifiedError{
		{
			Classification: ErrorClassification{
				Category:       CategoryDependencyInjection,
				Subcategory:    SubcategoryDIResolution,
				Severity:       SeverityCritical,
				Recoverability: RecoverabilityNone,
				UserImpact:     UserImpactBlocking,
				Tags:           []string{"di", "critical"},
			},
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			Classification: ErrorClassification{
				Category:       CategoryHTTP,
				Subcategory:    SubcategoryHTTPClientError,
				Severity:       SeverityMedium,
				Recoverability: RecoverabilityAutomatic,
				UserImpact:     UserImpactDegraded,
				Tags:           []string{"http", "client"},
			},
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
		{
			Classification: ErrorClassification{
				Category:       CategoryDependencyInjection,
				Subcategory:    SubcategoryDIRegistration,
				Severity:       SeverityHigh,
				Recoverability: RecoverabilityManual,
				UserImpact:     UserImpactBlocking,
				Tags:           []string{"di", "registration"},
			},
			Timestamp: time.Now(),
		},
	}

	stats := GetErrorStatistics(errors)

	assert.Equal(t, 3, stats.TotalErrors)
	assert.Equal(t, 1, stats.CriticalErrorCount)
	assert.Equal(t, 1, stats.RecoverableErrorCount)
	assert.Equal(t, 2, stats.UserBlockingErrorCount)

	// Category stats
	assert.Equal(t, 2, stats.ByCategory[CategoryDependencyInjection])
	assert.Equal(t, 1, stats.ByCategory[CategoryHTTP])

	// Severity stats
	assert.Equal(t, 1, stats.BySeverity[SeverityCritical])
	assert.Equal(t, 1, stats.BySeverity[SeverityHigh])
	assert.Equal(t, 1, stats.BySeverity[SeverityMedium])

	// Tag stats
	assert.Equal(t, 2, stats.MostFrequentTags["di"])
	assert.Equal(t, 1, stats.MostFrequentTags["http"])
	assert.Equal(t, 1, stats.MostFrequentTags["client"])
	assert.Equal(t, 1, stats.MostFrequentTags["registration"])

	// Time range
	assert.True(t, stats.TimeRange.Duration() > 0)
	assert.True(t, stats.TimeRange.Start.Before(stats.TimeRange.End))
}

func TestGetErrorStatistics_EmptySlice(t *testing.T) {
	stats := GetErrorStatistics([]*ClassifiedError{})

	assert.Equal(t, 0, stats.TotalErrors)
	assert.Equal(t, 0, stats.CriticalErrorCount)
	assert.Equal(t, 0, stats.RecoverableErrorCount)
	assert.Equal(t, 0, stats.UserBlockingErrorCount)
	assert.Empty(t, stats.ByCategory)
	assert.Empty(t, stats.BySeverity)
}

func TestTimeRange_Duration(t *testing.T) {
	start := time.Now()
	end := start.Add(5 * time.Minute)

	tr := TimeRange{
		Start: start,
		End:   end,
	}

	assert.Equal(t, 5*time.Minute, tr.Duration())
}

func TestGlobalErrorReporter(t *testing.T) {
	reporter1 := GetGlobalErrorReporter()
	reporter2 := GetGlobalErrorReporter()

	assert.NotNil(t, reporter1)
	assert.Same(t, reporter1, reporter2, "Should return the same instance")
}

func TestClassifyAndReport(t *testing.T) {
	err := fmt.Errorf("test error")
	contextualErr := &ContextualError{
		ServiceContext: ServiceErrorContext{
			ServiceName: "TestService",
		},
		OperationContext: OperationErrorContext{
			Operation: "TestOperation",
		},
	}

	classified := ClassifyAndReport(err, contextualErr)

	assert.NotNil(t, classified)
	assert.Equal(t, err, classified.OriginalError)
	assert.Equal(t, "TestService", classified.Context["service_name"])
	assert.Equal(t, "TestOperation", classified.Context["operation"])
}

func TestErrorClassifier_Classify_Performance(t *testing.T) {
	classifier := NewErrorClassifier()

	err := fmt.Errorf("operation timeout after 30 seconds")
	classified := classifier.Classify(err, nil)

	assert.Equal(t, CategoryPerformance, classified.Classification.Category)
	assert.Equal(t, SubcategoryHTTPTimeout, classified.Classification.Subcategory)
	assert.Equal(t, SeverityMedium, classified.Classification.Severity)
	assert.Equal(t, RecoverabilityTransient, classified.Classification.Recoverability)
	assert.Equal(t, UserImpactPerformance, classified.Classification.UserImpact)
}

func TestErrorClassifier_Classify_Validation(t *testing.T) {
	classifier := NewErrorClassifier()

	tests := []struct {
		name          string
		errorMsg      string
		contextualErr *ContextualError
	}{
		{
			name:     "validation error with context",
			errorMsg: "field validation failed",
			contextualErr: &ContextualError{
				OperationContext: OperationErrorContext{
					OperationType: "Validation",
				},
			},
		},
		{
			name:          "validation error without context",
			errorMsg:      "validation constraint violated",
			contextualErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf(tt.errorMsg)
			classified := classifier.Classify(err, tt.contextualErr)

			assert.Equal(t, CategoryValidation, classified.Classification.Category)
			assert.Equal(t, SubcategoryValidationConstraint, classified.Classification.Subcategory)
			assert.Equal(t, SeverityMedium, classified.Classification.Severity)
		})
	}
}
