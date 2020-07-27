package responseutil

// definition http status
const (
	HTTP_SUCCESS                = 200
	HTTP_ERROR                  = 500
	ERROR_NOT_AUTH_TOKEN        = 401
	ERROR_AUTH_CHECK_TOKEN_FAIL = 20001
	HTTP_REQUEST_BIND_ERROR     = 20002
	HTTP_GET_ARGS_ERRPR         = 20003
	SQL_EXEC_ERROR              = 20004
)

// definition map of custom message
var MsgFlags = map[int]string{
	HTTP_SUCCESS:                "请求成功",
	HTTP_ERROR:                  "请求失败",
	ERROR_NOT_AUTH_TOKEN:        "TOKEN不存在",
	ERROR_AUTH_CHECK_TOKEN_FAIL: "Token鉴权失败",
	HTTP_REQUEST_BIND_ERROR:     "绑定参数失败",
	HTTP_GET_ARGS_ERRPR:         "获取参数失败",
	SQL_EXEC_ERROR:              "SQL执行失败",
}

// GetRequestMsg  return custom message
func GetRequestMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[HTTP_ERROR]
}
