package sdk

type Resource struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema,omitempty"`
}
