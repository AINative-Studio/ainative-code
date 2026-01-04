package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ParseAccessToken parses and validates a JWT access token.
//
// This function:
//   - Parses the JWT token string
//   - Verifies the RS256 signature using the provided public key
//   - Validates standard claims (iss, aud, sub, exp)
//   - Extracts custom claims (email, roles)
//   - Returns a populated AccessToken struct
//
// The token must:
//   - Have a valid RS256 signature
//   - Not be expired (exp claim)
//   - Have issuer = "ainative-auth"
//   - Have audience = "ainative-code"
//   - Contain required claims: sub, email
//
// Returns:
//   - ErrMissingPublicKey if publicKey is nil
//   - ErrTokenParseFailed if token parsing fails
//   - ErrInvalidSignature if signature verification fails
//   - ErrInvalidClaims if required claims are missing or invalid
//   - ErrInvalidIssuer if issuer doesn't match "ainative-auth"
//   - ErrInvalidAudience if audience doesn't match "ainative-code"
//   - ErrTokenExpired if token has expired
//
// Example:
//
//	token, err := ParseAccessToken(tokenString, publicKey)
//	if err != nil {
//	    return fmt.Errorf("invalid access token: %w", err)
//	}
//	fmt.Printf("Authenticated user: %s\n", token.Email)
func ParseAccessToken(tokenString string, publicKey *rsa.PublicKey) (*AccessToken, error) {
	if publicKey == nil {
		return nil, ErrMissingPublicKey
	}

	// Parse token with RS256 signature verification
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v",
				ErrInvalidSignature, token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		// Check if error is due to expiration
		if jwt.ErrTokenExpired.Error() == err.Error() {
			return nil, fmt.Errorf("%w: %v", ErrTokenExpired, err)
		}
		// Check if error is due to signature verification
		if jwt.ErrSignatureInvalid.Error() == err.Error() {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenParseFailed, err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: failed to extract claims", ErrInvalidClaims)
	}

	// Extract and validate issuer
	issuer, err := claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("%w: missing issuer claim: %v", ErrInvalidClaims, err)
	}
	if issuer != "ainative-auth" {
		return nil, fmt.Errorf("%w: expected 'ainative-auth', got '%s'",
			ErrInvalidIssuer, issuer)
	}

	// Extract and validate audience
	audience, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("%w: missing audience claim: %v", ErrInvalidClaims, err)
	}
	if len(audience) == 0 || audience[0] != "ainative-code" {
		return nil, fmt.Errorf("%w: expected 'ainative-code', got '%v'",
			ErrInvalidAudience, audience)
	}

	// Extract subject (user ID)
	subject, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("%w: missing subject claim: %v", ErrInvalidClaims, err)
	}
	if subject == "" {
		return nil, fmt.Errorf("%w: empty subject claim", ErrInvalidClaims)
	}

	// Extract expiration time
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("%w: missing expiration claim: %v", ErrInvalidClaims, err)
	}
	if expiresAt == nil {
		return nil, fmt.Errorf("%w: expiration claim is nil", ErrInvalidClaims)
	}

	// Extract email (custom claim)
	email, err := extractStringClaim(claims, "email")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidClaims, err)
	}

	// Extract roles (custom claim, optional)
	roles, err := extractStringSliceClaim(claims, "roles")
	if err != nil {
		// Roles are optional, default to empty slice
		roles = []string{}
	}

	return &AccessToken{
		Raw:       tokenString,
		ExpiresAt: expiresAt.Time,
		UserID:    subject,
		Email:     email,
		Roles:     roles,
		Issuer:    issuer,
		Audience:  audience[0],
	}, nil
}

