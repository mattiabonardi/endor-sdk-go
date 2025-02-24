package managers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattiabonardi/endor-sdk-go/models"
	"gopkg.in/yaml.v3"
)

func NewDiscoveryConfiguration(microServiceId string, microServiceAddress string, endpoints []string) models.DiscoveryConfiguration {
	routers := make(map[string]models.Router)
	for i, endpoint := range endpoints {
		key := fmt.Sprintf("%s-router-%d", microServiceId, i)
		routers[key] = models.Router{
			Rule:        fmt.Sprintf("PathRegexp(`^%s$`)", endpoint),
			Service:     microServiceId,
			EntryPoints: []string{"web"},
		}
	}

	services := make(map[string]models.Service)
	services[microServiceId] = models.Service{
		LoadBalancer: models.LoadBalancer{
			Servers: []models.Server{
				{URL: microServiceAddress},
			},
		},
	}

	return models.DiscoveryConfiguration{
		HTTP: models.HTTPConfig{
			Routers:  routers,
			Services: services,
		},
	}
}

func RegisterDiscoveryConfiguration(discoveryConfiguration models.DiscoveryConfiguration, microServiceId string) error {
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
