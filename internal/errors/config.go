package errors

import "fmt"

// ConfigError represents configuration-related errors
type ConfigError struct {
	*BaseError
	ConfigKey  string
	ConfigPath string
}

// NewConfigError creates a new configuration error
func NewConfigError(code ErrorCode, message string, configKey string) *ConfigError {
	baseErr := newError(code, message, SeverityHigh, false)
	return &ConfigError{
		BaseError: baseErr,
		ConfigKey: configKey,
	}
}

// NewConfigInvalidError creates an error for invalid configuration
func NewConfigInvalidError(configKey, reason string) *ConfigError {
	msg := fmt.Sprintf("Invalid configuration for '%s': %s", configKey, reason)
	userMsg := fmt.Sprintf("Configuration error: %s is not properly configured. %s", configKey, reason)

	err := NewConfigError(ErrCodeConfigInvalid, msg, configKey)
	err.userMsg = userMsg
	return err
}

// NewConfigMissingError creates an error for missing configuration
func NewConfigMissingError(configKey string) *ConfigError {
	msg := fmt.Sprintf("Required configuration '%s' is missing", configKey)
	userMsg := fmt.Sprintf("Configuration error: Required setting '%s' is not configured. Please check your configuration file.", configKey)

	err := NewConfigError(ErrCodeConfigMissing, msg, configKey)
	err.userMsg = userMsg
	return err
}

// NewConfigParseError creates an error for configuration parsing failures
func NewConfigParseError(configPath string, cause error) *ConfigError {
	msg := fmt.Sprintf("Failed to parse configuration file: %s", configPath)
	userMsg := "Configuration error: Unable to parse the configuration file. Please check the file format and syntax."

	baseErr := newError(ErrCodeConfigParse, msg, SeverityCritical, false)
	baseErr.cause = cause
	baseErr.userMsg = userMsg

	err := &ConfigError{
		BaseError:  baseErr,
		ConfigPath: configPath,
	}
	return err
}

// NewConfigValidationError creates an error for configuration validation failures
func NewConfigValidationError(configKey, validationRule string) *ConfigError {
	msg := fmt.Sprintf("Configuration validation failed for '%s': %s", configKey, validationRule)
	userMsg := fmt.Sprintf("Configuration error: The value for '%s' does not meet the required criteria: %s", configKey, validationRule)

	err := NewConfigError(ErrCodeConfigValidation, msg, configKey)
	err.userMsg = userMsg
	return err
}

// WithPath adds the configuration file path to the error
func (e *ConfigError) WithPath(path string) *ConfigError {
	e.ConfigPath = path
	return e
}
