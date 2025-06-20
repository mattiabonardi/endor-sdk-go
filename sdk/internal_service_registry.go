package sdk

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	internalServiceDefinitions = make(map[string]ServiceDefinition)
	once                       sync.Once
)

type ServiceDefinition struct {
	instance EndorService
	callback func(*gin.Context)
}

// Singleton-style ServiceRegistry struct
type ServiceRegistry struct {
	client         *mongo.Client
	context        context.Context
	microserviceId string
}

var registry *ServiceRegistry

func (r *ServiceRegistry) Init(client *mongo.Client, ctx context.Context, microserviceId string) {
	r.client = client
	r.context = ctx
	r.microserviceId = microserviceId
}

func GetInternalServiceRegistry() *ServiceRegistry {
	once.Do(func() {
		registry = &ServiceRegistry{}
	})
	return registry
}

func (r *ServiceRegistry) GetService(path string) *ServiceDefinition {
	// search in internal services
	if service, ok := registry.GetInternalServices()[path]; ok {
		return &service
	} else {
		// search in dynamic services
		if service, ok := registry.GetDynamicServices()[path]; ok {
			return &service
		}
	}
	return nil
}

func (r *ServiceRegistry) GetInternalServices() map[string]ServiceDefinition {
	return internalServiceDefinitions
}

func (r *ServiceRegistry) GetDynamicServices() map[string]ServiceDefinition {
	config := LoadConfiguration()
	dynamicServices := map[string]ServiceDefinition{}
	if config.EndorResourceServiceEnabled {
		resources, err := NewResourceRepository(r.microserviceId, []EndorService{}, r.client, r.context, r.microserviceId).DynamiResourceList()
		if err == nil {
			for _, resource := range resources {
				defintion, err := resource.UnmarshalDefinition()
				if err == nil {
					service := NewAbstractResourceService(resource.ID, resource.Description, *defintion, r.client, r.microserviceId, r.context)
					r.RegisterService(&dynamicServices, service, r.microserviceId)
				} else {
					// TODO: non blocked log
				}
			}
		} else {
			// TODO: non blocked log
		}
	}
	return dynamicServices
}

func (r *ServiceRegistry) GetServices() map[string]ServiceDefinition {
	services := map[string]ServiceDefinition{}
	internalServiceDefinitions := r.GetInternalServices()
	for k, v := range internalServiceDefinitions {
		services[k] = v
	}
	dynamicDefinitions := r.GetDynamicServices()
	for k, v := range dynamicDefinitions {
		services[k] = v
	}
	return services
}

func (r *ServiceRegistry) RegisterService(registry *map[string]ServiceDefinition, service EndorService, microserviceId string) {
	version := "v1"
	if service.Version != "" {
		version = service.Version
	}
	basePath := filepath.Join("/api", version, service.Resource)
	for methodPath, method := range service.Methods {
		fullPath := filepath.Join(basePath, methodPath)
		(*registry)[fullPath] = ServiceDefinition{
			instance: service,
			callback: method.CreateHTTPCallback(microserviceId),
		}
	}
}
