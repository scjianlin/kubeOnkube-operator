package responseutil

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Gin struct {
	Ctx *gin.Context
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	State   bool        `json:"state"`
	Data    interface{} `json:"data"`
}

type SailorResp struct {
	Code       int         `json:"code"`
	TotalCount int         `json:"total_count"`
	Items      interface{} `json:"items"`
}

// definition Unauthorized response function
func (g *Gin) UnauthorizedResp(code int, err string) {
	g.Ctx.IndentedJSON(http.StatusBadRequest, gin.H{
		"success":   false,
		"message":   err,
		"resultMap": nil,
	})
	return
}

//
func (g *Gin) Response(code int, msg string, state bool, data interface{}) {
	g.Ctx.JSON(200, Response{
		Code:    code,
		Message: msg,
		State:   state,
		Data:    data,
	})
	return
}

func (g *Gin) BindError() {
	g.Ctx.JSON(400, Response{
		Code:    400,
		Message: GetRequestMsg(HTTP_REQUEST_BIND_ERROR),
		Data:    nil,
		State:   false,
	})
	return
}

func (g *Gin) GetArgsError() {
	g.Ctx.JSON(400, Response{
		Code:    HTTP_GET_ARGS_ERRPR,
		Message: GetRequestMsg(HTTP_GET_ARGS_ERRPR),
		Data:    nil,
		State:   false,
	})
	return
}

func (g *Gin) SqlExecError() {
	g.Ctx.JSON(400, Response{
		Code:    SQL_EXEC_ERROR,
		Message: GetRequestMsg(SQL_EXEC_ERROR),
		Data:    nil,
		State:   false,
	})
	return
}
