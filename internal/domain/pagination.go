package domain

type Cursor struct{ Offset, Limit int }

func NewCursor(offset, limit int) Cursor {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	return Cursor{Offset: offset, Limit: limit}
}
func (c Cursor) Next(count int) Cursor {
	if count < c.Limit {
		return Cursor{Offset: c.Offset + c.Limit, Limit: c.Limit}
	}
	return Cursor{Offset: c.Offset + c.Limit, Limit: c.Limit}
}
func (c Cursor) Page(total int) (int, int) {
	if c.Limit <= 0 {
		return 1, 0
	}
	page := c.Offset/c.Limit + 1
	pages := (total + c.Limit - 1) / c.Limit
	return page, pages
}
func ClampPage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}
	if size > 500 {
		size = 500
	}
	return page, size
}
