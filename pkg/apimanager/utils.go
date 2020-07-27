package apimanager

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"k8s.io/klog"
)

func Bind(s interface{}, c *gin.Context) (interface{}, error) {
	b := binding.Default(c.Request.Method, c.ContentType())
	if err := c.ShouldBindWith(s, b); err != nil {
		return nil, err
	}
	return s, nil
}

func generateId() (string, error) {
	u1, err := uuid.NewUUID()
	if err != nil {
		klog.Error("generate uuid error ", err)
		return "", err
	}
	return u1.String(), nil
}
