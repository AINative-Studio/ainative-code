package providers

import (
	"testing"
)

func TestWithMaxTokens(t *testing.T) {
	tests := []struct {
		name     string
		tokens   int
		expected int
	}{
		{"Set to 1024", 1024, 1024},
		{"Set to 4096", 4096, 4096},
		{"Set to zero", 0, 0},
		{"Set to negative", -1, -1}, // Allow negative for validation testing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RequestOptions{}
			option := WithMaxTokens(tt.tokens)
			option(opts)

			if opts.MaxTokens == nil {
				t.Errorf("WithMaxTokens did not set MaxTokens")
			} else if *opts.MaxTokens != tt.expected {
				t.Errorf("MaxTokens = %v, want %v", *opts.MaxTokens, tt.expected)
			}
		})
	}
}

func TestWithTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		expected    float64
	}{
		{"Set to 0.7", 0.7, 0.7},
		{"Set to 0.0", 0.0, 0.0},
		{"Set to 1.0", 1.0, 1.0},
		{"Set to 0.5", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RequestOptions{}
			option := WithTemperature(tt.temperature)
			option(opts)

			if opts.Temperature == nil {
				t.Errorf("WithTemperature did not set Temperature")
			} else if *opts.Temperature != tt.expected {
				t.Errorf("Temperature = %v, want %v", *opts.Temperature, tt.expected)
			}
		})
	}
}

func TestWithTopP(t *testing.T) {
	tests := []struct {
		name     string
		topP     float64
		expected float64
	}{
		{"Set to 0.9", 0.9, 0.9},
		{"Set to 0.0", 0.0, 0.0},
		{"Set to 1.0", 1.0, 1.0},
		{"Set to 0.95", 0.95, 0.95},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RequestOptions{}
			option := WithTopP(tt.topP)
			option(opts)

			if opts.TopP == nil {
				t.Errorf("WithTopP did not set TopP")
			} else if *opts.TopP != tt.expected {
				t.Errorf("TopP = %v, want %v", *opts.TopP, tt.expected)
			}
		})
	}
}

func TestWithStopSequences(t *testing.T) {
	tests := []struct {
		name      string
		sequences []string
	}{
		{"Single sequence", []string{"\n\n"}},
		{"Multiple sequences", []string{"END", "STOP", "\n\n"}},
		{"Empty sequence", []string{}},
		{"Nil sequence", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RequestOptions{}
			option := WithStopSequences(tt.sequences...)
			option(opts)

			if len(tt.sequences) == 0 {
				if len(opts.StopSequences) != 0 {
					t.Errorf("StopSequences = %v, want empty", opts.StopSequences)
				}
			} else {
				if len(opts.StopSequences) != len(tt.sequences) {
					t.Errorf("len(StopSequences) = %v, want %v", len(opts.StopSequences), len(tt.sequences))
				}
				for i, seq := range tt.sequences {
					if opts.StopSequences[i] != seq {
						t.Errorf("StopSequences[%d] = %v, want %v", i, opts.StopSequences[i], seq)
					}
				}
			}
		})
	}
}

func TestWithMetadata(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"String value", "session_id", "test-session"},
		{"Int value", "retry_count", 3},
		{"Bool value", "debug", true},
		{"Float value", "priority", 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RequestOptions{}
			option := WithMetadata(tt.key, tt.value)
			option(opts)

			if opts.Metadata == nil {
				t.Errorf("WithMetadata did not initialize Metadata map")
			} else if opts.Metadata[tt.key] != tt.value {
				t.Errorf("Metadata[%s] = %v, want %v", tt.key, opts.Metadata[tt.key], tt.value)
			}
		})
	}
}

func TestWithMetadata_MultipleKeys(t *testing.T) {
	opts := &RequestOptions{}

	option1 := WithMetadata("key1", "value1")
	option2 := WithMetadata("key2", 42)
	option3 := WithMetadata("key3", true)

	option1(opts)
	option2(opts)
	option3(opts)

	if len(opts.Metadata) != 3 {
		t.Errorf("len(Metadata) = %v, want 3", len(opts.Metadata))
	}

	if opts.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want value1", opts.Metadata["key1"])
	}
	if opts.Metadata["key2"] != 42 {
		t.Errorf("Metadata[key2] = %v, want 42", opts.Metadata["key2"])
	}
	if opts.Metadata["key3"] != true {
		t.Errorf("Metadata[key3] = %v, want true", opts.Metadata["key3"])
	}
}

func TestApplyOptions_SingleOption(t *testing.T) {
	req := &ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	ApplyOptions(req, WithMaxTokens(2048))

	if req.MaxTokens != 2048 {
		t.Errorf("MaxTokens = %v, want 2048", req.MaxTokens)
	}
}

func TestApplyOptions_MultipleOptions(t *testing.T) {
	req := &ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	ApplyOptions(req,
		WithMaxTokens(1024),
		WithTemperature(0.7),
		WithTopP(0.9),
		WithStopSequences("END", "STOP"),
		WithMetadata("session_id", "test-session"),
	)

	if req.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %v, want 1024", req.MaxTokens)
	}
	if req.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want 0.7", req.Temperature)
	}
	if req.TopP != 0.9 {
		t.Errorf("TopP = %v, want 0.9", req.TopP)
	}
	if len(req.StopSequences) != 2 || req.StopSequences[0] != "END" || req.StopSequences[1] != "STOP" {
		t.Errorf("StopSequences = %v, want [END STOP]", req.StopSequences)
	}
	if req.Metadata["session_id"] != "test-session" {
		t.Errorf("Metadata[session_id] = %v, want test-session", req.Metadata["session_id"])
	}
}

