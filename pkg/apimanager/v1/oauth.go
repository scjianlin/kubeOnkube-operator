package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/util/authutil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"net/http"
	"strings"
	"time"
)

func (m *Manager) AuthorizeHandler(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	//clientId := c.Query("client_id")
	responseType := c.Query("response_type")
	//redirectURI := c.Query("redirect_uri")

	if responseType != "token" {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: response type %s is not supported", responseType))
		resp.RespError(err.Error())
		return
	}

	authorization := c.GetHeader("Authorization")
	splitUserInfo := strings.Split(authorization, " ")
	decodeUser, _ := base64.StdEncoding.DecodeString(splitUserInfo[1])
	userInfo := strings.Split(string(decodeUser), ":")

	username := userInfo[0]
	password := userInfo[1]

	userState := authutil.Authenticate(password)
	if !userState {
		klog.Error("user: %s ,password error.", username)
		return
	}

	redirectURL := "*"
	Token, err := authutil.IssueTo(username)
	if err != nil {
		klog.Error("generate access token error.")
		return
	}
	redirectURL = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", redirectURL, Token.AccessToken)

	if Token.ExpiresIn > 0 {
		redirectURL = fmt.Sprintf("%s&expires_in=%v", redirectURL, Token.ExpiresIn)
	}
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Redirect(http.StatusFound, redirectURL)
}

func (m *Manager) getAuthConfig(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	res := map[string]time.Duration{
		"accessTokenMaxAge":            24 * time.Hour,
		"accessTokenInactivityTimeout": 24 * time.Hour,
	}
	resp.RespJson(res)
}

func (m *Manager) getUserDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	//userName := c.Param("username")
	result := map[string]string{
		"email":      "admin@gostship.io",
		"lang":       "zh",
		"username":   "admin",
		"globalrole": "true",
	}
	resp.RespJson(result)
}

func (m *Manager) getWorkSpace(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	obj, err := authutil.BuildWorkspaceTemplate()
	if err != nil {
		klog.Error("get workspace template error")
		resp.RespError("get workspace template error")
		return
	}
	resp.RespSuccess(true, "OK", obj, 1)
}

func (m *Manager) getClusterConfig(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	result := map[string]bool{
		"alerting":       false,
		"auditing":       false,
		"authentication": true,
		"authorization":  true,
		"devops":         false,
		"events":         false,
		"kubernetes":     true,
		"logging":        false,
		"monitoring":     true,
		"multicluster":   true,
		"network":        false,
		"notification":   false,
		"openpitrix":     false,
		"redis":          true,
		"s3":             false,
		"servicemesh":    false,
		"sonarqube":      false,
	}
	resp.RespJson(result)
}

func (m *Manager) getClusterUser(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	res, err := authutil.BuildUserMap()
	if err != nil {
		klog.Error("build user map error.")
		resp.RespError("build user map error.")
		return
	}
	resp.RespJson(res)
}

func (m *Manager) getGlobalRole(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	res, err := authutil.BuildGlobalRole()
	if err != nil {
		klog.Error("build global role error.")
		resp.RespError("build global role error.")
		return
	}
	resp.RespJson(res)
}
