package sdk

import (
	"fmt"
	"os"
	"path/filepath"

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
	Rule        string   `yaml:"rule"`
	Service     string   `yaml:"service"`
	Priority    *int     `yaml:"priority,omitempty"`
	EntryPoints []string `yaml:"entryPoints"`
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

func InitializeApiGatewayConfiguration(microServiceId string, microServiceAddress string, services []EndorService) error {
	// Create model
	routers := make(map[string]ApiGatewayConfigurationRouter)

	for _, service := range services {
		path := "^/api/"
		// version
		if service.Version != "" {
			path += service.Version + "/"
		} else {
			path += "v1/"
		}
		// resource
		path += service.Resource + "/.*$"

		// create router
		key := fmt.Sprintf("%s-router-%s", microServiceId, service.Resource)
		routers[key] = ApiGatewayConfigurationRouter{
			Rule:        fmt.Sprintf("PathRegexp(`%s`)", path),
			Service:     microServiceId,
			Priority:    service.Priority,
			EntryPoints: []string{"web"},
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
