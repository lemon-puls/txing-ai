package global

type Code int

const (
	CodeSuccess = iota
	CodeServerInternalError
	CodeInvalidParams
	CodeNotFound
	CodeNotLogin
)

var MsgMap = map[Code]string{
	CodeSuccess:             "success",
	CodeServerInternalError: "server internal error",
	CodeInvalidParams:       "invalid params",
	CodeNotFound:            "not found",
	CodeNotLogin:            "not login",
}

// code to msg
func (code Code) ToMsg() string {
	msg, ok := MsgMap[code]
	if !ok {
		msg = MsgMap[CodeServerInternalError]
	}
	return msg
}