func TestApplyOptions_NoOptions(t *testing.T) {
	req := &ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "Test"}},
		Model:       "test-model",
		MaxTokens:   500,
		Temperature: 0.5,
	}

	originalMaxTokens := req.MaxTokens
	originalTemperature := req.Temperature

	ApplyOptions(req)

	if req.MaxTokens != originalMaxTokens {
		t.Errorf("MaxTokens changed from %v to %v", originalMaxTokens, req.MaxTokens)
	}
	if req.Temperature != originalTemperature {
		t.Errorf("Temperature changed from %v to %v", originalTemperature, req.Temperature)
	}
}

func TestApplyOptions_MetadataMerge(t *testing.T) {
	req := &ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
		Metadata: map[string]interface{}{
			"existing_key": "existing_value",
		},
	}

	ApplyOptions(req,
		WithMetadata("new_key", "new_value"),
		WithMetadata("another_key", 123),
	)

	if len(req.Metadata) != 3 {
		t.Errorf("len(Metadata) = %v, want 3", len(req.Metadata))
	}
	if req.Metadata["existing_key"] != "existing_value" {
		t.Errorf("Metadata[existing_key] = %v, want existing_value", req.Metadata["existing_key"])
	}
	if req.Metadata["new_key"] != "new_value" {
		t.Errorf("Metadata[new_key] = %v, want new_value", req.Metadata["new_key"])
	}
	if req.Metadata["another_key"] != 123 {
		t.Errorf("Metadata[another_key] = %v, want 123", req.Metadata["another_key"])
	}
}

func TestApplyStreamOptions_SingleOption(t *testing.T) {
	req := &StreamRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	ApplyStreamOptions(req, WithMaxTokens(2048))

	if req.MaxTokens != 2048 {
		t.Errorf("MaxTokens = %v, want 2048", req.MaxTokens)
	}
}

func TestApplyStreamOptions_MultipleOptions(t *testing.T) {
	req := &StreamRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	ApplyStreamOptions(req,
		WithMaxTokens(1024),
		WithTemperature(0.8),
		WithTopP(0.95),
		WithStopSequences("END"),
		WithMetadata("stream_id", "stream-123"),
	)

	if req.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %v, want 1024", req.MaxTokens)
	}
	if req.Temperature != 0.8 {
		t.Errorf("Temperature = %v, want 0.8", req.Temperature)
	}
	if req.TopP != 0.95 {
		t.Errorf("TopP = %v, want 0.95", req.TopP)
	}
	if len(req.StopSequences) != 1 || req.StopSequences[0] != "END" {
		t.Errorf("StopSequences = %v, want [END]", req.StopSequences)
	}
	if req.Metadata["stream_id"] != "stream-123" {
		t.Errorf("Metadata[stream_id] = %v, want stream-123", req.Metadata["stream_id"])
	}
}

func TestApplyStreamOptions_NoOptions(t *testing.T) {
	req := &StreamRequest{
		Messages:    []Message{{Role: RoleUser, Content: "Test"}},
		Model:       "test-model",
		MaxTokens:   500,
		Temperature: 0.5,
	}

	originalMaxTokens := req.MaxTokens
	originalTemperature := req.Temperature

	ApplyStreamOptions(req)

	if req.MaxTokens != originalMaxTokens {
		t.Errorf("MaxTokens changed from %v to %v", originalMaxTokens, req.MaxTokens)
	}
	if req.Temperature != originalTemperature {
		t.Errorf("Temperature changed from %v to %v", originalTemperature, req.Temperature)
	}
}

func TestApplyStreamOptions_MetadataMerge(t *testing.T) {
	req := &StreamRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
		Metadata: map[string]interface{}{
			"existing_key": "existing_value",
		},
	}

	ApplyStreamOptions(req,
		WithMetadata("new_key", "new_value"),
		WithMetadata("another_key", 456),
	)

	if len(req.Metadata) != 3 {
		t.Errorf("len(Metadata) = %v, want 3", len(req.Metadata))
	}
	if req.Metadata["existing_key"] != "existing_value" {
		t.Errorf("Metadata[existing_key] = %v, want existing_value", req.Metadata["existing_key"])
	}
	if req.Metadata["new_key"] != "new_value" {
		t.Errorf("Metadata[new_key] = %v, want new_value", req.Metadata["new_key"])
	}
	if req.Metadata["another_key"] != 456 {
		t.Errorf("Metadata[another_key] = %v, want 456", req.Metadata["another_key"])
	}
}

func TestOptions_ChainedApplication(t *testing.T) {
	// Test that options can be created once and reused
	maxTokensOpt := WithMaxTokens(2048)
	temperatureOpt := WithTemperature(0.7)

	chatReq := &ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	streamReq := &StreamRequest{
		Messages: []Message{{Role: RoleUser, Content: "Test"}},
		Model:    "test-model",
	}

	ApplyOptions(chatReq, maxTokensOpt, temperatureOpt)
	ApplyStreamOptions(streamReq, maxTokensOpt, temperatureOpt)

	if chatReq.MaxTokens != 2048 || chatReq.Temperature != 0.7 {
		t.Errorf("ChatRequest options not applied correctly")
	}
	if streamReq.MaxTokens != 2048 || streamReq.Temperature != 0.7 {
		t.Errorf("StreamRequest options not applied correctly")
	}
}
