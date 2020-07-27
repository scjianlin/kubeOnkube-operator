package jwt

//type User interface {
//	// Name
//	GetName() string
//
//	// UID
//	GetUID() string
//}
//
//// Issuer issues token to user, tokens are required to perform mutating requests to resources
//type Issuer interface {
//	// IssueTo issues a token a User, return error if issuing process failed
//	IssueTo(user User, expiresIn time.Duration) (string, error)
//
//	// Verify verifies a token, and return a User if it's a valid token, otherwise return error
//	Verify(string) (User, error)
//
//	// Revoke a token,
//	Revoke(token string) error
//}
