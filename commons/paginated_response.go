package commons

type PaginatedResult[Entity any] struct {
	Data       []Entity `json:"data"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

func NewPaginatedResult[Entity any](data []Entity, total int64, pagination Pagination) *PaginatedResult[Entity] {
	return &PaginatedResult[Entity]{
		Data:       data,
		Total:      total,
		Page:       pagination.GetPage(),
		PageSize:   pagination.GetPageSize(),
		TotalPages: pagination.GetTotalPages(total),
	}
}
