package authutil

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gostship/kunkka/pkg/apimanager/model/auth"
	"k8s.io/klog"
	"time"
)

const DefaultIssuerName = "kunkka"

var (
	Password = "P@88w0rd"
)

type Claims struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims
}

func IssueTo(username string) (*auth.Token, error) {
	clm := &Claims{
		Username: username,
		UID:      "0",
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			Issuer:    DefaultIssuerName,
			NotBefore: time.Now().Unix(),
		},
	}
	JwtSecret := DefaultIssuerName
	ExpiresAt := clm.IssuedAt + int64(24*time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)

	tokenString, err := token.SignedString([]byte(JwtSecret))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := &auth.Token{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int(ExpiresAt),
	}

	return result, nil
}

func Authenticate(password string) bool {
	if password == Password {
		return true
	}
	return false
}
