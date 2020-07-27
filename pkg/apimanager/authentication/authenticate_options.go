package authentication

import (
	"fmt"
	"github.com/gostship/kunkka/pkg/apimanager/authentication/oauth"
	"github.com/spf13/pflag"
	"time"
)

type AuthenticationOptions struct {
	// authenticate rate limit will
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticationRateLimiterDuration" yaml:"authenticationRateLimiterDuration"`
	// allow multiple users login at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`
	// secret to signed jwt token
	JwtSecret string `json:"-" yaml:"jwtSecret"`
	// jwt options
	AuthOptions *oauth.Options `json:"oauthOptions" yaml:"oauthOptions"`
}

func NewAuthenticateOptions() *AuthenticationOptions {
	return &AuthenticationOptions{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		AuthOptions:                     oauth.NewOptions(),
		MultipleLogin:                   false,
		JwtSecret:                       "",
	}
}

func (options *AuthenticationOptions) Validate() []error {
	var errs []error
	if len(options.JwtSecret) == 0 {
		errs = append(errs, fmt.Errorf("jwt secret is empty"))
	}

	return errs
}

func (options *AuthenticationOptions) AddFlags(fs *pflag.FlagSet, s *AuthenticationOptions) {
	fs.IntVar(&options.AuthenticateRateLimiterMaxTries, "authenticate-rate-limiter-max-retries", s.AuthenticateRateLimiterMaxTries, "")
	fs.DurationVar(&options.AuthenticateRateLimiterDuration, "authenticate-rate-limiter-duration", s.AuthenticateRateLimiterDuration, "")
	fs.BoolVar(&options.MultipleLogin, "multiple-login", s.MultipleLogin, "Allow multiple login with the same account, disable means only one user can login at the same time.")
	fs.StringVar(&options.JwtSecret, "jwt-secret", s.JwtSecret, "Secret to sign jwt token, must not be empty.")
	fs.DurationVar(&options.AuthOptions.AccessTokenMaxAge, "access-token-max-age", s.AuthOptions.AccessTokenMaxAge, "AccessTokenMaxAgeSeconds  control the lifetime of access tokens, 0 means no expiration.")
}
