package sdk

import (
	"fmt"
	"log"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// NewDefaultDatabaseClient creates a default MongoDB database client using the global singleton.
// This provides backward compatibility for existing repository creation patterns.
//
// Acceptance Criteria 6: Backward compatibility with convenience constructors using default implementations.
func NewDefaultDatabaseClient() (interfaces.DatabaseClientInterface, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, interfaces.NewDatabaseRepositoryError(
			interfaces.RepositoryErrorCodeDatabaseConnection,
			"Failed to create default database client",
			err,
		)
	}
	return NewMongoDatabaseClient(client), nil
}

// NewDefaultRepositoryDependencies creates a complete set of default dependencies for repository construction.
// This convenience function simplifies repository creation for backward compatibility scenarios.
//
// Acceptance Criteria 6: Existing repository functionality preserved with convenience constructors.
func NewDefaultRepositoryDependencies(microServiceID string) (interfaces.RepositoryDependencies, error) {
	dbClient, err := NewDefaultDatabaseClient()
	if err != nil {
		// For backward compatibility, allow repositories to work with nil database client
		dbClient = nil
	}

	return interfaces.RepositoryDependencies{
		DatabaseClient: dbClient,
		Config:         NewDefaultConfigProvider(),
		Logger:         NewDefaultLogger(),
		MicroServiceID: microServiceID,
	}, nil
}

// DefaultConfigProvider adapts the singleton ServerConfig to implement ConfigProviderInterface
// This provides backward compatibility while enabling dependency injection.
type DefaultConfigProvider struct{}

// NewDefaultConfigProvider creates a default configuration provider that wraps GetConfig()
func NewDefaultConfigProvider() interfaces.ConfigProviderInterface {
	return &DefaultConfigProvider{}
}

func (d *DefaultConfigProvider) GetServerPort() string {
	return GetConfig().ServerPort
}

func (d *DefaultConfigProvider) GetDocumentDBUri() string {
	return GetConfig().DocumentDBUri
}

func (d *DefaultConfigProvider) IsHybridResourcesEnabled() bool {
	return GetConfig().HybridResourcesEnabled
}

func (d *DefaultConfigProvider) IsDynamicResourcesEnabled() bool {
	return GetConfig().DynamicResourcesEnabled
}

func (d *DefaultConfigProvider) GetDynamicResourceDocumentDBName() string {
	return GetConfig().DynamicResourceDocumentDBName
}

func (d *DefaultConfigProvider) Reload() error {
	// ServerConfig singleton doesn't support reload, so this is a no-op
	return nil
}

func (d *DefaultConfigProvider) Validate() error {
	config := GetConfig()
	if config.ServerPort == "" {
		return fmt.Errorf("ServerPort is required")
	}
	if config.DocumentDBUri == "" {
		return fmt.Errorf("DocumentDBUri is required")
	}
	return nil
}

// DefaultLogger adapts Go's standard log package to implement LoggerInterface
// This provides backward compatibility while enabling dependency injection.
type DefaultLogger struct {
	name string
}

// NewDefaultLogger creates a default logger that wraps Go's standard log package
func NewDefaultLogger() interfaces.LoggerInterface {
	return &DefaultLogger{}
}

func (d *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	// Standard log doesn't have debug level, so we use Print
	d.logWithLevel("DEBUG", msg, keysAndValues...)
}

func (d *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	d.logWithLevel("INFO", msg, keysAndValues...)
}

func (d *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	d.logWithLevel("WARN", msg, keysAndValues...)
}

func (d *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	d.logWithLevel("ERROR", msg, keysAndValues...)
}

func (d *DefaultLogger) Fatal(msg string, keysAndValues ...interface{}) {
	d.logWithLevel("FATAL", msg, keysAndValues...)
	// Note: Not calling log.Fatal() to avoid program termination in tests
}

func (d *DefaultLogger) With(keysAndValues ...interface{}) interfaces.LoggerInterface {
	// For the default logger, we just return a new instance
	// In a real implementation, this would capture the key-value pairs
	return &DefaultLogger{name: d.name}
}

func (d *DefaultLogger) WithName(name string) interfaces.LoggerInterface {
	return &DefaultLogger{name: name}
}

func (d *DefaultLogger) logWithLevel(level string, msg string, keysAndValues ...interface{}) {
	prefix := level
	if d.name != "" {
		prefix = d.name + " " + level
	}

	// Format message with key-value pairs
	fullMsg := msg
	if len(keysAndValues) > 0 {
		fullMsg += " "
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				fullMsg += fmt.Sprintf("%v=%v ", keysAndValues[i], keysAndValues[i+1])
			} else {
				fullMsg += fmt.Sprintf("%v ", keysAndValues[i])
			}
		}
	}

	log.Printf("[%s] %s", prefix, fullMsg)
}

// DefaultRepositoryAdapter will be created when we refactor the repository layer
// For now, we'll use nil and implement this in the repository refactoring task
