package cmd

import (
	"strings"
	"testing"
)

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		// Positive cases - should be sensitive
		{
			name:      "api_key lowercase",
			fieldName: "api_key",
			want:      true,
		},
		{
			name:      "API_KEY uppercase",
			fieldName: "API_KEY",
			want:      true,
		},
		{
			name:      "ApiKey mixed case",
			fieldName: "ApiKey",
			want:      true,
		},
		{
			name:      "token (exact match)",
			fieldName: "token",
			want:      true,
		},
		{
			name:      "access_token",
			fieldName: "access_token",
			want:      true,
		},
		{
			name:      "refresh_token",
			fieldName: "refresh_token",
			want:      true,
		},
		{
			name:      "secret",
			fieldName: "secret",
			want:      true,
		},
		{
			name:      "client_secret",
			fieldName: "client_secret",
			want:      true,
		},
		{
			name:      "password",
			fieldName: "password",
			want:      true,
		},
		{
			name:      "passwd",
			fieldName: "passwd",
			want:      true,
		},
		{
			name:      "pwd",
			fieldName: "pwd",
			want:      true,
		},
		{
			name:      "encryption_key",
			fieldName: "encryption_key",
			want:      true,
		},
		{
			name:      "private_key",
			fieldName: "private_key",
			want:      true,
		},
		{
			name:      "access_key_id",
			fieldName: "access_key_id",
			want:      true,
		},
		{
			name:      "secret_access_key",
			fieldName: "secret_access_key",
			want:      true,
		},
		{
			name:      "session_token",
			fieldName: "session_token",
			want:      true,
		},
		{
			name:      "bearer_token",
			fieldName: "bearer_token",
			want:      true,
		},
		{
			name:      "credential",
			fieldName: "credential",
			want:      true,
		},

		// Negative cases - should NOT be sensitive
		{
			name:      "provider",
			fieldName: "provider",
			want:      false,
		},
		{
			name:      "model",
			fieldName: "model",
			want:      false,
		},
		{
			name:      "endpoint",
			fieldName: "endpoint",
			want:      false,
		},
		{
			name:      "base_url",
			fieldName: "base_url",
			want:      false,
		},
		{
			name:      "temperature",
			fieldName: "temperature",
			want:      false,
		},
		{
			name:      "max_tokens",
			fieldName: "max_tokens",
			want:      false,
		},
		{
			name:      "timeout",
			fieldName: "timeout",
			want:      false,
		},
		{
			name:      "enabled",
			fieldName: "enabled",
			want:      false,
		},
		{
			name:      "verbose",
			fieldName: "verbose",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSensitiveField(tt.fieldName)
			if got != tt.want {
				t.Errorf("isSensitiveField(%q) = %v, want %v", tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "very short value (1 char)",
			value: "a",
			want:  "*",
		},
		{
			name:  "very short value (3 chars)",
			value: "abc",
			want:  "***",
		},
		{
			name:  "short value (8 chars)",
			value: "abcdefgh",
			want:  "abc*****",
		},
		{
			name:  "OpenAI API key format",
			value: "sk-1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  "sk-...123456",
		},
		{
			name:  "Anthropic API key format",
			value: "sk-ant-api03-1234567890abcdefghijklmnopqrstuvwxyz",
			want:  "sk-...uvwxyz",
		},
		{
			name:  "AWS access key format",
			value: "AKIAIOSFODNN7EXAMPLE",
			want:  "AKI...XAMPLE",
		},
		{
			name:  "AWS secret key format",
			value: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			want:  "wJa...PLEKEY",
		},
		{
			name:  "JWT token",
			value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			want:  "eyJ...Qssw5c",
		},
		{
			name:  "minimum length for masking (12 chars)",
			value: "abcdefghijkl",
			want:  "abc...ghijkl",
		},
		{
			name:  "exact boundary (11 chars)",
			value: "abcdefghijk",
			want:  "abc********",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskValue(tt.value)
			if got != tt.want {
				t.Errorf("maskValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, result interface{})
	}{
		{
			name:  "nil input",
			input: nil,
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			},
		},
		{
			name: "simple map with api_key",
			input: map[string]interface{}{
				"provider": "openai",
				"api_key":  "sk-1234567890abcdefghijklmnopqrstuvwxyz123456",
				"model":    "gpt-4",
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}

				// Provider and model should be unchanged
				if m["provider"] != "openai" {
					t.Errorf("provider should be unchanged, got %v", m["provider"])
				}
				if m["model"] != "gpt-4" {
					t.Errorf("model should be unchanged, got %v", m["model"])
				}

				// API key should be masked
				maskedKey, ok := m["api_key"].(string)
				if !ok {
					t.Fatal("api_key is not a string")
				}
				if !strings.Contains(maskedKey, "...") {
					t.Errorf("api_key should be masked, got %v", maskedKey)
				}
				if maskedKey == "sk-1234567890abcdefghijklmnopqrstuvwxyz123456" {
					t.Error("api_key should not be the original value")
				}
			},
		},
		{
			name: "nested map with multiple sensitive fields",
			input: map[string]interface{}{
				"llm": map[string]interface{}{
					"openai": map[string]interface{}{
						"api_key": "sk-openai123456789",
						"model":   "gpt-4",
					},
					"anthropic": map[string]interface{}{
						"api_key": "sk-ant-api03-anthropic123",
						"model":   "claude-3-5-sonnet",
					},
				},
				"platform": map[string]interface{}{
					"authentication": map[string]interface{}{
						"client_secret": "secret12345678901234",
						"client_id":     "client123",
					},
				},
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}

				// Check OpenAI api_key is masked
				llm := m["llm"].(map[string]interface{})
				openai := llm["openai"].(map[string]interface{})
				openaiKey := openai["api_key"].(string)
				if !strings.Contains(openaiKey, "...") {
					t.Errorf("OpenAI api_key should be masked, got %v", openaiKey)
				}

				// Check Anthropic api_key is masked
				anthropic := llm["anthropic"].(map[string]interface{})
				anthropicKey := anthropic["api_key"].(string)
				if !strings.Contains(anthropicKey, "...") {
					t.Errorf("Anthropic api_key should be masked, got %v", anthropicKey)
				}

				// Check client_secret is masked
				platform := m["platform"].(map[string]interface{})
				auth := platform["authentication"].(map[string]interface{})
				clientSecret := auth["client_secret"].(string)
				// The value is "secret12345678901234" which is > 12 chars, should show prefix and suffix
				if !strings.Contains(clientSecret, "...") && !strings.Contains(clientSecret, "*") {
					t.Errorf("client_secret should be masked, got %v", clientSecret)
				}
				if clientSecret == "secret12345678901234" {
					t.Error("client_secret should not be the original value")
				}

				// Check non-sensitive fields are unchanged
				if openai["model"] != "gpt-4" {
					t.Errorf("model should be unchanged, got %v", openai["model"])
				}
			},
		},
		{
			name: "map with empty api_key",
			input: map[string]interface{}{
				"api_key": "",
				"model":   "gpt-4",
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}

				// Empty api_key should remain empty
				if m["api_key"] != "" {
					t.Errorf("empty api_key should remain empty, got %v", m["api_key"])
				}
			},
		},
		{
			name: "map with various sensitive field types",
			input: map[string]interface{}{
				"password":       "mypassword123",
				"token":          "bearer_token_12345",
				"secret":         "secret_value_xyz",
				"encryption_key": "encryption_key_abc",
				"normal_field":   "normal_value",
			},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}

				// All sensitive fields should be masked
				sensitiveFields := []string{"password", "token", "secret", "encryption_key"}
				for _, field := range sensitiveFields {
					value := m[field].(string)
					// Should contain either "..." or "*" for masking
					if !strings.Contains(value, "...") && !strings.Contains(value, "*") {
						t.Errorf("%s should be masked, got %v", field, value)
					}
				}

				// Normal field should be unchanged
				if m["normal_field"] != "normal_value" {
					t.Errorf("normal_field should be unchanged, got %v", m["normal_field"])
				}
			},
		},
		{
			name: "slice of values",
			input: []interface{}{
				"value1",
				"value2",
				"value3",
			},
			check: func(t *testing.T, result interface{}) {
				slice, ok := result.([]interface{})
				if !ok {
					t.Fatal("result is not a slice")
				}

				if len(slice) != 3 {
					t.Errorf("expected slice of length 3, got %d", len(slice))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSensitiveData(tt.input)
			tt.check(t, result)
		})
	}
}

func TestMaskValueEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		validate func(t *testing.T, masked string)
	}{
		{
			name:  "single character",
			value: "x",
			validate: func(t *testing.T, masked string) {
				if masked != "*" {
					t.Errorf("expected *, got %s", masked)
				}
			},
		},
		{
			name:  "two characters",
			value: "xy",
			validate: func(t *testing.T, masked string) {
				if masked != "**" {
					t.Errorf("expected **, got %s", masked)
				}
			},
		},
		{
			name:  "three characters",
			value: "xyz",
			validate: func(t *testing.T, masked string) {
				if masked != "***" {
					t.Errorf("expected ***, got %s", masked)
				}
			},
		},
		{
			name:  "exactly 4 characters (minimum)",
			value: "abcd",
			validate: func(t *testing.T, masked string) {
				if !strings.HasPrefix(masked, "abc") {
					t.Errorf("expected to start with 'abc', got %s", masked)
				}
			},
		},
		{
			name:  "unicode characters",
			value: "sk-日本語-test123456",
			validate: func(t *testing.T, masked string) {
				if masked == "" {
					t.Error("masked value should not be empty")
				}
			},
		},
		{
			name:  "special characters",
			value: "sk-!@#$%^&*()_+-=[]{}|;:,.<>?/~`",
			validate: func(t *testing.T, masked string) {
				if !strings.HasPrefix(masked, "sk-") {
					t.Errorf("expected to start with 'sk-', got %s", masked)
				}
				if masked == "sk-!@#$%^&*()_+-=[]{}|;:,.<>?/~`" {
					t.Error("value should be masked")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked := maskValue(tt.value)
			tt.validate(t, masked)
		})
	}
}

func TestFormatConfigOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		indent int
		check  func(t *testing.T, output string)
	}{
		{
			name:   "simple map",
			input:  map[string]interface{}{"key1": "value1", "key2": "value2"},
			indent: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "key1: value1") {
					t.Error("output should contain key1: value1")
				}
				if !strings.Contains(output, "key2: value2") {
					t.Error("output should contain key2: value2")
				}
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "value",
				},
			},
			indent: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "level1:") {
					t.Error("output should contain level1:")
				}
				if !strings.Contains(output, "level2: value") {
					t.Error("output should contain level2: value")
				}
			},
		},
		{
			name:   "slice",
			input:  []interface{}{"item1", "item2", "item3"},
			indent: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "- item1") {
					t.Error("output should contain - item1")
				}
			},
		},
		{
			name:   "nil value",
			input:  nil,
			indent: 0,
			check: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("expected empty output for nil, got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatConfigOutput(tt.input, tt.indent)
			tt.check(t, output)
		})
	}
}

// BenchmarkMaskValue benchmarks the maskValue function
func BenchmarkMaskValue(b *testing.B) {
	testValue := "sk-1234567890abcdefghijklmnopqrstuvwxyz123456"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = maskValue(testValue)
	}
}

// BenchmarkMaskSensitiveData benchmarks the maskSensitiveData function
func BenchmarkMaskSensitiveData(b *testing.B) {
	testData := map[string]interface{}{
		"llm": map[string]interface{}{
			"openai": map[string]interface{}{
				"api_key": "sk-openai123456789",
				"model":   "gpt-4",
			},
			"anthropic": map[string]interface{}{
				"api_key": "sk-ant-api03-anthropic123",
				"model":   "claude-3-5-sonnet",
			},
		},
		"platform": map[string]interface{}{
			"authentication": map[string]interface{}{
				"client_secret": "secret123456789",
				"client_id":     "client123",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = maskSensitiveData(testData)
	}
}
