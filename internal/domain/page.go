package domain

type PageReq struct {
	PageIndex int `form:"page_index" binding:"required"`
	PageSize  int `form:"page_size" binding:"required"`
}

func (p *PageReq) Normalize() *PageReq {
	if p.PageIndex < 1 {
		p.PageIndex = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 1000 {
		p.PageSize = 1000
	}
	return p
}

func (p *PageReq) Offset() int {
	return (p.PageIndex - 1) * p.PageSize
}

type PageResp struct {
	List       any            `json:"list"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	PageIndex int   `json:"page_index"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
}

func NewPageResp(data any, total int64, pageReq PageReq) PageResp {
	return PageResp{
		List: data,
		Pagination: PaginationMeta{
			PageIndex: pageReq.PageIndex,
			PageSize:  pageReq.PageSize,
			Total:     total,
		},
	}
}

type CursorPageReq struct {
	Cursor string `form:"cursor" binding:"required"`
	Limit  int    `form:"limit" binding:"required"`
}
type CursorPageResp struct {
	List any    `json:"list"`
	Next string `json:"next"`
}
