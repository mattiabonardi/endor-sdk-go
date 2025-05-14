package sdk

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type OpenAPIConfiguration struct {
	OpenAPI    string                     `json:"openapi" yaml:"openapi"` // should be "3.0.0"
	Info       OpenAPIInfo                `json:"info" yaml:"info"`
	Servers    []OpenAPIServer            `json:"servers" yaml:"servers"`
	Paths      map[string]OpenAPIPathItem `json:"paths" yaml:"paths"`
	Components OpenApiSecuritySchemes     `json:"components" yaml:"components"`
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

type OpenApiSecuritySchemes struct {
	SecuritySchemes map[string]OpenApiAuth `json:"securitySchemes" yaml:"secutorySchemes"`
}

type OpenApiAuth struct {
	Type string `json:"type" yaml:"type"`
	In   string `json:"in" yaml:"in"`
	Name string `json:"name" yaml:"name"`
}

type OpenApiResponse struct {
	Description string `yaml:"description" json:"description"`
}

type OpenApiResponses map[string]OpenApiResponse

func InitializeSwaggerConfiguration(microServiceId string, microServiceAddress string, services []EndorService, baseApiPath string) (string, error) {
	swaggerConfiguration := OpenAPIConfiguration{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       microServiceId,
			Description: fmt.Sprintf("%s docs", microServiceId),
		},
		Servers: []OpenAPIServer{
			{
				URL: "/",
			},
		},
		Components: OpenApiSecuritySchemes{
			SecuritySchemes: map[string]OpenApiAuth{
				"cookieAuth": {
					Type: "apiKey",
					In:   "cookie",
					Name: "sessionId",
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
							Description: "no description",
						},
					},
				},
			}
			// find payload using reflection
			payload, err := resolvePayloadType(method)
			if err != nil {
				return "", err
			}
			path.Post.RequestBody = &OpenAPIRequestBody{
				Content: map[string]OpenAPIMediaType{
					"application/json": {
						Schema: newFieldSchema(payload),
					},
				},
			}
			//TODO: check authorization handler

			version := service.Version
			if version == "" {
				version = "v1"
			}
			paths[fmt.Sprintf("%s/%s/%s/%s", baseApiPath, version, service.Resource, methodKey)] = path
		}
	}

	swaggerConfiguration.Paths = paths

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	swaggerFolder := filepath.Join(homeDir, fmt.Sprintf("etc/endor/endor-api-gateway/swagger/%s", microServiceId))

	// copy swagger files
	copyDir("./swagger", swaggerFolder)
	// serialize openapi file
	filePath := filepath.Join(swaggerFolder, "openapi.yaml")

	data, err := yaml.Marshal(swaggerConfiguration)
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

// CopyDir copies the src directory to dst, overwriting dst if it already exists.
func copyDir(src string, dst string) error {
	// Remove destination if it exists
	if _, err := os.Stat(dst); err == nil {
		err = os.RemoveAll(dst)
		if err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}

	// Create the destination root directory
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Walk and copy files
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string, perm fs.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
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
