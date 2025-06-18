package sdk

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	ID          string `json:"id" bson:"_id"`
	Description string `json:"description"`
	Service     string `json:"service"`
	Definition  string `json:"definition"` // YAML string, raw
}

type ResourceDefinition struct {
	Schema      RootSchema     `yaml:"schema"`
	DataSources DataSourceList `yaml:"dataSources"`
	Id          string         `yaml:"id"`
}

type DataSourceList []DataSource

type DataSource interface {
	GetName() string
	GetType() string
}

type BaseDataSource struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`

	// define the dependency of the data source. For example in the multi datasource case
	Dependencies []string `yaml:"dependencies"`
}

func (b *BaseDataSource) GetName() string { return b.Name }
func (b *BaseDataSource) GetType() string { return b.Type }

type MongoDataSource struct {
	BaseDataSource `yaml:",inline"`
	Collection     string                       `yaml:"collection"`
	Mappings       map[string]MongoFieldMapping `yaml:"mappings"`
}

type MongoFieldMapping struct {
	Path string `yaml:"path"`
}

func (dsl *DataSourceList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected a sequence for dataSources")
	}

	var result []DataSource

	for _, item := range value.Content {
		if item.Kind != yaml.MappingNode {
			return fmt.Errorf("expected a mapping node in dataSources")
		}

		var typeName string
		for i := 0; i < len(item.Content); i += 2 {
			keyNode := item.Content[i]
			valNode := item.Content[i+1]
			if keyNode.Value == "type" {
				typeName = valNode.Value
				break
			}
		}

		if typeName == "" {
			return fmt.Errorf("missing or invalid 'type' field in data source")
		}

		var ds DataSource
		switch typeName {
		case "mongodb":
			ds = &MongoDataSource{}
		default:
			return fmt.Errorf("unsupported data source type: %s", typeName)
		}

		if err := item.Decode(ds); err != nil {
			return fmt.Errorf("error decoding data source of type %s: %w", typeName, err)
		}

		result = append(result, ds)
	}

	*dsl = result
	return nil
}

func (h *ResourceDefinition) ToYAML() (string, error) {
	yamlData, err := yaml.Marshal(&h)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

func (h *Resource) UnmarshalDefinition() (*ResourceDefinition, error) {
	var def ResourceDefinition
	err := yaml.Unmarshal([]byte(h.Definition), &def)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &def, nil
}

type ResurceRepositoryInterface interface {
	Instance(dto ReadInstanceDTO) (any, error)
	List() ([]any, error)
	Create(dto CreateDTO[any]) error
	Delete(dto DeleteByIdDTO) error
	Update(dto UpdateByIdDTO[any]) (any, error)
}

type ResourceSliceContext struct {
	dataSource DataSource
	repository ResurceRepositoryInterface
}
