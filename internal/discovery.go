package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type DiscoveryConfiguration struct {
	HTTP HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Routers  map[string]Router  `yaml:"routers"`
	Services map[string]Service `yaml:"services"`
}

type Router struct {
	Rule        string   `yaml:"rule"`
	Service     string   `yaml:"service"`
	EntryPoints []string `yaml:"entryPoints"`
}

type Service struct {
	LoadBalancer LoadBalancer `yaml:"loadBalancer"`
}

type LoadBalancer struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	URL string `yaml:"url"`
}

func InitServiceDiscovery(microServiceId string, microServiceAddress string, services []EndorService) error {
	// Create model
	routers := make(map[string]Router)

	for _, service := range services {
		// app
		path := "^/api/"
		if len(service.Apps) > 0 {
			path += fmt.Sprintf("(%s)/", strings.Join(service.Apps, "|"))
		} else {
			path += "[^/]+/"
		}
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
		routers[key] = Router{
			Rule:        path,
			Service:     microServiceId,
			EntryPoints: []string{"web"},
		}
	}

	discoveryServices := make(map[string]Service)
	discoveryServices[microServiceId] = Service{
		LoadBalancer: LoadBalancer{
			Servers: []Server{
				{URL: microServiceAddress},
			},
		},
	}

	discoveryConfiguration := DiscoveryConfiguration{
		HTTP: HTTPConfig{
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
