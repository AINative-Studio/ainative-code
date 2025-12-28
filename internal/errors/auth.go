package errors

import "fmt"

// AuthenticationError represents authentication and authorization errors
type AuthenticationError struct {
	*BaseError
	Provider string
	UserID   string
	Resource string
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(code ErrorCode, message string) *AuthenticationError {
	baseErr := newError(code, message, SeverityHigh, false)
	return &AuthenticationError{
		BaseError: baseErr,
	}
}

// NewAuthFailedError creates an error for general authentication failures
func NewAuthFailedError(provider string, cause error) *AuthenticationError {
	msg := fmt.Sprintf("Authentication failed for provider '%s'", provider)
	userMsg := "Authentication failed. Please check your credentials and try again."

	baseErr := newError(ErrCodeAuthFailed, msg, SeverityHigh, false)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	return &AuthenticationError{
		BaseError: baseErr,
		Provider:  provider,
	}
}

// NewInvalidTokenError creates an error for invalid authentication tokens
func NewInvalidTokenError(provider string) *AuthenticationError {
	msg := fmt.Sprintf("Invalid authentication token for provider '%s'", provider)
	userMsg := "Authentication error: Your authentication token is invalid. Please log in again."

	err := NewAuthenticationError(ErrCodeAuthInvalidToken, msg)
	err.userMsg = userMsg
	err.Provider = provider
	return err
}

// NewExpiredTokenError creates an error for expired authentication tokens
func NewExpiredTokenError(provider string) *AuthenticationError {
	msg := fmt.Sprintf("Authentication token expired for provider '%s'", provider)
	userMsg := "Authentication error: Your session has expired. Please log in again."

	err := NewAuthenticationError(ErrCodeAuthExpiredToken, msg)
	err.userMsg = userMsg
	err.Provider = provider
	err.retryable = true // Token refresh might be possible
	return err
}

// NewPermissionDeniedError creates an error for permission/authorization failures
func NewPermissionDeniedError(resource, action string) *AuthenticationError {
	msg := fmt.Sprintf("Permission denied: insufficient privileges to %s on resource '%s'", action, resource)
	userMsg := fmt.Sprintf("Access denied: You don't have permission to %s this resource.", action)

	err := NewAuthenticationError(ErrCodeAuthPermissionDenied, msg)
	err.userMsg = userMsg
	err.Resource = resource
	return err
}

// NewInvalidCredentialsError creates an error for invalid credentials
func NewInvalidCredentialsError(provider string) *AuthenticationError {
	msg := fmt.Sprintf("Invalid credentials provided for provider '%s'", provider)
	userMsg := "Authentication error: The credentials you provided are incorrect. Please verify and try again."

	err := NewAuthenticationError(ErrCodeAuthInvalidCredentials, msg)
	err.userMsg = userMsg
	err.Provider = provider
	return err
}

// WithProvider sets the authentication provider
func (e *AuthenticationError) WithProvider(provider string) *AuthenticationError {
	e.Provider = provider
	return e
}

// WithUserID sets the user ID associated with the error
func (e *AuthenticationError) WithUserID(userID string) *AuthenticationError {
	e.UserID = userID
	return e
}

// WithResource sets the resource that was being accessed
func (e *AuthenticationError) WithResource(resource string) *AuthenticationError {
	e.Resource = resource
	return e
}
