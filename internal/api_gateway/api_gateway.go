package api_gateway

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
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

func InitializeApiGatewayConfiguration(microServiceId string, microServiceAddress string, services []sdk.EndorService) error {
	// Create model
	routers := make(map[string]ApiGatewayConfigurationRouter)

	for _, service := range services {
		basePath := "/api/"
		// version
		if service.Version != "" {
			basePath += service.Version + "/"
		} else {
			basePath += "v1/"
		}
		// entity
		basePath += service.Entity

		// methods
		for methodKey, method := range service.Actions {
			// create router
			key := fmt.Sprintf("%s-router-%s-%s", microServiceId, service.Entity, methodKey)
			router := ApiGatewayConfigurationRouter{
				Rule:        fmt.Sprintf("PathPrefix(`%s`)", path.Join(basePath, methodKey)),
				Service:     microServiceId,
				Priority:    service.Priority,
				EntryPoints: []string{"web"},
			}
			if !method.GetOptions().Public {
				router.Middlewares = &[]string{"authMiddleware"}
			}
			routers[key] = router
		}
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
