// Package jwt provides JSON Web Token (JWT) functionality for authentication.
//
// This package implements JWT token structures, creation, and validation
// for the AINative Code CLI tool. It supports both access tokens and refresh
// tokens with RS256 signing algorithm.
//
// Token Types:
//   - Access Token: Short-lived (24 hours) token for API authentication
//   - Refresh Token: Long-lived (7 days) token for obtaining new access tokens
//
// Features:
//   - RS256 (RSA with SHA-256) signing algorithm
//   - Custom claims for user identity and roles
//   - Token validation with expiration checking
//   - Public key caching for performance
//
// Example usage:
//
//	import (
//	    "github.com/AINative-studio/ainative-code/internal/auth/jwt"
//	    "time"
//	)
//
//	// Create access token
//	claims := &jwt.AccessTokenClaims{
//	    UserID: "user-123",
//	    Email:  "user@example.com",
//	    Roles:  []string{"user"},
//	}
//	token, err := jwt.CreateAccessToken(claims, privateKey)
//
//	// Validate token
//	validatedClaims, err := jwt.ValidateAccessToken(token, publicKey)
package jwt
