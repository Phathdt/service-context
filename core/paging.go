package core

type Paging struct {
	Page       int    `json:"page" form:"page"`
	Limit      int    `json:"limit" form:"limit"`
	Total      int64  `json:"total" form:"-"`
	Cursor     string `json:"cursor" form:"cursor"`
	NextCursor string `json:"next_cursor"`
}

func (p *Paging) Process() {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.Page >= 100 {
		p.Page = 100
	}

	if p.Limit <= 0 {
		p.Limit = 100
	}

	if p.Limit >= 200 {
		p.Limit = 200
	}
}
