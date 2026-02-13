package dto

type FilterRequest struct {
	Search     string         `json:"search"`
	SortBy     map[string]any `json:"sort_by"`
	FilterBy   map[string]any `json:"filter_by"`
	Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	} `json:"pagination"`
}
