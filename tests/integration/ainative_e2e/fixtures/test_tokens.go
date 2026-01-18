package fixtures

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// testSecret is the secret key used for signing test JWT tokens
var testSecret = []byte("test-secret-key-for-jwt-signing-in-e2e-tests")

// GenerateTestToken generates a test JWT token with the given email and duration
// Each call generates a unique token by including nanosecond timestamp
func GenerateTestToken(email string, duration time.Duration) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   "user-" + email,
		"email": email,
		"exp":   now.Add(duration).Unix(),
		"iat":   now.Unix(),
		"jti":   fmt.Sprintf("%d", now.UnixNano()), // Unique token ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(testSecret)
	if err != nil {
		panic(err) // Should never happen in tests
	}

	return tokenString
}

// CreateExpiredToken creates a token that expired 1 hour ago
func CreateExpiredToken() string {
	return GenerateTestToken("test@example.com", -1*time.Hour)
}

// CreateValidToken creates a valid token that expires in 15 minutes
func CreateValidToken() string {
	return GenerateTestToken("test@example.com", 15*time.Minute)
}

// CreateValidTokenForEmail creates a valid token for a specific email
func CreateValidTokenForEmail(email string) string {
	return GenerateTestToken(email, 15*time.Minute)
}

// CreateRefreshToken creates a refresh token that expires in 7 days
func CreateRefreshToken(email string) string {
	return GenerateTestToken(email, 7*24*time.Hour)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return testSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// ExtractEmailFromToken extracts the email from a JWT token
func ExtractEmailFromToken(tokenString string) string {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return ""
	}

	if email, ok := claims["email"].(string); ok {
		return email
	}

	return ""
}
