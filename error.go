package gozilla

// Error 用来指示错误code
type Error struct {
	Code int
	Msg  string
}

func NewError(code int, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

func (e *Error) Error() string {
	return e.Msg
}

// ErrorWithData 用来在指示错误的同时携带data数据
type ErrorWithData struct {
	Code int
	Msg  string
	Data interface{}
}

func NewErrorWithData(code int, msg string, data interface{}) *ErrorWithData {
	return &ErrorWithData{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func (e *ErrorWithData) Error() string {
	return e.Msg
}
