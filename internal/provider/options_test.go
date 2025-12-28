package provider

import (
	"testing"
)

func TestDefaultChatOptions(t *testing.T) {
	opts := DefaultChatOptions()

	if opts == nil {
		t.Fatal("DefaultChatOptions returned nil")
	}

	// Verify defaults
	if opts.MaxTokens != 1024 {
		t.Errorf("expected MaxTokens 1024, got: %d", opts.MaxTokens)
	}

	if opts.Temperature != 0.7 {
		t.Errorf("expected Temperature 0.7, got: %f", opts.Temperature)
	}

	if opts.TopP != 1.0 {
		t.Errorf("expected TopP 1.0, got: %f", opts.TopP)
	}

	if opts.Stream {
		t.Error("expected Stream false, got true")
	}

	if opts.Metadata == nil {
		t.Error("expected Metadata initialized, got nil")
	}

	if len(opts.Metadata) != 0 {
		t.Errorf("expected empty Metadata, got: %v", opts.Metadata)
	}

	if opts.Model != "" {
		t.Errorf("expected empty Model, got: %s", opts.Model)
	}

	if opts.SystemPrompt != "" {
		t.Errorf("expected empty SystemPrompt, got: %s", opts.SystemPrompt)
	}

	if len(opts.StopSequences) != 0 {
		t.Errorf("expected empty StopSequences, got: %v", opts.StopSequences)
	}
}

func TestWithModel(t *testing.T) {
	opts := &ChatOptions{}
	option := WithModel("gpt-4")
	option(opts)

	if opts.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got: %s", opts.Model)
	}

	// Test empty model
	opts2 := &ChatOptions{}
	WithModel("")(opts2)
	if opts2.Model != "" {
		t.Errorf("expected empty Model, got: %s", opts2.Model)
	}
}

func TestWithMaxTokens(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
	}{
		{"positive value", 2048},
		{"zero", 0},
		{"large value", 100000},
		{"negative value", -1}, // Implementation may or may not validate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			option := WithMaxTokens(tt.maxTokens)
			option(opts)

			if opts.MaxTokens != tt.maxTokens {
				t.Errorf("expected MaxTokens %d, got: %d", tt.maxTokens, opts.MaxTokens)
			}
		})
	}
}

func TestWithTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
	}{
		{"zero temperature", 0.0},
		{"low temperature", 0.3},
		{"default temperature", 0.7},
		{"high temperature", 1.0},
		{"above one", 1.5}, // May be valid for some providers
		{"negative", -0.5}, // May be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			option := WithTemperature(tt.temperature)
			option(opts)

			if opts.Temperature != tt.temperature {
				t.Errorf("expected Temperature %f, got: %f", tt.temperature, opts.Temperature)
			}
		})
	}
}

func TestWithTopP(t *testing.T) {
	tests := []struct {
		name string
		topP float64
	}{
		{"zero", 0.0},
		{"low value", 0.1},
		{"mid value", 0.5},
		{"default", 1.0},
		{"above one", 1.5}, // May be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			option := WithTopP(tt.topP)
			option(opts)

			if opts.TopP != tt.topP {
				t.Errorf("expected TopP %f, got: %f", tt.topP, opts.TopP)
			}
		})
	}
}

func TestWithStopSequences(t *testing.T) {
	tests := []struct {
		name      string
		sequences []string
	}{
		{"empty", []string{}},
		{"single sequence", []string{"STOP"}},
		{"multiple sequences", []string{"END", "STOP", "###"}},
		{"nil sequences", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			option := WithStopSequences(tt.sequences...)
			option(opts)

			if len(opts.StopSequences) != len(tt.sequences) {
				t.Errorf("expected %d sequences, got: %d", len(tt.sequences), len(opts.StopSequences))
			}

			for i, seq := range tt.sequences {
				if opts.StopSequences[i] != seq {
					t.Errorf("sequence %d: expected %q, got: %q", i, seq, opts.StopSequences[i])
				}
			}
		})
	}
}

func TestWithSystemPrompt(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
	}{
		{"simple prompt", "You are a helpful assistant."},
		{"empty prompt", ""},
		{"long prompt", "You are a helpful assistant with extensive knowledge in multiple domains..."},
		{"prompt with newlines", "You are helpful.\n\nBe concise."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			option := WithSystemPrompt(tt.prompt)
			option(opts)

			if opts.SystemPrompt != tt.prompt {
				t.Errorf("expected SystemPrompt %q, got: %q", tt.prompt, opts.SystemPrompt)
			}
		})
	}
}

