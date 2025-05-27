package sdk

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
