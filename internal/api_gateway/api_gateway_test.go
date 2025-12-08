package api_gateway_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	test_utils_service "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/service"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"gopkg.in/yaml.v3"
)

func TestInitializeApiGatewayConfiguration(t *testing.T) {
	// Setup test data
	microServiceId := "test-service"
	microServiceAddress := "http://localhost:8080"

	// Use Service1 as test EndorService
	service1 := test_utils_service.NewService1()
	services := []sdk.EndorService{service1}

	// Test the function
	err := api_gateway.InitializeApiGatewayConfiguration(microServiceId, microServiceAddress, services)
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
		expectedRouters := []string{
			"test-service-router-resource-1-action1",
			"test-service-router-resource-1-cat_1/action1",
			"test-service-router-resource-1-public-action",
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

			// Check middleware assignment based on action type
			if routerName == "test-service-router-resource-1-public-action" {
				// Public action should not have auth middleware
				if router.Middlewares != nil {
					t.Errorf("Public router %s should not have middlewares, but has: %v", routerName, *router.Middlewares)
				}
			} else {
				// Private actions should have auth middleware
				if router.Middlewares == nil {
					t.Errorf("Private router %s should have auth middleware", routerName)
				} else if len(*router.Middlewares) != 1 || (*router.Middlewares)[0] != "authMiddleware" {
					t.Errorf("Private router %s has wrong middlewares: got %v, want [authMiddleware]", routerName, *router.Middlewares)
				}
			}
		}

		// Verify we have exactly the expected number of routers
		if len(config.HTTP.Routers) != len(expectedRouters) {
			t.Errorf("Expected %d routers, got %d", len(expectedRouters), len(config.HTTP.Routers))
		}
	})

	t.Run("VerifyServices", func(t *testing.T) {
		service, exists := config.HTTP.Services[microServiceId]
		if !exists {
			t.Fatalf("Expected service %s not found", microServiceId)
		}

		if len(service.LoadBalancer.Servers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(service.LoadBalancer.Servers))
		}

		if service.LoadBalancer.Servers[0].URL != microServiceAddress {
			t.Errorf("Expected server URL %s, got %s", microServiceAddress, service.LoadBalancer.Servers[0].URL)
		}
	})

	t.Run("VerifyRulePaths", func(t *testing.T) {
		// Verify that rules have correct path patterns
		expectedPaths := map[string]string{
			"test-service-router-resource-1-action1":       "PathPrefix(`/api/v1/resource-1/action1`)",
			"test-service-router-resource-1-cat_1/action1": "PathPrefix(`/api/v1/resource-1/cat_1/action1`)",
			"test-service-router-resource-1-public-action": "PathPrefix(`/api/v1/resource-1/public-action`)",
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
	// Test with a custom version
	microServiceId := "test-service-v2"
	microServiceAddress := "http://localhost:8081"

	// Create a service with custom version
	service1 := test_utils_service.NewService1()
	service1.Version = "v2"
	services := []sdk.EndorService{service1}

	err := api_gateway.InitializeApiGatewayConfiguration(microServiceId, microServiceAddress, services)
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

	// Verify that the version is correctly included in the path
	for routerName, router := range config.HTTP.Routers {
		if !strings.Contains(router.Rule, "/api/v2/resource-1/") {
			t.Errorf("Router %s should contain v2 version in rule: %s", routerName, router.Rule)
		}
	}

	// Cleanup
	defer func() {
		os.Remove(filePath)
		os.Remove(filepath.Dir(filePath))
	}()
}

func TestInitializeApiGatewayConfigurationWithPriority(t *testing.T) {
	// Test with priority setting
	microServiceId := "test-service-priority"
	microServiceAddress := "http://localhost:8082"

	service1 := test_utils_service.NewService1()
	priority := 100
	service1.Priority = &priority
	services := []sdk.EndorService{service1}

	err := api_gateway.InitializeApiGatewayConfiguration(microServiceId, microServiceAddress, services)
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

	// Verify that all routers have the correct priority
	for routerName, router := range config.HTTP.Routers {
		if router.Priority == nil {
			t.Errorf("Router %s should have priority set", routerName)
		} else if *router.Priority != priority {
			t.Errorf("Router %s has wrong priority: got %d, want %d", routerName, *router.Priority, priority)
		}
	}

	// Cleanup
	defer func() {
		os.Remove(filePath)
		os.Remove(filepath.Dir(filePath))
	}()
}