func TestWithMetadata(t *testing.T) {
	// Single metadata entry
	opts := &ChatOptions{}
	option := WithMetadata("user_id", "12345")
	option(opts)

	if opts.Metadata == nil {
		t.Fatal("Metadata map not initialized")
	}

	if opts.Metadata["user_id"] != "12345" {
		t.Errorf("expected Metadata['user_id'] = '12345', got: %s", opts.Metadata["user_id"])
	}

	// Add multiple metadata entries
	opts2 := &ChatOptions{}
	WithMetadata("key1", "value1")(opts2)
	WithMetadata("key2", "value2")(opts2)
	WithMetadata("key3", "value3")(opts2)

	if len(opts2.Metadata) != 3 {
		t.Errorf("expected 3 metadata entries, got: %d", len(opts2.Metadata))
	}

	expectedMetadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, expectedValue := range expectedMetadata {
		if opts2.Metadata[key] != expectedValue {
			t.Errorf("Metadata[%q]: expected %q, got: %q", key, expectedValue, opts2.Metadata[key])
		}
	}

	// Test metadata overwrite
	opts3 := &ChatOptions{}
	WithMetadata("key", "value1")(opts3)
	WithMetadata("key", "value2")(opts3)

	if opts3.Metadata["key"] != "value2" {
		t.Errorf("expected metadata to be overwritten to 'value2', got: %s", opts3.Metadata["key"])
	}
}

func TestApplyChatOptions(t *testing.T) {
	opts := &ChatOptions{}

	// Apply multiple options
	ApplyChatOptions(opts,
		WithModel("claude-3-sonnet"),
		WithMaxTokens(2048),
		WithTemperature(0.8),
		WithTopP(0.9),
		WithStopSequences("END", "STOP"),
		WithSystemPrompt("You are helpful"),
		WithMetadata("session_id", "abc123"),
	)

	// Verify all options were applied
	if opts.Model != "claude-3-sonnet" {
		t.Errorf("Model: expected 'claude-3-sonnet', got: %s", opts.Model)
	}

	if opts.MaxTokens != 2048 {
		t.Errorf("MaxTokens: expected 2048, got: %d", opts.MaxTokens)
	}

	if opts.Temperature != 0.8 {
		t.Errorf("Temperature: expected 0.8, got: %f", opts.Temperature)
	}

	if opts.TopP != 0.9 {
		t.Errorf("TopP: expected 0.9, got: %f", opts.TopP)
	}

	if len(opts.StopSequences) != 2 {
		t.Errorf("StopSequences: expected 2, got: %d", len(opts.StopSequences))
	}

	if opts.SystemPrompt != "You are helpful" {
		t.Errorf("SystemPrompt: expected 'You are helpful', got: %s", opts.SystemPrompt)
	}

	if opts.Metadata["session_id"] != "abc123" {
		t.Errorf("Metadata: expected session_id='abc123', got: %s", opts.Metadata["session_id"])
	}

	if opts.Stream {
		t.Error("Stream should be false for ChatOptions")
	}
}

func TestApplyChatOptions_Empty(t *testing.T) {
	opts := &ChatOptions{}
	ApplyChatOptions(opts)

	// Should not panic with no options
	if opts.Model != "" {
		t.Errorf("expected empty options, got Model: %s", opts.Model)
	}
}

func TestApplyChatOptions_Order(t *testing.T) {
	// Test that later options override earlier ones
	opts := &ChatOptions{}

	ApplyChatOptions(opts,
		WithModel("gpt-4"),
		WithMaxTokens(1024),
		WithModel("claude-3-opus"), // Override
		WithMaxTokens(2048),        // Override
	)

	if opts.Model != "claude-3-opus" {
		t.Errorf("expected last Model to win, got: %s", opts.Model)
	}

	if opts.MaxTokens != 2048 {
		t.Errorf("expected last MaxTokens to win, got: %d", opts.MaxTokens)
	}
}

