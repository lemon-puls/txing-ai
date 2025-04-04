package global

type CustomError struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CustomError) Error() string {
	return e.Msg
}

func NewCustomError(code Code) *CustomError {
	return &CustomError{code, code.ToMsg()}
}

func NewCustomErrorWithMsg(msg string) *CustomError {
	return &CustomError{CodeServerInternalError, msg}
}
