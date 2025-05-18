package sdk

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/sdk/dao"
	"gopkg.in/yaml.v3"
)

//go:embed swagger/*
var swaggerFS embed.FS

type OpenAPIConfiguration struct {
	OpenAPI    string                     `json:"openapi" yaml:"openapi"` // should be "3.0.0"
	Info       OpenAPIInfo                `json:"info" yaml:"info"`
	Servers    []OpenAPIServer            `json:"servers" yaml:"servers"`
	Paths      map[string]OpenAPIPathItem `json:"paths" yaml:"paths"`
	Components OpenApiComponents          `json:"components" yaml:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty" yaml:"delete,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string                       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                       `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string                     `json:"tags,omitempty" yaml:"tags,omitempty"`
	OperationID string                       `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []OpenApiParameter           `json:"parameters" yaml:"parameters"`
	RequestBody *OpenAPIRequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   OpenApiResponses             `json:"responses" yaml:"responses"`
	Security    []SwaggerSecurityRequirement `json:"security" yaml:"security"`
}

type SwaggerSecurityRequirement map[string][]string

type OpenApiParameter struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	In          string            `json:"in,omitempty" yaml:"in,omitempty"`
	Required    bool              `json:"required,omitempty" yaml:"required,omitempty"`
	Default     string            `json:"default,omitempty" yaml:"default,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      map[string]string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content" yaml:"content"`
	Required    bool                        `json:"required,omitempty" yaml:"required,omitempty"`
}

type OpenAPIMediaType struct {
	Schema Schema `json:"schema" yaml:"schema"`
}

type OpenApiComponents struct {
	SecuritySchemas map[string]OpenApiAuth `json:"securitySchemas" yaml:"securitySchemas"`
	Schemas         map[string]Schema      `json:"schemas" yaml:"schemas"`
}

type OpenApiAuth struct {
	Type string `json:"type" yaml:"type"`
	In   string `json:"in" yaml:"in"`
	Name string `json:"name" yaml:"name"`
}

type OpenApiResponse struct {
	Description string `yaml:"description" json:"description"`
	Content     map[string]OpenAPIMediaType
}

type OpenApiResponses map[string]OpenApiResponse

var baseSwaggerFolder = "etc/endor/endor-api-gateway/swagger/"
var configurationFileName = "openapi.yaml"

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

	// copy swagger files
	err = copySwagger(swaggerFolder)
	if err != nil {
		return "", err
	}
	// serialize openapi file
	filePath := filepath.Join(swaggerFolder, configurationFileName)

	data, err := yaml.Marshal(definition)
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

func GetSwaggerConfigurations() ([]OpenAPIConfiguration, error) {
	configs := []OpenAPIConfiguration{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return configs, err
	}
	swaggerFolder := filepath.Join(homeDir, baseSwaggerFolder)
	baseFolderDao, err := dao.NewFileSystemDAO(swaggerFolder)
	if err != nil {
		return configs, err
	}
	folders, err := baseFolderDao.ListFolders()
	if err != nil {
		return configs, err
	}
	// cicle folder and unmarshal configuration
	for _, folder := range folders {
		configFilePath := filepath.Join(swaggerFolder, folder)
		configFsDao, err := dao.NewFileSystemDAO(configFilePath)
		if err != nil {
			return configs, err
		}
		content, err := configFsDao.Instace(configurationFileName)
		if err != nil {
			return configs, err
		}
		var cfg OpenAPIConfiguration
		err = yaml.Unmarshal([]byte(content), &cfg)
		if err != nil {
			return configs, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
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
		Components: OpenApiComponents{
			SecuritySchemas: map[string]OpenApiAuth{
				"cookieAuth": {
					Type: "apiKey",
					In:   "cookie",
					Name: "sessionId",
				},
			},
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

	paths := map[string]OpenAPIPathItem{}
	for _, service := range services {
		for methodKey, method := range service.Methods {
			path := OpenAPIPathItem{
				Post: &OpenAPIOperation{
					OperationID: fmt.Sprintf("%s - %s", service.Resource, methodKey),
					Tags:        []string{service.Resource},
					Parameters: []OpenApiParameter{
						{
							Name:        "app",
							In:          "path",
							Required:    true,
							Default:     "",
							Description: "app",
							Schema: map[string]string{
								"type": "string",
							},
						},
					},
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
				},
			}
			// find payload using reflection
			payload, err := resolvePayloadType(method)
			if err != nil {
				return swaggerConfiguration, err
			}
			requestSchema := NewSchemaByType(payload)
			originalRef := requestSchema.Reference
			// put all payload schemas to components
			for schemaName, schema := range requestSchema.Definitions {
				if _, ok := swaggerConfiguration.Components.Schemas[schemaName]; !ok {
					swaggerConfiguration.Components.Schemas[schemaName] = schema
				}
			}
			// add payload
			if originalRef != "" {
				parts := strings.Split(originalRef, "/")
				last := parts[len(parts)-1]
				path.Post.RequestBody = &OpenAPIRequestBody{
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

			//TODO: check authorization handler

			version := service.Version
			if version == "" {
				version = "v1"
			}
			apiPath := strings.ReplaceAll(baseApiPath, ":app", "{app}")
			paths[fmt.Sprintf("%s/%s/%s/%s", apiPath, version, service.Resource, methodKey)] = path
		}
	}

	swaggerConfiguration.Paths = paths

	return swaggerConfiguration, nil
}

func AdaptSwaggerSchemaToSchema(openApiComponents OpenApiComponents, schema *Schema) RootSchema {
	visited := make(map[string]bool)
	defs := make(map[string]Schema)

	// resolve root schema
	root := resolveSchema(schema, openApiComponents, defs, visited)

	return RootSchema{
		Schema:      *root,
		Definitions: defs,
	}
}

func resolveSchema(s *Schema, components OpenApiComponents, defs map[string]Schema, visited map[string]bool) *Schema {
	if s == nil {
		return nil
	}

	if s.Reference != "" {
		refName := extractRefName(s.Reference)
		if visited[refName] {
			// Already visited
			return &Schema{Reference: "#/$defs/" + refName}
		}

		refSchema, ok := components.Schemas[refName]
		if !ok {
			// Reference not found
			return &Schema{}
		}

		// Mark as visited before resolving to prevent circular recursion
		visited[refName] = true

		// Recursively resolve referenced schema
		resolved := resolveSchema(&refSchema, components, defs, visited)

		// Add to $defs
		defs[refName] = *resolved

		// Return a $ref to the definition
		return &Schema{
			Reference: "#/$defs/" + refName,
		}
	}

	// Deep copy of the schema
	result := &Schema{
		Type:  s.Type,
		Enum:  s.Enum,
		Items: resolveSchema(s.Items, components, defs, visited),
	}

	if s.Properties != nil {
		props := make(map[string]Schema)
		for key, prop := range *s.Properties {
			resolvedProp := resolveSchema(&prop, components, defs, visited)
			props[key] = *resolvedProp
		}
		result.Properties = &props
	}

	return result
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

func resolvePayloadType(method EndorServiceMethod) (reflect.Type, error) {
	val := reflect.ValueOf(method)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	handlers := val.FieldByName("handlers")
	if !handlers.IsValid() || handlers.Len() == 0 {
		return nil, fmt.Errorf("handlers not found or empty")
	}

	handlerFunc := handlers.Index(0)
	handlerType := handlerFunc.Type()

	if handlerType.Kind() != reflect.Func || handlerType.NumIn() == 0 {
		return nil, fmt.Errorf("invalid handler function")
	}

	argType := handlerType.In(0) // should be *sdk.EndorContext[T]
	if argType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("handler parameter is not a pointer")
	}

	elemType := argType.Elem() // should be sdk.EndorContext[T]

	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got: %s", elemType.Kind())
	}

	// Just check for presence of Payload field
	payloadField, ok := elemType.FieldByName("Payload")
	if !ok {
		return nil, fmt.Errorf("payload field not found")
	}

	return payloadField.Type, nil
}
