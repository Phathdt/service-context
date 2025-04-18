package core

import "net/http"

// Response helpers
var (
	SimpleSuccessResponse = func(data any) Response {
		return newResponse(http.StatusOK, data, nil, nil)
	}

	ResponseWithPaging = func(data, param any, other any) Response {
		if v, ok := other.(Paging); ok {
			return newResponse(http.StatusOK, data, param, v)
		}
		return newResponse(http.StatusOK, data, param, other)
	}
)

type Response struct {
	Code   int `json:"code"`
	Data   any `json:"data"`
	Param  any `json:"param,omitempty"`
	Paging any `json:"paging,omitempty"`
}

func newResponse(code int, data, param, other any) Response {
	return Response{
		Code:   code,
		Data:   data,
		Param:  param,
		Paging: other,
	}
}
