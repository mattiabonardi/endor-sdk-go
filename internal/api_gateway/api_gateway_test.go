package api_gateway_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	test_utils_handlers "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"gopkg.in/yaml.v3"
)

func TestInitializeApiGatewayConfiguration(t *testing.T) {
	// Setup test data
	domain := "test-service"
	version := "v1"
	microHandlerAddress := "http://localhost:8080"

	// Use BaseHandler as test EndorHandler
	baseHandler := test_utils_handlers.NewBaseHandlerHandler()
	services := []sdk.EndorHandler{baseHandler.ToEndorHandler()}

	// Test the function
	err := api_gateway.InitializeApiGatewayConfiguration(domain, version, microHandlerAddress, services)
	if err != nil {
		t.Fatalf("InitializeApiGatewayConfiguration failed: %v", err)
	}

	// Verify the configuration file was created
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", domain))

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
			"test-service-router",
			"test-service-router-base-handler-public-action",
		}

		for _, routerName := range expectedRouters {
			router, exists := config.HTTP.Routers[routerName]
			if !exists {
				t.Errorf("Expected router %s not found", routerName)
				continue
			}

			// Verify router properties
			if router.Service != domain {
				t.Errorf("Router %s has wrong service: got %s, want %s", routerName, router.Service, domain)
			}

			if len(router.EntryPoints) != 1 || router.EntryPoints[0] != "web" {
				t.Errorf("Router %s has wrong entry points: got %v, want [web]", routerName, router.EntryPoints)
			}

			switch routerName {
			case "test-service-router":
				// Wildcard router must have auth middleware
				if router.Middlewares == nil || len(*router.Middlewares) != 1 || (*router.Middlewares)[0] != "authMiddleware" {
					t.Errorf("Wildcard router %s should have authMiddleware, got %v", routerName, router.Middlewares)
				}
			case "test-service-router-base-handler-public-action":
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
		service, exists := config.HTTP.Services[domain]
		if !exists {
			t.Fatalf("Expected service %s not found", domain)
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
			"test-service-router":                            "PathPrefix(`/api/test-service`)",
			"test-service-router-base-handler-public-action": "PathPrefix(`/api/test-service/v1/base-handler/public-action`)",
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

func TestInitializeApiGatewayConfigurationWithVersion(t *testing.T) {
	domain := "test-service-v2"
	version := "v2"
	microHandlerAddress := "http://localhost:8081"

	// Create a service with custom version
	baseHandler := test_utils_handlers.NewBaseHandlerHandler()
	endorHandler := baseHandler.ToEndorHandler()
	services := []sdk.EndorHandler{endorHandler}

	err := api_gateway.InitializeApiGatewayConfiguration(domain, version, microHandlerAddress, services)
	if err != nil {
		t.Fatalf("InitializeApiGatewayConfiguration failed: %v", err)
	}

	// Read the configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", domain))
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
	}

	var config api_gateway.ApiGatewayConfiguration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Failed to parse configuration YAML: %v", err)
	}

	// Verify that the version is correctly included in the public action rule
	publicRouterName := "test-service-v2-router-base-handler-public-action"
	publicRouter, exists := config.HTTP.Routers[publicRouterName]
	if !exists {
		t.Errorf("Expected public action router %s not found", publicRouterName)
	} else if !strings.Contains(publicRouter.Rule, "/api/test-service-v2/v2/base-handler/") {
		t.Errorf("Public router %s should contain microserviceId and v2 version in rule: %s", publicRouterName, publicRouter.Rule)
	}

	// Verify wildcard router uses the microserviceId context
	wildcardRouterName := "test-service-v2-router"
	wildcardRouter, exists := config.HTTP.Routers[wildcardRouterName]
	if !exists {
		t.Errorf("Expected wildcard router %s not found", wildcardRouterName)
	} else if wildcardRouter.Rule != "PathPrefix(`/api/test-service-v2`)" {
		t.Errorf("Wildcard router has wrong rule: %s", wildcardRouter.Rule)
	}

	// Cleanup
	defer func() {
		os.Remove(filePath)
		os.Remove(filepath.Dir(filePath))
	}()
}

func TestInitializeApiGatewayConfigurationWithPriority(t *testing.T) {
	// Test with priority setting
	domain := "test-service-priority"
	version := "v1"
	microHandlerAddress := "http://localhost:8082"

	baseHandler := test_utils_handlers.NewBaseHandlerHandler()
	endorHandler := baseHandler.ToEndorHandler()
	priority := 100
	endorHandler.Priority = &priority
	services := []sdk.EndorHandler{endorHandler}

	err := api_gateway.InitializeApiGatewayConfiguration(domain, version, microHandlerAddress, services)
	if err != nil {
		t.Fatalf("InitializeApiGatewayConfiguration failed: %v", err)
	}

	// Read the configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", domain))
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
	publicRouterName := "test-service-priority-router-base-handler-public-action"
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
	wildcardRouterName := "test-service-priority-router"
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
