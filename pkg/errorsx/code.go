package errorsx

import "net/http"

// errorsx 预定义标准的错误.
var (
	// ErrInternal 表示所有未知的服务器端错误.
	ErrInternal = &ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError", Message: "Internal server error."}
)
