package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gostship/kunkka/pkg/apimanager/authentication"
	authoptions "github.com/gostship/kunkka/pkg/apimanager/authentication"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"time"
)

const DefaultIssuerName = "gostship"

var (
	errInvalidToken = errors.New("invalid token")
	errTokenExpired = errors.New("expired token")
)

type Claims struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims
}

type JwtTokenIssuer struct {
	name    string
	options *authentication.AuthenticationOptions
	//cache   cache.Interface
	keyFunc jwt.Keyfunc
}

//
//func (s *jwtTokenIssuer) Verify(tokenString string) (User, error) {
//	if len(tokenString) == 0 {
//		return nil, errInvalidToken
//	}
//
//	clm := &Claims{}
//	_, err := jwt.ParseWithClaims(tokenString, clm, s.keyFunc)
//	if err != nil {
//		return nil, err
//	}
//
//	// accessTokenMaxAge = 0 or token without expiration time means that the token will not expire
//	// do not validate token cache
//	if s.options.AuthOptions.AccessTokenMaxAge > 0 && clm.ExpiresAt > 0 {
//		_, err = s.cache.Get(tokenCacheKey(tokenString))
//
//		if err != nil {
//			if err == cache.ErrNoSuchKey {
//				return nil, errTokenExpired
//			}
//			return nil, err
//		}
//	}
//
//	return &user.DefaultInfo{Name: clm.Username, UID: clm.UID}, nil
//}

func (s *JwtTokenIssuer) IssueTo(user user.Info, expiresIn time.Duration) (string, error) {
	clm := &Claims{
		Username: user.GetName(),
		UID:      user.GetUID(),
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			Issuer:    s.name,
			NotBefore: time.Now().Unix(),
		},
	}

	if expiresIn > 0 {
		clm.ExpiresAt = clm.IssuedAt + int64(expiresIn.Seconds())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)

	tokenString, err := token.SignedString([]byte(""))

	if err != nil {
		klog.Error(err)
		return "", err
	}

	// 0 means no expiration.
	// validate token cache
	//if s.options.OAuthOptions.AccessTokenMaxAge > 0 {
	//	err = s.cache.Set(tokenCacheKey(tokenString), tokenString, s.options.OAuthOptions.AccessTokenMaxAge)
	//	if err != nil {
	//		klog.Error(err)
	//		return "", err
	//	}
	//}

	return tokenString, nil
}

func NewJwtTokenIssuer(options *authoptions.AuthenticationOptions) *JwtTokenIssuer {
	return &JwtTokenIssuer{
		name:    DefaultIssuerName,
		options: options,
		//cache:   cache,
		keyFunc: func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(options.JwtSecret), nil
			} else {
				return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
			}
		},
	}
}
