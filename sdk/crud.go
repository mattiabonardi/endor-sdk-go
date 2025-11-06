package sdk

type ReadInstanceDTO[ID comparable] struct {
	Id ID `json:"id,omitempty"`
}

type CreateDTO[T any] struct {
	Data T `json:"data" binding:"required"`
}

type UpdateByIdDTO[T any, ID comparable] struct {
	Id   ID `json:"id,omitempty"`
	Data T  `json:"data" binding:"required"`
}

type DeleteByIdDTO[ID comparable] struct {
	Id ID `json:"id,omitempty"`
}

type ReadDTO struct {
	Filter     map[string]interface{} `json:"filter"`
	Projection map[string]interface{} `json:"projection"`
}
