package internal

type Resource struct {
	ID          string      `json:"id,omitempty" yaml:"id,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      Schema      `json:"schema,omitempty" yaml:"schema,omitempty"`
	Persistence Persistence `json:"persistence,omitempty" yaml:"persistence,omitempty"`
}

type Persistence struct {
	Type    string            `json:"id,omitempty" yaml:"id,omitempty"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}
