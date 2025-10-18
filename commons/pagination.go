package commons

import "fmt"

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type Pagination interface {
	GetPage() int
	GetPageSize() int
	GetOrderBy() string
	GetSort() Order
	Offset() int
	Limit() int
	Order() string
	GetQuery() string
	GetTotalPages(totalItems int64) int
}

type pagination struct {
	Page     int
	PageSize int
	OrderBy  string
	Sort     Order
	Query    string
}

func (p *pagination) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

func (p *pagination) GetPageSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	return p.PageSize
}

func (p *pagination) GetOrderBy() string {
	if p.OrderBy == "" {
		return "updated_at"
	}
	return p.OrderBy
}

func (p *pagination) GetSort() Order {
	if p.Sort == "" {
		return OrderDesc
	}
	return p.Sort
}

func (p *pagination) Offset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}

func (p *pagination) Limit() int {
	return p.GetPageSize()
}

func (p *pagination) Order() string {
	return fmt.Sprintf("%s %s", p.GetOrderBy(), p.GetSort())
}

func (p *pagination) GetTotalPages(totalItems int64) int {
	if totalItems == 0 {
		return 1
	}
	return int((totalItems + int64(p.GetPageSize()) - 1) / int64(p.GetPageSize()))
}

func (p *pagination) GetQuery() string {
	return p.Query
}

func NewPagination(page, pageSize int, orderBy string, sort Order, query string) Pagination {
	return &pagination{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  orderBy,
		Sort:     sort,
		Query:    query,
	}
}
