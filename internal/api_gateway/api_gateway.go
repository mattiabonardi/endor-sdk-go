package api_gateway

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"gopkg.in/yaml.v3"
)

type ApiGatewayConfiguration struct {
	HTTP ApiGatewayConfigurationHTTP `yaml:"http"`
}

type ApiGatewayConfigurationHTTP struct {
	Routers  map[string]ApiGatewayConfigurationRouter  `yaml:"routers"`
	Services map[string]ApiGatewayConfigurationService `yaml:"services"`
}

type ApiGatewayConfigurationRouter struct {
	Rule        string    `yaml:"rule"`
	Service     string    `yaml:"service"`
	Priority    *int      `yaml:"priority,omitempty"`
	EntryPoints []string  `yaml:"entryPoints"`
	Middlewares *[]string `yaml:"middlewares,omitempty"`
}

type ApiGatewayConfigurationService struct {
	LoadBalancer ApiGatewayConfigurationLoadBalancer `yaml:"loadBalancer"`
}

type ApiGatewayConfigurationLoadBalancer struct {
	Servers []ApiGatewayConfigurationServer `yaml:"servers"`
}

type ApiGatewayConfigurationServer struct {
	URL string `yaml:"url"`
}

func InitializeApiGatewayConfiguration(microServiceId string, module string, microServiceAddress string, services []sdk.EndorHandler) error {
	// Create model
	routers := make(map[string]ApiGatewayConfigurationRouter)

	basePath := fmt.Sprintf("/api/v1/%s", module)

	// Single wildcard rule for the entire microservice context with forward auth enabled
	routers[fmt.Sprintf("%s-router", microServiceId)] = ApiGatewayConfigurationRouter{
		Rule:        fmt.Sprintf("PathPrefix(`%s`)", basePath),
		Service:     microServiceId,
		EntryPoints: []string{"web"},
		Middlewares: &[]string{"authMiddleware"},
	}

	for _, s := range services {
		// entity path per service (do not mutate basePath)
		entityPath := basePath + "/" + s.Entity

		// Individual routes only for public actions (override the wildcard, no auth)
		for methodKey, method := range s.Actions {
			if method.GetOptions().Public {
				key := fmt.Sprintf("%s-router-%s-%s", microServiceId, s.Entity, methodKey)
				router := ApiGatewayConfigurationRouter{
					Rule:        fmt.Sprintf("PathPrefix(`%s`)", path.Join(entityPath, methodKey)),
					Service:     microServiceId,
					Priority:    s.Priority,
					EntryPoints: []string{"web"},
				}
				routers[key] = router
			}
		}
	}

	config := sdk_configuration.GetConfig()
	if config.Development == true {
		microServiceAddress = fmt.Sprintf("http://172.17.0.1:%s", config.ServerPort)
	}

	discoveryServices := make(map[string]ApiGatewayConfigurationService)
	discoveryServices[microServiceId] = ApiGatewayConfigurationService{
		LoadBalancer: ApiGatewayConfigurationLoadBalancer{
			Servers: []ApiGatewayConfigurationServer{
				{URL: microServiceAddress},
			},
		},
	}

	discoveryConfiguration := ApiGatewayConfiguration{
		HTTP: ApiGatewayConfigurationHTTP{
			Routers:  routers,
			Services: discoveryServices,
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/dynamic/%s.yaml", microServiceId))

	data, err := yaml.Marshal(discoveryConfiguration)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
