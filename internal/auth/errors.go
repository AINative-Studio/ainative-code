package auth

import "errors"

// PKCE Errors
//
// Errors related to PKCE (Proof Key for Code Exchange) parameter generation
// and validation.
var (
	// ErrPKCEGeneration indicates failure to generate PKCE parameters.
	// This typically occurs when the cryptographic random number generator fails.
	ErrPKCEGeneration = errors.New("failed to generate PKCE parameters")

	// ErrInvalidCodeVerifier indicates the code verifier format is invalid.
	// Code verifiers must be 43-128 characters using [A-Z]/[a-z]/[0-9]/-/./\_/~
	ErrInvalidCodeVerifier = errors.New("invalid code verifier format")

	// ErrInvalidCodeChallenge indicates the code challenge is invalid.
	// Code challenges must be base64url-encoded SHA-256 hashes without padding.
	ErrInvalidCodeChallenge = errors.New("invalid code challenge")
)

// OAuth Flow Errors
//
// Errors related to the OAuth 2.0 authorization code flow.
var (
	// ErrInvalidState indicates the state parameter validation failed.
	// This is a critical security error indicating a potential CSRF attack.
	ErrInvalidState = errors.New("invalid state parameter (CSRF check failed)")

	// ErrCodeExchangeFailed indicates the authorization code exchange failed.
	// This occurs when the token endpoint rejects the authorization code.
	ErrCodeExchangeFailed = errors.New("failed to exchange code for token")

	// ErrAuthorizationDenied indicates the user denied the authorization request.
	ErrAuthorizationDenied = errors.New("user denied authorization")

	// ErrAuthorizationTimeout indicates the authorization process timed out.
	// This occurs when the user doesn't complete authorization within the timeout period.
	ErrAuthorizationTimeout = errors.New("authorization timeout")

	// ErrMissingAuthorizationCode indicates no authorization code was received.
	ErrMissingAuthorizationCode = errors.New("missing authorization code in callback")
)

// JWT Token Errors
//
// Errors related to JWT token parsing, validation, and verification.
var (
	// ErrTokenExpired indicates the token has expired.
	// Check the "exp" claim and refresh the token if needed.
	ErrTokenExpired = errors.New("token has expired")

	// ErrInvalidSignature indicates the JWT signature verification failed.
	// This is a critical security error indicating potential token forgery.
	ErrInvalidSignature = errors.New("invalid token signature")

	// ErrInvalidClaims indicates required JWT claims are missing or invalid.
	// This includes validation of iss, aud, sub, exp claims.
	ErrInvalidClaims = errors.New("invalid token claims")

	// ErrInvalidIssuer indicates the token issuer doesn't match expected value.
	// Expected issuer: "ainative-auth"
	ErrInvalidIssuer = errors.New("invalid token issuer")

	// ErrInvalidAudience indicates the token audience doesn't match expected value.
	// Expected audience: "ainative-code"
	ErrInvalidAudience = errors.New("invalid token audience")

	// ErrTokenParseFailed indicates JWT token parsing failed.
	// This occurs when the token format is invalid or malformed.
	ErrTokenParseFailed = errors.New("failed to parse JWT token")

	// ErrMissingPublicKey indicates the RSA public key for signature verification is missing.
	ErrMissingPublicKey = errors.New("missing RSA public key for token verification")
)

// Keychain Storage Errors
//
// Errors related to OS keychain operations for secure token storage.
var (
	// ErrKeychainAccess indicates keychain access was denied.
	// This may occur if the user denies keychain access or permissions are insufficient.
	ErrKeychainAccess = errors.New("keychain access denied")

	// ErrKeychainStore indicates storing data in the keychain failed.
	ErrKeychainStore = errors.New("failed to store data in keychain")

	// ErrKeychainRetrieve indicates retrieving data from the keychain failed.
	ErrKeychainRetrieve = errors.New("failed to retrieve data from keychain")

	// ErrKeychainDelete indicates deleting data from the keychain failed.
	ErrKeychainDelete = errors.New("failed to delete data from keychain")

	// ErrKeychainNotFound indicates the requested item was not found in the keychain.
	ErrKeychainNotFound = errors.New("item not found in keychain")
)

// Callback Server Errors
//
// Errors related to the local OAuth callback server.
var (
	// ErrCallbackServerStart indicates the callback server failed to start.
	// This typically occurs when the port is already in use.
	ErrCallbackServerStart = errors.New("failed to start callback server")

	// ErrCallbackServerShutdown indicates the callback server failed to shutdown gracefully.
	ErrCallbackServerShutdown = errors.New("failed to shutdown callback server")

	// ErrCallbackTimeout indicates the callback was not received within the timeout period.
	ErrCallbackTimeout = errors.New("callback timeout")

	// ErrCallbackError indicates the callback contained an error response.
	ErrCallbackError = errors.New("callback error response received")
)

// HTTP Client Errors
//
// Errors related to HTTP requests to OAuth endpoints.
var (
	// ErrHTTPRequest indicates an HTTP request failed.
	ErrHTTPRequest = errors.New("HTTP request failed")

	// ErrHTTPResponse indicates the HTTP response status is not successful (not 2xx).
	ErrHTTPResponse = errors.New("HTTP response error")

	// ErrHTTPTimeout indicates the HTTP request timed out.
	ErrHTTPTimeout = errors.New("HTTP request timeout")

	// ErrResponseParseFailed indicates parsing the HTTP response body failed.
	ErrResponseParseFailed = errors.New("failed to parse HTTP response")
)

// Browser Errors
//
// Errors related to opening the browser for user authorization.
var (
	// ErrBrowserOpen indicates opening the browser failed.
	// This may occur if no browser is available or the command fails.
	ErrBrowserOpen = errors.New("failed to open browser")
)

// Configuration Errors
//
// Errors related to client configuration and validation.
var (
	// ErrMissingClientID indicates the OAuth client ID is missing.
	ErrMissingClientID = errors.New("missing OAuth client ID")

	// ErrMissingAuthEndpoint indicates the authorization endpoint URL is missing.
	ErrMissingAuthEndpoint = errors.New("missing authorization endpoint")

	// ErrMissingTokenEndpoint indicates the token endpoint URL is missing.
	ErrMissingTokenEndpoint = errors.New("missing token endpoint")

	// ErrMissingRedirectURI indicates the redirect URI is missing.
	ErrMissingRedirectURI = errors.New("missing redirect URI")

	// ErrInvalidURL indicates a URL is malformed or invalid.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrInvalidPort indicates the callback port is invalid.
	ErrInvalidPort = errors.New("invalid callback port")
)

// Validation Errors
//
// Errors related to parameter validation.
var (
	// ErrEmptyUserID indicates the user ID is empty.
	ErrEmptyUserID = errors.New("empty user ID")

	// ErrEmptySessionID indicates the session ID is empty.
	ErrEmptySessionID = errors.New("empty session ID")

	// ErrNilContext indicates a nil context was provided.
	ErrNilContext = errors.New("nil context provided")

	// ErrNilRequest indicates a nil request was provided.
	ErrNilRequest = errors.New("nil request provided")
)
