package article

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Pagination struct {
	Page     int
	PageSize int
}

func NewPagination(page, pageSize int) Pagination {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return Pagination{Page: page, PageSize: pageSize}
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}
