package responseutil

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Gin struct {
	Ctx *gin.Context
}

func (g *Gin) Bind(s interface{}) (interface{}, error) {
	b := binding.Default(g.Ctx.Request.Method, g.Ctx.ContentType())
	if err := g.Ctx.ShouldBindWith(s, b); err != nil {
		return nil, err
	}
	return s, nil
}

// http bind Params error response
func (g *Gin) RespError(str string) {
	g.Ctx.AbortWithStatusJSON(400, gin.H{
		"success": false,
		"message": str,
		"data":    nil,
	})
	return
}

// http success response
func (g *Gin) RespSuccess(state bool, msg interface{}, data interface{}, total int) {
	g.Ctx.IndentedJSON(200, gin.H{
		"success":     state,
		"message":     msg,
		"items":       data,
		"total_count": total,
	})
	return
}
