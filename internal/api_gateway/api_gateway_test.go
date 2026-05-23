package api_gateway_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	examples_handlers "github.com/mattiabonardi/endor-sdk-go/internal/examples/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"gopkg.in/yaml.v3"
)

func TestInitializeApiGatewayConfiguration(t *testing.T) {
	// Setup test data
	module := "test-service"
	microServiceId := fmt.Sprintf("endor-%s-service", module)
	microHandlerAddress := "http://localhost:8080"

	// Use BaseHandler as test EndorHandler
	baseHandler := examples_handlers.NewBaseHandlerHandler()
	services := []sdk.EndorHandler{baseHandler.ToEndorHandler()}

	// Test the function
	err := api_gateway.InitializeApiGatewayConfiguration(microServiceId, module, microHandlerAddress, services)
	if err != nil {
		t.Fatalf("InitializeApiGatewayConfiguration failed: %v", err)
	}

	// Verify the configuration file was created
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", microServiceId))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Configuration file was not created: %s", filePath)
	}

	// Read and parse the generated configuration
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	var config api_gateway.ApiGatewayConfiguration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse configuration YAML: %v", err)
	}

	// Verify the configuration content
	t.Run("VerifyRouters", func(t *testing.T) {
		// Expect: one wildcard router for the microservice + one router per public action
		expectedRouters := []string{
			fmt.Sprintf("%s-router", microServiceId),
			fmt.Sprintf("%s-router-base-handler-public-action", microServiceId),
		}

		for _, routerName := range expectedRouters {
			router, exists := config.HTTP.Routers[routerName]
			if !exists {
				t.Errorf("Expected router %s not found", routerName)
				continue
			}

			// Verify router properties
			if router.Service != microServiceId {
				t.Errorf("Router %s has wrong service: got %s, want %s", routerName, router.Service, microServiceId)
			}

			if len(router.EntryPoints) != 1 || router.EntryPoints[0] != "web" {
				t.Errorf("Router %s has wrong entry points: got %v, want [web]", routerName, router.EntryPoints)
			}

			switch routerName {
			case fmt.Sprintf("%s-router", microServiceId):
				// Wildcard router must have auth middleware
				if router.Middlewares == nil || len(*router.Middlewares) != 1 || (*router.Middlewares)[0] != "authMiddleware" {
					t.Errorf("Wildcard router %s should have authMiddleware, got %v", routerName, router.Middlewares)
				}
			case fmt.Sprintf("%s-router-base-handler-public-action", microServiceId):
				// Public action router must not have auth middleware
				if router.Middlewares != nil {
					t.Errorf("Public router %s should not have middlewares, but has: %v", routerName, *router.Middlewares)
				}
			}
		}

		// Verify we have exactly the expected number of routers
		if len(config.HTTP.Routers) != len(expectedRouters) {
			t.Errorf("Expected %d routers, got %d", len(expectedRouters), len(config.HTTP.Routers))
		}
	})

	t.Run("VerifyHandlers", func(t *testing.T) {
		service, exists := config.HTTP.Services[microServiceId]
		if !exists {
			t.Fatalf("Expected service %s not found", microServiceId)
		}

		if len(service.LoadBalancer.Servers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(service.LoadBalancer.Servers))
		}

		if service.LoadBalancer.Servers[0].URL != microHandlerAddress {
			t.Errorf("Expected server URL %s, got %s", microHandlerAddress, service.LoadBalancer.Servers[0].URL)
		}
	})

	t.Run("VerifyRulePaths", func(t *testing.T) {
		// Verify that rules have correct path patterns
		expectedPaths := map[string]string{
			fmt.Sprintf("%s-router", microServiceId):                            fmt.Sprintf("PathPrefix(`/api/v1/%s`)", module),
			fmt.Sprintf("%s-router-base-handler-public-action", microServiceId): fmt.Sprintf("PathPrefix(`/api/v1/%s/base-handler/public-action`)", module),
		}

		for routerName, expectedRule := range expectedPaths {
			router, exists := config.HTTP.Routers[routerName]
			if !exists {
				t.Errorf("Router %s not found", routerName)
				continue
			}

			if router.Rule != expectedRule {
				t.Errorf("Router %s has wrong rule: got %s, want %s", routerName, router.Rule, expectedRule)
			}
		}
	})

	// Cleanup: remove the test file
	defer func() {
		os.Remove(filePath)
		// Also try to remove the directory if it's empty
		os.Remove(filepath.Dir(filePath))
	}()
}

func TestInitializeApiGatewayConfigurationWithPriority(t *testing.T) {
	// Test with priority setting

	module := "test-service-priority"
	microServiceId := fmt.Sprintf("endor-%s-service", module)
	microHandlerAddress := "http://localhost:8082"

	baseHandler := examples_handlers.NewBaseHandlerHandler()
	endorHandler := baseHandler.ToEndorHandler()
	priority := 100
	endorHandler.Priority = &priority
	services := []sdk.EndorHandler{endorHandler}

	err := api_gateway.InitializeApiGatewayConfiguration(microServiceId, module, microHandlerAddress, services)
	if err != nil {
		t.Fatalf("InitializeApiGatewayConfiguration failed: %v", err)
	}

	// Read the configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", microServiceId))
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	var config api_gateway.ApiGatewayConfiguration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse configuration YAML: %v", err)
	}

	// The wildcard router has no per-service priority; public action routers inherit it
	publicRouterName := fmt.Sprintf("%s-router-base-handler-public-action", microServiceId)
	publicRouter, exists := config.HTTP.Routers[publicRouterName]
	if !exists {
		t.Errorf("Expected public action router %s not found", publicRouterName)
	} else {
		if publicRouter.Priority == nil {
			t.Errorf("Public action router %s should have priority set", publicRouterName)
		} else if *publicRouter.Priority != priority {
			t.Errorf("Public action router %s has wrong priority: got %d, want %d", publicRouterName, *publicRouter.Priority, priority)
		}
	}

	// Wildcard router should not have a priority override
	wildcardRouterName := fmt.Sprintf("%s-router", microServiceId)
	wildcardRouter, exists := config.HTTP.Routers[wildcardRouterName]
	if !exists {
		t.Errorf("Expected wildcard router %s not found", wildcardRouterName)
	} else if wildcardRouter.Priority != nil {
		t.Errorf("Wildcard router %s should not have a priority override", wildcardRouterName)
	}

	// Cleanup
	defer func() {
		os.Remove(filePath)
		os.Remove(filepath.Dir(filePath))
	}()
}
