package types

type BaseResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"string"`
	Data    T      `json:"data,omitempty"`
}
