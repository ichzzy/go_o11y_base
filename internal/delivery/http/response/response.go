package response

import "github.com/ichzzy/go_o11y_base/internal/domain"

type Response struct {
	Code    int    `json:"code"`    // 全局狀態碼
	Message string `json:"message"` // 訊息
	Data    any    `json:"data"`
}

func Success(data any) *Response {
	return &Response{
		Code:    domain.CodeSuccess.Int(),
		Message: "success",
		Data:    data,
	}
}

func Error(err domain.AppError) *Response {
	return &Response{
		Code:    err.Code().Int(),
		Message: err.Message(),
		Data:    struct{}{},
	}
}
