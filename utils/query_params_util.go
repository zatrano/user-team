package utils

import "math"

// ListParams genel filtreleme, sıralama ve sayfalama parametrelerini tutar
type ListParams struct {
	// Filtreler (Projenize göre genişletin)
	Name string `query:"name"` // Fiber'ın query parser'ı için etiket

	// Sıralama
	SortBy  string `query:"sortBy"`
	OrderBy string `query:"orderBy"` // asc veya desc

	// Sayfalama
	Page    int `query:"page"`
	PerPage int `query:"perPage"`
}

// PaginationMeta sayfalama meta verilerini tutar
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

// PaginatedResult genel sayfalama sonucunu tutar
type PaginatedResult struct {
	Data interface{}    `json:"data"` // Herhangi bir model listesi olabilir (örn: []User)
	Meta PaginationMeta `json:"meta"`
}

// CalculateOffset sayfalama için offset değerini hesaplar
func (p *ListParams) CalculateOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

// CalculateTotalPages toplam sayfa sayısını hesaplar
func CalculateTotalPages(totalItems int64, perPage int) int {
	if perPage <= 0 {
		return 1 // Veya 0, duruma göre
	}
	return int(math.Ceil(float64(totalItems) / float64(perPage)))
}
