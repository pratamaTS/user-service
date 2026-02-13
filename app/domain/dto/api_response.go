package dto

type APIResponse[T any] struct {
	Error      bool   `json:"error"`
	Message    string `json:"message"`
	Data       T      `json:"data"`
	Pagination *T     `json:"pagination"`
}
