package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CreateAccessToken creates a new access token with the given claims.
func CreateAccessToken(userID, email string, roles []string, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	expiresAt := now.Add(AccessTokenDuration)

	claims := &AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID: userID,
		Email:  email,
		Roles:  roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// CreateRefreshToken creates a new refresh token with the given claims.
func CreateRefreshToken(userID, sessionID string, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	expiresAt := now.Add(RefreshTokenDuration)

	claims := &RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:    userID,
		SessionID: sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

// CreateTokenPair creates both access and refresh tokens.
func CreateTokenPair(userID, email string, roles []string, sessionID string, privateKey *rsa.PrivateKey) (*TokenPair, error) {
	accessToken, err := CreateAccessToken(userID, email, roles, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := CreateRefreshToken(userID, sessionID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims.
func ValidateAccessToken(tokenString string, publicKey *rsa.PublicKey) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if token.Method.Alg() != SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate issuer and audience
	if claims.Issuer != Issuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	expectedAudience := false
	for _, aud := range claims.Audience {
		if aud == Audience {
			expectedAudience = true
			break
		}
	}
	if !expectedAudience {
		return nil, fmt.Errorf("invalid audience: %v", claims.Audience)
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims.
func ValidateRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if token.Method.Alg() != SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate issuer and audience
	if claims.Issuer != Issuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	expectedAudience := false
	for _, aud := range claims.Audience {
		if aud == Audience {
			expectedAudience = true
			break
		}
	}
	if !expectedAudience {
		return nil, fmt.Errorf("invalid audience: %v", claims.Audience)
	}

	return claims, nil
}

// ValidateToken performs basic validation on a token string and returns metadata.
func ValidateToken(tokenString string, publicKey *rsa.PublicKey) (*ValidationResult, error) {
	// Try parsing as access token first
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	result := &ValidationResult{
		Valid: false,
	}

	if err != nil {
		result.Error = err
		// Check if it's an expiration error
		if err.Error() == jwt.ErrTokenExpired.Error() ||
		   (token != nil && !token.Valid && claims.ExpiresAt != nil) {
			result.Expired = true
			if claims.ExpiresAt != nil {
				result.ExpiresAt = claims.ExpiresAt.Time
			}
		}
		return result, nil
	}

	if token.Valid {
		result.Valid = true
		result.Claims = claims
		if claims.ExpiresAt != nil {
			result.ExpiresAt = claims.ExpiresAt.Time
		}
	}

	return result, nil
}

// IsTokenExpired checks if a token has expired without full validation.
func IsTokenExpired(tokenString string) (bool, error) {
	// Parse without verification to check expiration
	claims := &AccessTokenClaims{}
	parser := jwt.NewParser()

	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims.ExpiresAt == nil {
		return false, fmt.Errorf("token has no expiration")
	}

	return time.Now().After(claims.ExpiresAt.Time), nil
}

// GetTokenExpiration returns the expiration time of a token without validation.
func GetTokenExpiration(tokenString string) (time.Time, error) {
	claims := &AccessTokenClaims{}
	parser := jwt.NewParser()

	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("token has no expiration")
	}

	return claims.ExpiresAt.Time, nil
}
