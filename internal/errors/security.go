package errors

import "fmt"

// SecurityError represents security-related errors
type SecurityError struct {
	*BaseError
	Resource string
	Action   string
}

// NewSecurityError creates a new security error
func NewSecurityError(code ErrorCode, message string, resource string) *SecurityError {
	baseErr := newError(code, message, SeverityCritical, false)
	return &SecurityError{
		BaseError: baseErr,
		Resource:  resource,
	}
}

// NewSecurityViolationError creates an error for security policy violations
func NewSecurityViolationError(resource, action, reason string) *SecurityError {
	msg := fmt.Sprintf("Security violation: %s - %s not allowed for %s", reason, action, resource)
	userMsg := fmt.Sprintf("Security error: The requested operation is not permitted. %s", reason)

	err := NewSecurityError(ErrCodeSecurityViolation, msg, resource)
	err.Action = action
	err.userMsg = userMsg
	return err
}

// NewInvalidKeyError creates an error for invalid encryption/API keys
func NewInvalidKeyError(keyType, reason string) *SecurityError {
	msg := fmt.Sprintf("Invalid %s key: %s", keyType, reason)
	userMsg := fmt.Sprintf("Security error: The %s key is invalid. %s", keyType, reason)

	err := NewSecurityError(ErrCodeSecurityInvalidKey, msg, keyType)
	err.userMsg = userMsg
	return err
}

// NewEncryptionError creates an error for encryption failures
func NewEncryptionError(reason string) *SecurityError {
	msg := fmt.Sprintf("Encryption failed: %s", reason)
	userMsg := "Security error: Unable to encrypt sensitive data. Please check your encryption configuration."

	err := NewSecurityError(ErrCodeSecurityEncryption, msg, "encryption")
	err.userMsg = userMsg
	return err
}

// NewDecryptionError creates an error for decryption failures
func NewDecryptionError(reason string) *SecurityError {
	msg := fmt.Sprintf("Decryption failed: %s", reason)
	userMsg := "Security error: Unable to decrypt data. The encryption key may be invalid or the data may be corrupted."

	err := NewSecurityError(ErrCodeSecurityDecryption, msg, "decryption")
	err.userMsg = userMsg
	return err
}

// WithAction adds the action being attempted to the error
func (e *SecurityError) WithAction(action string) *SecurityError {
	e.Action = action
	return e
}
