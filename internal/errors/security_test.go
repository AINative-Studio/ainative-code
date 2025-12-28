package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityError(t *testing.T) {
	t.Run("creates security error with basic info", func(t *testing.T) {
		err := NewSecurityError(ErrCodeSecurityViolation, "test violation", "api_key")

		assert.NotNil(t, err)
		assert.Equal(t, ErrCodeSecurityViolation, err.Code())
		assert.Contains(t, err.Error(), "test violation")
		assert.Equal(t, "api_key", err.Resource)
		assert.Equal(t, SeverityCritical, err.Severity())
		assert.False(t, err.IsRetryable())
	})
}

func TestNewSecurityViolationError(t *testing.T) {
	t.Run("creates security violation error", func(t *testing.T) {
		err := NewSecurityViolationError("command_execution", "execute", "command execution is disabled")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeSecurityViolation, err.Code())
		assert.Contains(t, err.Error(), "Security violation")
		assert.Contains(t, err.Error(), "command execution is disabled")
		assert.Equal(t, "command_execution", err.Resource)
		assert.Equal(t, "execute", err.Action)
		assert.Contains(t, err.UserMessage(), "not permitted")
	})

	t.Run("includes all context in error message", func(t *testing.T) {
		err := NewSecurityViolationError("file_system", "write", "path is outside allowed directories")

		assert.Contains(t, err.Error(), "file_system")
		assert.Contains(t, err.Error(), "write")
		assert.Contains(t, err.Error(), "path is outside allowed directories")
	})
}

func TestNewInvalidKeyError(t *testing.T) {
	t.Run("creates invalid key error", func(t *testing.T) {
		err := NewInvalidKeyError("API", "key format is invalid")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeSecurityInvalidKey, err.Code())
		assert.Contains(t, err.Error(), "Invalid API key")
		assert.Contains(t, err.Error(), "key format is invalid")
		assert.Equal(t, "API", err.Resource)
		assert.Contains(t, err.UserMessage(), "API key is invalid")
	})

	t.Run("works with different key types", func(t *testing.T) {
		testCases := []struct {
			keyType string
			reason  string
		}{
			{"encryption", "key length must be 32 bytes"},
			{"JWT", "token signature is invalid"},
			{"OAuth", "token has expired"},
		}

		for _, tc := range testCases {
			err := NewInvalidKeyError(tc.keyType, tc.reason)
			assert.Contains(t, err.Error(), tc.keyType)
			assert.Contains(t, err.Error(), tc.reason)
		}
	})
}

func TestNewEncryptionError(t *testing.T) {
	t.Run("creates encryption error", func(t *testing.T) {
		err := NewEncryptionError("cipher initialization failed")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeSecurityEncryption, err.Code())
		assert.Contains(t, err.Error(), "Encryption failed")
		assert.Contains(t, err.Error(), "cipher initialization failed")
		assert.Equal(t, "encryption", err.Resource)
		assert.Contains(t, err.UserMessage(), "Unable to encrypt")
	})

	t.Run("provides user-friendly message", func(t *testing.T) {
		err := NewEncryptionError("internal error")

		userMsg := err.UserMessage()
		assert.NotContains(t, userMsg, "internal error")
		assert.Contains(t, userMsg, "encrypt")
		assert.Contains(t, userMsg, "configuration")
	})
}

func TestNewDecryptionError(t *testing.T) {
	t.Run("creates decryption error", func(t *testing.T) {
		err := NewDecryptionError("invalid ciphertext")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeSecurityDecryption, err.Code())
		assert.Contains(t, err.Error(), "Decryption failed")
		assert.Contains(t, err.Error(), "invalid ciphertext")
		assert.Equal(t, "decryption", err.Resource)
		assert.Contains(t, err.UserMessage(), "Unable to decrypt")
	})

	t.Run("provides helpful user message", func(t *testing.T) {
		err := NewDecryptionError("mac verification failed")

		userMsg := err.UserMessage()
		assert.Contains(t, userMsg, "encryption key may be invalid")
		assert.Contains(t, userMsg, "data may be corrupted")
	})
}

func TestSecurityError_WithAction(t *testing.T) {
	t.Run("adds action to security error", func(t *testing.T) {
		err := NewSecurityError(ErrCodeSecurityViolation, "test error", "resource")
		err = err.WithAction("read")

		assert.Equal(t, "read", err.Action)
	})

	t.Run("chains with error creation", func(t *testing.T) {
		err := NewSecurityError(ErrCodeSecurityViolation, "test", "res").WithAction("write")

		assert.Equal(t, "write", err.Action)
		assert.Equal(t, "res", err.Resource)
	})
}

func TestSecurityError_ErrorCodes(t *testing.T) {
	t.Run("security error codes are defined", func(t *testing.T) {
		assert.Equal(t, ErrorCode("SECURITY_VIOLATION"), ErrCodeSecurityViolation)
		assert.Equal(t, ErrorCode("SECURITY_INVALID_KEY"), ErrCodeSecurityInvalidKey)
		assert.Equal(t, ErrorCode("SECURITY_ENCRYPTION_FAILED"), ErrCodeSecurityEncryption)
		assert.Equal(t, ErrorCode("SECURITY_DECRYPTION_FAILED"), ErrCodeSecurityDecryption)
	})
}

func TestSecurityError_Metadata(t *testing.T) {
	t.Run("can add metadata to security errors", func(t *testing.T) {
		err := NewSecurityError(ErrCodeSecurityViolation, "test", "resource")
		err.WithMetadata("user_id", "user123")
		err.WithMetadata("ip_address", "192.168.1.1")

		metadata := err.Metadata()
		assert.Equal(t, "user123", metadata["user_id"])
		assert.Equal(t, "192.168.1.1", metadata["ip_address"])
	})
}

func TestSecurityError_Severity(t *testing.T) {
	t.Run("all security errors have critical severity", func(t *testing.T) {
		errors := []*SecurityError{
			NewSecurityViolationError("res", "act", "reason"),
			NewInvalidKeyError("type", "reason"),
			NewEncryptionError("reason"),
			NewDecryptionError("reason"),
		}

		for _, err := range errors {
			assert.Equal(t, SeverityCritical, err.Severity(), "error: %v", err)
		}
	})
}

func TestSecurityError_Retryable(t *testing.T) {
	t.Run("security errors are not retryable", func(t *testing.T) {
		errors := []*SecurityError{
			NewSecurityViolationError("res", "act", "reason"),
			NewInvalidKeyError("type", "reason"),
			NewEncryptionError("reason"),
			NewDecryptionError("reason"),
		}

		for _, err := range errors {
			assert.False(t, err.IsRetryable(), "error: %v", err)
		}
	})
}
