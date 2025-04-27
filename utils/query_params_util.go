package utils

import "math"

type ListParams struct {
	Name string `query:"name"`

	SortBy  string `query:"sortBy"`
	OrderBy string `query:"orderBy"`

	Page    int `query:"page"`
	PerPage int `query:"perPage"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

type PaginatedResult struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func (p *ListParams) CalculateOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

func CalculateTotalPages(totalItems int64, perPage int) int {
	if perPage <= 0 {
		return 1
	}
	return int(math.Ceil(float64(totalItems) / float64(perPage)))
}
