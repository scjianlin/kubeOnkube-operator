package jwt

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	authoptions "github.com/gostship/kunkka/pkg/apimanager/authentication"
	"github.com/gostship/kunkka/pkg/apimanager/authentication/oauth"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"net/http"
	"strings"
)

type OauthhMgr struct {
	Jwt     *JwtTokenIssuer
	Options *authoptions.AuthenticationOptions
}

// GetOauth token returns user token .
func (h *OauthhMgr) AuthorizeHandler(c *gin.Context) {
	clientId := c.Query("client_id")
	responseType := c.Query("response_type")
	redirectURI := c.Query("redirect_uri")

	// get request Authorization header
	auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
	resp := responseutil.Gin{Ctx: c}

	if len(auth) != 2 || auth[0] != "Basic" {
		resp.UnauthorizedResp(http.StatusUnauthorized, "Authorization type don't Baisc")
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		resp.UnauthorizedResp(http.StatusUnauthorized, "Authorization no found user or passwd")
	}

	conf, err := h.Options.AuthOptions.OAuthClient(clientId)

	if responseType != "token" {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: response type %s is not supported", responseType))
		resp.UnauthorizedResp(http.StatusUnauthorized, err.Error())
	}

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.UnauthorizedResp(http.StatusUnauthorized, err.Error())
	}

	redirectURL, err := conf.ResolveRedirectURL(redirectURI)

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.UnauthorizedResp(http.StatusUnauthorized, err.Error())
	}

	// Get user access token
	token, err := h.issueTo(pair[0])
	if err != nil {
		resp.UnauthorizedResp(http.StatusUnauthorized, err.Error())
	}

	redirectURL = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", redirectURL, token.AccessToken)

	if token.ExpiresIn > 0 {
		redirectURL = fmt.Sprintf("%s&expires_in=%v", redirectURL, token.ExpiresIn)
	}

	c.Header("Content-Type", "text/plain")
	c.Redirect(http.StatusFound, redirectURL)
}

func (h *OauthhMgr) issueTo(username string) (*oauth.Token, error) {
	expiresIn := h.Options.AuthOptions.AccessTokenMaxAge

	accessToken, err := h.Jwt.IssueTo(&authuser.DefaultInfo{
		Name: username,
	}, expiresIn)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := &oauth.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(expiresIn.Seconds()),
	}
	return result, nil
}

// return kunkka server config map
func (k *OauthhMgr) GetConfigMap(c *gin.Context) {
	result := make(map[string]bool, 0)
	result["alerting"] = false
	result["auditing"] = false
	result["authentication"] = true
	result["authorization"] = true
	result["devops"] = false
	result["events"] = false
	result["kubernetes"] = true
	result["logging"] = false
	result["monitoring"] = true
	result["multicluster"] = false
	result["network"] = false
	result["notification"] = false
	result["openpitrix"] = false
	result["redis"] = true
	result["s3"] = false
	result["servicemesh"] = false
	result["sonarqube"] = false
	c.IndentedJSON(200, result)
}
