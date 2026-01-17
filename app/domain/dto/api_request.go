package dto

type APIRequest[T any] struct {
	Search     string `json:"search"`
	SortBy     T      `json:"sort_by"`
	FilterBy   T      `json:"filter_by"`
	Pagination T      `json:"pagination"`
}

type RequestParams[T any] struct {
	Params T `json:"params"`
}
