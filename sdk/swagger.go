package sdk

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed swagger/*
var swaggerFS embed.FS

type OpenAPIConfiguration struct {
	OpenAPI    string                                 `json:"openapi"`
	Info       OpenAPIInfo                            `json:"info"`
	Servers    []OpenAPIServer                        `json:"servers"`
	Tags       []OpenAPITag                           `json:"tags"`
	Paths      map[string]map[string]OpenAPIOperation `json:"paths"`
	Components OpenApiComponents                      `json:"components"`
}

type OpenAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter  `json:"parameters"`
	RequestBody *OpenAPIRequestBody `json:"requestBody,omitempty"`
	Responses   OpenApiResponses    `json:"responses"`
}

type OpenAPIParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Schema      Schema `json:"schema"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
	Required    bool                        `json:"required,omitempty"`
}

type OpenAPIMediaType struct {
	Schema Schema `json:"schema"`
}

type OpenApiComponents struct {
	Schemas map[string]Schema `json:"schemas"`
}

type OpenApiAuth struct {
	Type string `json:"type"`
	In   string `json:"in"`
	Name string `json:"name"`
}

type OpenApiResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}

type OpenApiResponses map[string]OpenApiResponse

var baseSwaggerFolder = "etc/endor/endor-api-gateway/swagger/"
var configurationFileName = "openapi.json"

func CreateSwaggerConfiguration(microServiceId string, microServiceAddress string, services []EndorService, baseApiPath string) (string, error) {
	definition, err := CreateSwaggerDefinition(microServiceId, microServiceAddress, services, baseApiPath)
	if err != nil {
		return "", err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	swaggerFolder := filepath.Join(homeDir, baseSwaggerFolder, microServiceId)

	err = copySwagger(swaggerFolder)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(swaggerFolder, configurationFileName)

	data, err := json.MarshalIndent(definition, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", err
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(data)
	return swaggerFolder, err
}

func CreateSwaggerDefinition(microServiceId string, microServiceAddress string, services []EndorService, baseApiPath string) (OpenAPIConfiguration, error) {
	swaggerConfiguration := OpenAPIConfiguration{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:       microServiceId,
			Description: fmt.Sprintf("%s docs", microServiceId),
		},
		Servers: []OpenAPIServer{
			{
				URL: "/",
			},
		},
		Tags: []OpenAPITag{},
		Components: OpenApiComponents{
			Schemas: map[string]Schema{
				"DefaultEndorResponse": {
					Type: ObjectType,
					Properties: &map[string]Schema{
						"messages": {
							Items: &Schema{
								Type: ObjectType,
								Properties: &map[string]Schema{
									"gravity": {
										Type: StringType,
										Enum: &[]string{string(Info), string(Warning), string(Error), string(Fatal)},
									},
									"value": {
										Type: StringType,
									},
								},
							},
						},
						"data": {
							Type: ObjectType,
						},
						"schema": {
							Type: ObjectType,
						},
					},
				},
			},
		},
	}

	paths := make(map[string]map[string]OpenAPIOperation)
	for _, service := range services {
		for methodKey, method := range service.Methods {
			parameters := []OpenAPIParameter{}
			if !method.GetOptions().Public {
				parameters = append(parameters, []OpenAPIParameter{
					{
						Name: "X-User-ID",
						In:   "header",
						Schema: Schema{
							Type: StringType,
						},
						Description: "User id",
						Required:    false,
					},
					{
						Name: "X-User-Session",
						In:   "header",
						Schema: Schema{
							Type: StringType,
						},
						Description: "User session",
						Required:    false,
					},
					{
						Name: "X-Development",
						In:   "header",
						Schema: Schema{
							Type: BooleanType,
						},
						Description: "Development flag",
						Required:    false,
					},
				}...)
			}
			operation := OpenAPIOperation{
				OperationID: fmt.Sprintf("%s - %s", service.Resource, methodKey),
				Tags:        []string{service.Resource},
				Parameters:  parameters,
				Summary:     method.GetOptions().Description,
				Responses: OpenApiResponses{
					"default": OpenApiResponse{
						Description: "Default response",
						Content: map[string]OpenAPIMediaType{
							"application/json": {
								Schema: Schema{
									Reference: "#/components/schemas/DefaultEndorResponse",
								},
							},
						},
					},
				},
			}
			if method.GetOptions().InputSchema != nil {
				requestSchema := method.GetOptions().InputSchema
				originalRef := requestSchema.Reference
				// put all payload schemas to components
				for schemaName, schema := range requestSchema.Definitions {
					if _, ok := swaggerConfiguration.Components.Schemas[schemaName]; !ok {
						swaggerConfiguration.Components.Schemas[schemaName] = schema
						// normalize attributes references
						for propertyname, property := range *swaggerConfiguration.Components.Schemas[schemaName].Properties {
							if property.Reference != "" {
								propertyName := extractRefName(property.Reference)
								// set reference to the schema
								prop := (*swaggerConfiguration.Components.Schemas[schemaName].Properties)[propertyname]
								prop.Reference = fmt.Sprintf("#/components/schemas/%s", propertyName)
								(*swaggerConfiguration.Components.Schemas[schemaName].Properties)[propertyname] = prop
							}
							if property.Items != nil && property.Items.Reference != "" {
								propertyName := extractRefName(property.Items.Reference)
								// set reference to the schema
								prop := (*swaggerConfiguration.Components.Schemas[schemaName].Properties)[propertyname]
								prop.Items.Reference = fmt.Sprintf("#/components/schemas/%s", propertyName)
								(*swaggerConfiguration.Components.Schemas[schemaName].Properties)[propertyname] = prop
							}
						}
					}
				}
				// add payload
				if originalRef != "" {
					last := extractFinalSegment(originalRef)
					operation.RequestBody = &OpenAPIRequestBody{
						Content: map[string]OpenAPIMediaType{
							"application/json": {
								Schema: Schema{
									// calculate reference
									Reference: fmt.Sprintf("#/components/schemas/%s", last),
								},
							},
						},
					}
				}
			}

			version := service.Version
			if version == "" {
				version = "v1"
			}

			path := map[string]OpenAPIOperation{}
			path["post"] = operation
			paths[fmt.Sprintf("%s/%s/%s/%s", baseApiPath, version, service.Resource, methodKey)] = path
		}
		tag := OpenAPITag{
			Name:        service.Resource,
			Description: service.Description,
		}
		swaggerConfiguration.Tags = append(swaggerConfiguration.Tags, tag)
	}

	swaggerConfiguration.Paths = paths

	return swaggerConfiguration, nil
}

func extractRefName(ref string) string {
	// assuming refs look like "#/components/schemas/ModelName"
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func copySwagger(toDir string) error {
	// Remove the target directory if it exists
	if err := os.RemoveAll(toDir); err != nil {
		return err
	}

	// Recreate the target directory
	if err := os.MkdirAll(toDir, 0755); err != nil {
		return err
	}

	// Walk and copy embedded Swagger files
	return fs.WalkDir(swaggerFS, "swagger", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, "swagger")
		relPath = strings.TrimPrefix(relPath, string(os.PathSeparator)) // Remove leading slash if any
		dstPath := filepath.Join(toDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		data, err := swaggerFS.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, 0644)
	})
}

// extractFinalSegment extracts the final segment of a string after the last slash
func extractFinalSegment(input string) string {
	var lastSlashOutsideBrackets = -1
	bracketDepth := 0

	for i, r := range input {
		switch r {
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '/':
			if bracketDepth == 0 {
				lastSlashOutsideBrackets = i
			}
		}
	}

	if lastSlashOutsideBrackets != -1 && lastSlashOutsideBrackets+1 < len(input) {
		return input[lastSlashOutsideBrackets+1:]
	}
	return input
}
