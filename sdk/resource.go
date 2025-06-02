package sdk

type Resource struct {
	ID          string `json:"id" bson:"_id"`
	Description string `json:"description"`
	Service     string `json:"service"`
	Schema      string `json:"schema"`
}
