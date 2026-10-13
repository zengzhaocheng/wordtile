package apperr

import "fmt"

// Error 携带 HTTP 状态码的业务错误，Handler 层统一映射为 JSON 响应。
type Error struct {
	Status int    `json:"-"`
	Msg    string `json:"error"`
}

func (e *Error) Error() string { return e.Msg }

func New(status int, format string, args ...any) *Error {
	return &Error{Status: status, Msg: fmt.Sprintf(format, args...)}
}

var (
	ErrUnauthorized = New(401, "未认证或令牌无效")
	ErrForbidden    = New(403, "无权操作该资源")
	ErrNotFound     = New(404, "资源不存在")
)
