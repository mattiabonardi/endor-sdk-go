// Package lifecycle provides service lifecycle management for composed services
// with dependency-aware ordering, health aggregation, and graceful degradation.
package lifecycle

import (
	"context"
	"time"
)

// ServiceState represents the current state of a service in its lifecycle
type ServiceState int

const (
	// Created indicates the service has been instantiated but not started
	Created ServiceState = iota
	// Starting indicates the service is currently in the startup process
	Starting
	// Running indicates the service is fully operational
	Running
	// Stopping indicates the service is currently shutting down
	Stopping
	// Stopped indicates the service has been shut down cleanly
	Stopped
	// Failed indicates the service has failed and cannot operate
	Failed
)

// String returns the string representation of the service state
func (s ServiceState) String() string {
	switch s {
	case Created:
		return "Created"
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus int

const (
	// Healthy indicates the service is operating normally
	Healthy ServiceHealthStatus = iota
	// Degraded indicates the service is operational but with reduced functionality
	Degraded
	// Unhealthy indicates the service is not functioning correctly
	Unhealthy
	// Unknown indicates the health status cannot be determined
	Unknown
)

// String returns the string representation of the health status
func (h ServiceHealthStatus) String() string {
	switch h {
	case Healthy:
		return "Healthy"
	case Degraded:
		return "Degraded"
	case Unhealthy:
		return "Unhealthy"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// DependencyHealth represents the health status of a dependency
type DependencyHealth struct {
	Name   string              `json:"name"`
	Status ServiceHealthStatus `json:"status"`
	Error  string              `json:"error,omitempty"`
}

// HealthStatus represents the comprehensive health status of a service
type HealthStatus struct {
	// Status is the overall health status of the service
	Status ServiceHealthStatus `json:"status"`
	// Details contains additional health information specific to the service
	Details map[string]interface{} `json:"details,omitempty"`
	// LastCheck is the timestamp of the last health check
	LastCheck time.Time `json:"lastCheck"`
	// Dependencies contains the health status of service dependencies
	Dependencies []DependencyHealth `json:"dependencies,omitempty"`
}

// ServiceLifecycleInterface defines the core lifecycle operations for services
type ServiceLifecycleInterface interface {
	// Start initiates the service startup process
	// Returns an error if startup fails
	Start(ctx context.Context) error

	// Stop initiates the service shutdown process
	// Returns an error if shutdown fails
	Stop(ctx context.Context) error

	// HealthCheck performs a health check and returns the current health status
	// Returns the health status with dependency information
	HealthCheck(ctx context.Context) HealthStatus

	// GetState returns the current lifecycle state of the service
	GetState() ServiceState

	// AddHook registers a lifecycle hook for this service
	// Returns an error if the hook cannot be added
	AddHook(hook LifecycleHook) error
}
