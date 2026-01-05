package config

import (
	"github.com/AINative-studio/ainative-code/internal/errors"
)

// DefaultExtendedThinkingConfig returns the default extended thinking configuration
func DefaultExtendedThinkingConfig() *ExtendedThinkingConfig {
	return &ExtendedThinkingConfig{
		Enabled:    true,
		AutoExpand: false,
		MaxDepth:   10,
	}
}

// ValidateExtendedThinkingConfig validates extended thinking configuration
func ValidateExtendedThinkingConfig(cfg *ExtendedThinkingConfig) error {
	if cfg == nil {
		return nil // Optional config
	}

	// Validate max depth
	if cfg.MaxDepth < 1 {
		return errors.NewConfigValidationError("extended_thinking.max_depth", "must be at least 1")
	}

	if cfg.MaxDepth > 100 {
		return errors.NewConfigValidationError("extended_thinking.max_depth", "must not exceed 100")
	}

	return nil
}

// IsExtendedThinkingEnabled checks if extended thinking is enabled in the configuration
func IsExtendedThinkingEnabled(cfg *Config) bool {
	if cfg == nil {
		return false
	}

	// Check Anthropic config for extended thinking
	if cfg.LLM.Anthropic != nil && cfg.LLM.Anthropic.ExtendedThinking != nil {
		return cfg.LLM.Anthropic.ExtendedThinking.Enabled
	}

	return false
}

// GetExtendedThinkingConfig retrieves extended thinking config or returns defaults
func GetExtendedThinkingConfig(cfg *Config) *ExtendedThinkingConfig {
	if cfg == nil {
		return DefaultExtendedThinkingConfig()
	}

	// Check Anthropic config for extended thinking
	if cfg.LLM.Anthropic != nil && cfg.LLM.Anthropic.ExtendedThinking != nil {
		return cfg.LLM.Anthropic.ExtendedThinking
	}

	// Return defaults if not configured
	return DefaultExtendedThinkingConfig()
}

// ShouldAutoExpandThinking checks if thinking blocks should be auto-expanded
func ShouldAutoExpandThinking(cfg *Config) bool {
	thinkingCfg := GetExtendedThinkingConfig(cfg)
	return thinkingCfg.AutoExpand
}

// GetMaxThinkingDepth returns the maximum allowed thinking depth
func GetMaxThinkingDepth(cfg *Config) int {
	thinkingCfg := GetExtendedThinkingConfig(cfg)
	return thinkingCfg.MaxDepth
}
