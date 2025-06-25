package sdk

type ResourceMethod struct {
	ID          string      `json:"id"`
	Resource    string      `json:"resource"`
	Description string      `json:"description"`
	Schema      *RootSchema `json:"schema"`
}
