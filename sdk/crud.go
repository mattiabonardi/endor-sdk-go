package sdk

type CreateDTO[T any] struct {
	Data T `json:"data" binding:"required"`
}
