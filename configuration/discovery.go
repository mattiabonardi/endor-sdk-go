package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mattiabonardi/endor-sdk-go/models"
	"gopkg.in/yaml.v3"
)

func InitServiceDiscovery(microServiceId string, microServiceAddress string, endpoints []string) error {
	// create model
	routers := make(map[string]models.Router)
	pattern := regexp.MustCompile(`^(/api/v1/[^/]+/[^/]+)(/.*)?$`)

	for i, endpoint := range endpoints {
		key := fmt.Sprintf("%s-router-%d", microServiceId, i)

		// Extract base path up to the first segment after /:app/
		match := pattern.FindStringSubmatch(endpoint)
		truncatedEndpoint := endpoint
		if len(match) > 1 {
			// keep only first segment "/api/v1/:app/{first-segment}"
			truncatedEndpoint = match[1]
		}

		routers[key] = models.Router{
			Rule:        fmt.Sprintf("PathRegexp(`^%s/.*$`)", truncatedEndpoint),
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

	discoveryConfiguration := models.DiscoveryConfiguration{
		HTTP: models.HTTPConfig{
			Routers:  routers,
			Services: services,
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
