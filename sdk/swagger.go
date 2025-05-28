package sdk

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/sdk/dao"
)

//go:embed swagger/*
var swaggerFS embed.FS

type OpenAPIConfiguration struct {
	OpenAPI        string                          `json:"openapi"`
	Info           OpenAPIInfo                     `json:"info"`
	Servers        []OpenAPIServer                 `json:"servers"`
	EndorResources map[string]OpenAPIEndorResource `json:"endorResources"`
	Paths          map[string]OpenAPIPathItem      `json:"paths"`
	Components     OpenApiComponents               `json:"components"`
}

type OpenAPIEndorResource struct {
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

type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string                       `json:"summary,omitempty"`
	Description string                       `json:"description,omitempty"`
	Tags        []string                     `json:"tags,omitempty"`
	OperationID string                       `json:"operationId,omitempty"`
	RequestBody *OpenAPIRequestBody          `json:"requestBody,omitempty"`
	Responses   OpenApiResponses             `json:"responses"`
	Security    []SwaggerSecurityRequirement `json:"security"`
}

type SwaggerSecurityRequirement map[string][]string

type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
	Required    bool                        `json:"required,omitempty"`
}

type OpenAPIMediaType struct {
	Schema Schema `json:"schema"`
}

type OpenApiComponents struct {
	SecuritySchemas map[string]OpenApiAuth `json:"securitySchemas"`
	Schemas         map[string]Schema      `json:"schemas"`
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
		err = json.Unmarshal([]byte(content), &cfg)
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
		EndorResources: map[string]OpenAPIEndorResource{},
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
			if service.Resource == "resource" {
				// special case for resource service
				paths[fmt.Sprintf("%s/%s/%s/%s/%s", baseApiPath, microServiceId, version, service.Resource, methodKey)] = path
			} else {
				paths[fmt.Sprintf("%s/%s/%s/%s", baseApiPath, version, service.Resource, methodKey)] = path
			}

		}
		swaggerConfiguration.EndorResources[service.Resource] = OpenAPIEndorResource{
			Description: service.Description,
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