// ParseRefreshToken parses and validates a JWT refresh token.
//
// This function:
//   - Parses the JWT token string
//   - Verifies the RS256 signature using the provided public key
//   - Validates standard claims (iss, aud, sub, exp)
//   - Extracts custom claims (session_id)
//   - Returns a populated RefreshToken struct
//
// The token must:
//   - Have a valid RS256 signature
//   - Not be expired (exp claim)
//   - Have issuer = "ainative-auth"
//   - Have audience = "ainative-code"
//   - Contain required claims: sub, session_id
//
// Returns:
//   - ErrMissingPublicKey if publicKey is nil
//   - ErrTokenParseFailed if token parsing fails
//   - ErrInvalidSignature if signature verification fails
//   - ErrInvalidClaims if required claims are missing or invalid
//   - ErrInvalidIssuer if issuer doesn't match "ainative-auth"
//   - ErrInvalidAudience if audience doesn't match "ainative-code"
//   - ErrTokenExpired if token has expired
//
// Example:
//
//	token, err := ParseRefreshToken(tokenString, publicKey)
//	if err != nil {
//	    return fmt.Errorf("invalid refresh token: %w", err)
//	}
//	fmt.Printf("Session ID: %s\n", token.SessionID)
func ParseRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*RefreshToken, error) {
	if publicKey == nil {
		return nil, ErrMissingPublicKey
	}

	// Parse token with RS256 signature verification
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v",
				ErrInvalidSignature, token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		// Check if error is due to expiration
		if jwt.ErrTokenExpired.Error() == err.Error() {
			return nil, fmt.Errorf("%w: %v", ErrTokenExpired, err)
		}
		// Check if error is due to signature verification
		if jwt.ErrSignatureInvalid.Error() == err.Error() {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenParseFailed, err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: failed to extract claims", ErrInvalidClaims)
	}

	// Extract and validate issuer
	issuer, err := claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("%w: missing issuer claim: %v", ErrInvalidClaims, err)
	}
	if issuer != "ainative-auth" {
		return nil, fmt.Errorf("%w: expected 'ainative-auth', got '%s'",
			ErrInvalidIssuer, issuer)
	}

	// Extract and validate audience
	audience, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("%w: missing audience claim: %v", ErrInvalidClaims, err)
	}
	if len(audience) == 0 || audience[0] != "ainative-code" {
		return nil, fmt.Errorf("%w: expected 'ainative-code', got '%v'",
			ErrInvalidAudience, audience)
	}

	// Extract subject (user ID)
	subject, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("%w: missing subject claim: %v", ErrInvalidClaims, err)
	}
	if subject == "" {
		return nil, fmt.Errorf("%w: empty subject claim", ErrInvalidClaims)
	}

	// Extract expiration time
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("%w: missing expiration claim: %v", ErrInvalidClaims, err)
	}
	if expiresAt == nil {
		return nil, fmt.Errorf("%w: expiration claim is nil", ErrInvalidClaims)
	}

	// Extract session_id (custom claim, required for refresh tokens)
	sessionID, err := extractStringClaim(claims, "session_id")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidClaims, err)
	}

	return &RefreshToken{
		Raw:       tokenString,
		ExpiresAt: expiresAt.Time,
		UserID:    subject,
		SessionID: sessionID,
		Issuer:    issuer,
		Audience:  audience[0],
	}, nil
}

// extractStringClaim extracts a string value from JWT claims.
//
// Returns an error if the claim is missing or not a string.
func extractStringClaim(claims jwt.MapClaims, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing claim: %s", key)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not a string: got %T", key, value)
	}

	if strValue == "" {
		return "", fmt.Errorf("claim %s is empty", key)
	}

	return strValue, nil
}

// extractStringSliceClaim extracts a string slice value from JWT claims.
//
// Returns an error if the claim exists but is not a string slice.
// Returns an empty slice if the claim is missing (optional claim).
func extractStringSliceClaim(claims jwt.MapClaims, key string) ([]string, error) {
	value, ok := claims[key]
	if !ok {
		// Claim is optional, return empty slice
		return []string{}, nil
	}

	// Handle case where value is a single string
	if strValue, ok := value.(string); ok {
		return []string{strValue}, nil
	}

	// Handle case where value is a slice
	slice, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("claim %s is not a string or slice: got %T", key, value)
	}

	result := make([]string, 0, len(slice))
	for i, item := range slice {
		strItem, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("claim %s[%d] is not a string: got %T", key, i, item)
		}
		result = append(result, strItem)
	}

	return result, nil
}
