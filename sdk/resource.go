package sdk

type Resource struct {
	ID          string   `json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Apps        []string `json:"apps,omitempty" yaml:"apps,omitempty"`
	Schema      string   `json:"schema,omitempty" yaml:"schema,omitempty"`
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

type ResourceUpdateByIdDTO struct {
	App  string   `json:"app" binding:"required"`
	Id   string   `json:"id" binding:"required"`
	Data Resource `json:"data" binding:"required"`
}
