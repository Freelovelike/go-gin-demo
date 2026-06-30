package dto

import "net/http"

// Response 是标准的 API 响应封包。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 生成表示成功（状态码 0）的 API 响应。
func OK(c interface{}) Response {
	return Response{Code: 0, Message: "success", Data: c}
}

// Fail 生成表示失败的 API 响应，并附带指定的错误代码和消息。
func Fail(code int, msg string) Response {
	return Response{Code: code, Message: msg}
}

// HTTPStatusFromCode 将业务错误代码映射到 HTTP 状态代码。
func HTTPStatusFromCode(code int) int {
	switch {
	case code == 0:
		return http.StatusOK
	case code >= 1000 && code < 2000:
		return http.StatusBadRequest
	case code >= 2000 && code < 3000:
		return http.StatusUnauthorized
	case code >= 3000 && code < 4000:
		return http.StatusForbidden
	case code >= 4000 && code < 5000:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
