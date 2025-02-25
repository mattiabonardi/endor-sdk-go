package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/models"
	"gopkg.in/yaml.v3"
)

func InitServiceDiscovery(microServiceId string, microServiceAddress string, endpoints []string) error {
	// Create model
	routers := make(map[string]models.Router)
	seenEndpoints := make(map[string]bool) // Track unique endpoints

	// Updated regex to capture the first segment after /api/v1/:app/
	pattern := regexp.MustCompile(`^/api/v1/[^/]+/([^/]+)`)

	for _, endpoint := range endpoints {
		match := pattern.FindStringSubmatch(endpoint)
		var lastSegment string

		if len(match) > 1 {
			lastSegment = match[1] // Extract the first segment after /api/v1/:app/
		} else {
			lastSegment = endpoint // Fallback to full endpoint if no match
		}

		// Skip if endpoint is already seen
		if seenEndpoints[lastSegment] {
			continue
		}

		// Mark endpoint as seen
		seenEndpoints[lastSegment] = true
	}
	// create the pattern to match all endpoing
	var segments []string
	for key := range seenEndpoints {
		segments = append(segments, key)
	}
	endpointRegex := strings.Join(segments, "|")

	key := fmt.Sprintf("%s-router", microServiceId)
	routers[key] = models.Router{
		Rule:        fmt.Sprintf("PathRegexp(`^/api/v1/[^/]+/(%s)/.*$`)", endpointRegex),
		Service:     microServiceId,
		EntryPoints: []string{"web"},
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
