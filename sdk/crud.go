package sdk

type ReadInstanceDTO struct {
	Id string `json:"id,omitempty"`
}

type CreateDTO[T any] struct {
	Data T `json:"data" binding:"required"`
}

type UpdateByIdDTO[T any] struct {
	Id   string `json:"id,omitempty"`
	Data T      `json:"data" binding:"required"`
}

type DeleteByIdDTO struct {
	Id string `json:"id,omitempty"`
}

type ReadDTO struct {
	Filter     map[string]interface{} `json:"filter"`
	Projection map[string]interface{} `json:"projection"`
}
