package dto

import "net/http"

// Response is the standard API response envelope.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c interface{}) Response {
	return Response{Code: 0, Message: "success", Data: c}
}

func Fail(code int, msg string) Response {
	return Response{Code: code, Message: msg}
}

// HTTPStatusFromCode maps business error codes to HTTP status codes.
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