func TestStreamWithModel(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithModel("gpt-4")
	option(opts)

	if opts.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got: %s", opts.Model)
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithMaxTokens(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithMaxTokens(2048)
	option(opts)

	if opts.MaxTokens != 2048 {
		t.Errorf("expected MaxTokens 2048, got: %d", opts.MaxTokens)
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithTemperature(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithTemperature(0.5)
	option(opts)

	if opts.Temperature != 0.5 {
		t.Errorf("expected Temperature 0.5, got: %f", opts.Temperature)
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithTopP(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithTopP(0.9)
	option(opts)

	if opts.TopP != 0.9 {
		t.Errorf("expected TopP 0.9, got: %f", opts.TopP)
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithStopSequences(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithStopSequences("END", "STOP")
	option(opts)

	if len(opts.StopSequences) != 2 {
		t.Errorf("expected 2 stop sequences, got: %d", len(opts.StopSequences))
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithSystemPrompt(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithSystemPrompt("You are helpful")
	option(opts)

	if opts.SystemPrompt != "You are helpful" {
		t.Errorf("expected SystemPrompt 'You are helpful', got: %s", opts.SystemPrompt)
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestStreamWithMetadata(t *testing.T) {
	opts := &ChatOptions{}
	option := StreamWithMetadata("request_id", "xyz789")
	option(opts)

	if opts.Metadata == nil {
		t.Fatal("Metadata map not initialized")
	}

	if opts.Metadata["request_id"] != "xyz789" {
		t.Errorf("expected Metadata['request_id'] = 'xyz789', got: %s", opts.Metadata["request_id"])
	}

	if !opts.Stream {
		t.Error("expected Stream true, got false")
	}
}

func TestApplyStreamOptions(t *testing.T) {
	opts := &ChatOptions{}

	// Apply multiple stream options
	ApplyStreamOptions(opts,
		StreamWithModel("gpt-4-turbo"),
		StreamWithMaxTokens(4096),
		StreamWithTemperature(0.6),
		StreamWithTopP(0.95),
		StreamWithStopSequences("###"),
		StreamWithSystemPrompt("Be concise"),
		StreamWithMetadata("stream_id", "stream123"),
	)

	// Verify all options were applied
	if opts.Model != "gpt-4-turbo" {
		t.Errorf("Model: expected 'gpt-4-turbo', got: %s", opts.Model)
	}

	if opts.MaxTokens != 4096 {
		t.Errorf("MaxTokens: expected 4096, got: %d", opts.MaxTokens)
	}

	if opts.Temperature != 0.6 {
		t.Errorf("Temperature: expected 0.6, got: %f", opts.Temperature)
	}

	if opts.TopP != 0.95 {
		t.Errorf("TopP: expected 0.95, got: %f", opts.TopP)
	}

	if len(opts.StopSequences) != 1 || opts.StopSequences[0] != "###" {
		t.Errorf("StopSequences: expected ['###'], got: %v", opts.StopSequences)
	}

	if opts.SystemPrompt != "Be concise" {
		t.Errorf("SystemPrompt: expected 'Be concise', got: %s", opts.SystemPrompt)
	}

	if opts.Metadata["stream_id"] != "stream123" {
		t.Errorf("Metadata: expected stream_id='stream123', got: %s", opts.Metadata["stream_id"])
	}

	if !opts.Stream {
		t.Error("Stream should be true for StreamOptions")
	}
}

func TestApplyStreamOptions_SetsStreamFlag(t *testing.T) {
	// Test that ApplyStreamOptions sets Stream=true even with no options
	opts := &ChatOptions{Stream: false}
	ApplyStreamOptions(opts)

	if !opts.Stream {
		t.Error("ApplyStreamOptions should set Stream=true")
	}
}

func TestApplyStreamOptions_Empty(t *testing.T) {
	opts := &ChatOptions{}
	ApplyStreamOptions(opts)

	// Should not panic with no options
	if !opts.Stream {
		t.Error("Stream should be true even with no options")
	}
}

func TestMixedChatAndStreamOptions(t *testing.T) {
	// Test that you can use Chat options but still enable streaming manually
	opts := &ChatOptions{}

	ApplyChatOptions(opts,
		WithModel("gpt-4"),
		WithMaxTokens(2048),
	)

	// Manually enable streaming
	opts.Stream = true

	if !opts.Stream {
		t.Error("expected Stream true")
	}

	if opts.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got: %s", opts.Model)
	}
}

func TestOptionComposition(t *testing.T) {
	// Test that options compose correctly with defaults
	opts := DefaultChatOptions()

	// Override some defaults
	ApplyChatOptions(opts,
		WithModel("custom-model"),
		WithTemperature(0.9),
	)

	// Check overridden values
	if opts.Model != "custom-model" {
		t.Errorf("Model: expected 'custom-model', got: %s", opts.Model)
	}

	if opts.Temperature != 0.9 {
		t.Errorf("Temperature: expected 0.9, got: %f", opts.Temperature)
	}

	// Check that other defaults remain
	if opts.MaxTokens != 1024 {
		t.Errorf("MaxTokens should retain default 1024, got: %d", opts.MaxTokens)
	}

	if opts.TopP != 1.0 {
		t.Errorf("TopP should retain default 1.0, got: %f", opts.TopP)
	}

	if opts.Stream {
		t.Error("Stream should remain false")
	}
}

func TestOptionsImmutability(t *testing.T) {
	// Test that applying options to one ChatOptions doesn't affect another
	opts1 := &ChatOptions{}
	opts2 := &ChatOptions{}

	WithModel("model1")(opts1)
	WithModel("model2")(opts2)

	if opts1.Model != "model1" {
		t.Errorf("opts1.Model: expected 'model1', got: %s", opts1.Model)
	}

	if opts2.Model != "model2" {
		t.Errorf("opts2.Model: expected 'model2', got: %s", opts2.Model)
	}
}

func TestMetadataInitialization(t *testing.T) {
	// Test that metadata is initialized on first use
	opts := &ChatOptions{}

	// Metadata should be nil initially
	if opts.Metadata != nil {
		t.Error("Metadata should be nil before first WithMetadata call")
	}

	// First WithMetadata call should initialize the map
	WithMetadata("key", "value")(opts)

	if opts.Metadata == nil {
		t.Fatal("Metadata should be initialized after WithMetadata call")
	}

	if len(opts.Metadata) != 1 {
		t.Errorf("expected 1 metadata entry, got: %d", len(opts.Metadata))
	}
}

func TestStreamMetadataInitialization(t *testing.T) {
	// Test that metadata is initialized on first use for stream options
	opts := &ChatOptions{}

	// Metadata should be nil initially
	if opts.Metadata != nil {
		t.Error("Metadata should be nil before first StreamWithMetadata call")
	}

	// First StreamWithMetadata call should initialize the map
	StreamWithMetadata("key", "value")(opts)

	if opts.Metadata == nil {
		t.Fatal("Metadata should be initialized after StreamWithMetadata call")
	}

	if len(opts.Metadata) != 1 {
		t.Errorf("expected 1 metadata entry, got: %d", len(opts.Metadata))
	}

	if !opts.Stream {
		t.Error("Stream should be true")
	}
}

func TestAllStreamOptionsFlagStream(t *testing.T) {
	// Verify that every Stream* option sets Stream=true
	streamOptions := []struct {
		name   string
		option StreamOption
	}{
		{"StreamWithModel", StreamWithModel("test")},
		{"StreamWithMaxTokens", StreamWithMaxTokens(100)},
		{"StreamWithTemperature", StreamWithTemperature(0.5)},
		{"StreamWithTopP", StreamWithTopP(0.9)},
		{"StreamWithStopSequences", StreamWithStopSequences("END")},
		{"StreamWithSystemPrompt", StreamWithSystemPrompt("test")},
		{"StreamWithMetadata", StreamWithMetadata("key", "value")},
	}

	for _, tt := range streamOptions {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			tt.option(opts)

			if !opts.Stream {
				t.Errorf("%s should set Stream=true", tt.name)
			}
		})
	}
}

func TestNoChatOptionSetsStream(t *testing.T) {
	// Verify that regular Chat options do NOT set Stream=true
	chatOptions := []struct {
		name   string
		option ChatOption
	}{
		{"WithModel", WithModel("test")},
		{"WithMaxTokens", WithMaxTokens(100)},
		{"WithTemperature", WithTemperature(0.5)},
		{"WithTopP", WithTopP(0.9)},
		{"WithStopSequences", WithStopSequences("END")},
		{"WithSystemPrompt", WithSystemPrompt("test")},
		{"WithMetadata", WithMetadata("key", "value")},
	}

	for _, tt := range chatOptions {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ChatOptions{}
			tt.option(opts)

			if opts.Stream {
				t.Errorf("%s should NOT set Stream=true", tt.name)
			}
		})
	}
}
