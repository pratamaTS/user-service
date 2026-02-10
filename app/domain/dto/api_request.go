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

type APIQuery struct {
	Search   string            `form:"search"`
	SortBy   map[string]string `form:"sort_by"`
	FilterBy map[string]string `form:"filter_by"`
	Page     int               `form:"page"`
	PageSize int               `form:"page_size"`
}
