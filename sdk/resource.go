package sdk

type Resource struct {
	ID          string `json:"id,omitempty" yaml:"id,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type ReadInstanceDTO struct {
	Id string `json:"id,omitempty"`
}

type ResourceListDTO struct {
	App string `json:"app" binding:"required"`
}

type ResourceInstanceDTO struct {
	App string `json:"app" binding:"required"`
	Id  string `json:"id" binding:"required"`
}
