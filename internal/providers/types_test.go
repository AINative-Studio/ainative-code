package providers

import (
	"testing"
	"time"
)

func TestRole_Constants(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected string
	}{
		{"User role", RoleUser, "user"},
		{"Assistant role", RoleAssistant, "assistant"},
		{"System role", RoleSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("Role = %v, want %v", tt.role, tt.expected)
			}
		})
	}
}

func TestEventType_Constants(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  string
	}{
		{"TextDelta", EventTextDelta, "text_delta"},
		{"ContentStart", EventContentStart, "content_start"},
		{"ContentEnd", EventContentEnd, "content_end"},
		{"MessageStart", EventMessageStart, "message_start"},
		{"MessageStop", EventMessageStop, "message_stop"},
		{"Error", EventError, "error"},
		{"Usage", EventUsage, "usage"},
		{"Thinking", EventThinking, "thinking"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("EventType = %v, want %v", tt.eventType, tt.expected)
			}
		})
	}
}

func TestMessage_Creation(t *testing.T) {
	msg := Message{
		Role:    RoleUser,
		Content: "Hello, world!",
	}

	if msg.Role != RoleUser {
		t.Errorf("Message.Role = %v, want %v", msg.Role, RoleUser)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("Message.Content = %v, want %v", msg.Content, "Hello, world!")
	}
}

func TestResponse_Creation(t *testing.T) {
	now := time.Now()
	usage := &UsageInfo{
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
	}
	metadata := map[string]interface{}{
		"key": "value",
	}

	resp := Response{
		Content:      "Test response",
		Model:        "test-model",
		Provider:     "test-provider",
		FinishReason: "stop",
		Usage:        usage,
		Metadata:     metadata,
		CreatedAt:    now,
	}

	if resp.Content != "Test response" {
		t.Errorf("Response.Content = %v, want %v", resp.Content, "Test response")
	}
	if resp.Model != "test-model" {
		t.Errorf("Response.Model = %v, want %v", resp.Model, "test-model")
	}
	if resp.Provider != "test-provider" {
		t.Errorf("Response.Provider = %v, want %v", resp.Provider, "test-provider")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("Response.FinishReason = %v, want %v", resp.FinishReason, "stop")
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 30 {
		t.Errorf("Response.Usage.TotalTokens = %v, want %v", resp.Usage.TotalTokens, 30)
	}
	if resp.Metadata["key"] != "value" {
		t.Errorf("Response.Metadata[key] = %v, want %v", resp.Metadata["key"], "value")
	}
	if !resp.CreatedAt.Equal(now) {
		t.Errorf("Response.CreatedAt = %v, want %v", resp.CreatedAt, now)
	}
}

func TestEvent_Creation(t *testing.T) {
	now := time.Now()
	usage := &UsageInfo{
		PromptTokens:     5,
		CompletionTokens: 10,
		TotalTokens:      15,
	}

	event := Event{
		Type:      EventTextDelta,
		Data:      "test data",
		Usage:     usage,
		Timestamp: now,
	}

	if event.Type != EventTextDelta {
		t.Errorf("Event.Type = %v, want %v", event.Type, EventTextDelta)
	}
	if event.Data != "test data" {
		t.Errorf("Event.Data = %v, want %v", event.Data, "test data")
	}
	if event.Usage == nil || event.Usage.TotalTokens != 15 {
		t.Errorf("Event.Usage.TotalTokens = %v, want %v", event.Usage.TotalTokens, 15)
	}
	if !event.Timestamp.Equal(now) {
		t.Errorf("Event.Timestamp = %v, want %v", event.Timestamp, now)
	}
}

func TestUsageInfo_Creation(t *testing.T) {
	usage := UsageInfo{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("UsageInfo.PromptTokens = %v, want %v", usage.PromptTokens, 100)
	}
	if usage.CompletionTokens != 200 {
		t.Errorf("UsageInfo.CompletionTokens = %v, want %v", usage.CompletionTokens, 200)
	}
	if usage.TotalTokens != 300 {
		t.Errorf("UsageInfo.TotalTokens = %v, want %v", usage.TotalTokens, 300)
	}
}

func TestModel_Creation(t *testing.T) {
	capabilities := []string{"chat", "streaming"}
	model := Model{
		ID:           "test-model-1",
		Name:         "Test Model",
		Provider:     "test-provider",
		MaxTokens:    4096,
		Capabilities: capabilities,
	}

	if model.ID != "test-model-1" {
		t.Errorf("Model.ID = %v, want %v", model.ID, "test-model-1")
	}
	if model.Name != "Test Model" {
		t.Errorf("Model.Name = %v, want %v", model.Name, "Test Model")
	}
	if model.Provider != "test-provider" {
		t.Errorf("Model.Provider = %v, want %v", model.Provider, "test-provider")
	}
	if model.MaxTokens != 4096 {
		t.Errorf("Model.MaxTokens = %v, want %v", model.MaxTokens, 4096)
	}
	if len(model.Capabilities) != 2 {
		t.Errorf("len(Model.Capabilities) = %v, want %v", len(model.Capabilities), 2)
	}
}

func TestChatRequest_Creation(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
		{Role: RoleAssistant, Content: "Hi there!"},
	}
	metadata := map[string]interface{}{
		"session_id": "test-session",
	}

	req := ChatRequest{
		Messages:      messages,
		Model:         "test-model",
		MaxTokens:     1024,
		Temperature:   0.7,
		TopP:          0.9,
		StopSequences: []string{"\n\n"},
		Metadata:      metadata,
	}

	if len(req.Messages) != 2 {
		t.Errorf("len(ChatRequest.Messages) = %v, want %v", len(req.Messages), 2)
	}
	if req.Model != "test-model" {
		t.Errorf("ChatRequest.Model = %v, want %v", req.Model, "test-model")
	}
	if req.MaxTokens != 1024 {
		t.Errorf("ChatRequest.MaxTokens = %v, want %v", req.MaxTokens, 1024)
	}
	if req.Temperature != 0.7 {
		t.Errorf("ChatRequest.Temperature = %v, want %v", req.Temperature, 0.7)
	}
	if req.TopP != 0.9 {
		t.Errorf("ChatRequest.TopP = %v, want %v", req.TopP, 0.9)
	}
	if len(req.StopSequences) != 1 || req.StopSequences[0] != "\n\n" {
		t.Errorf("ChatRequest.StopSequences = %v, want %v", req.StopSequences, []string{"\n\n"})
	}
	if req.Metadata["session_id"] != "test-session" {
		t.Errorf("ChatRequest.Metadata[session_id] = %v, want %v", req.Metadata["session_id"], "test-session")
	}
}

func TestStreamRequest_Creation(t *testing.T) {
	messages := []Message{
		{Role: RoleSystem, Content: "You are helpful."},
		{Role: RoleUser, Content: "Tell me a story."},
	}

	req := StreamRequest{
		Messages:      messages,
		Model:         "stream-model",
		MaxTokens:     2048,
		Temperature:   0.8,
		TopP:          0.95,
		StopSequences: []string{"END"},
	}

	if len(req.Messages) != 2 {
		t.Errorf("len(StreamRequest.Messages) = %v, want %v", len(req.Messages), 2)
	}
	if req.Messages[0].Role != RoleSystem {
		t.Errorf("StreamRequest.Messages[0].Role = %v, want %v", req.Messages[0].Role, RoleSystem)
	}
	if req.Model != "stream-model" {
		t.Errorf("StreamRequest.Model = %v, want %v", req.Model, "stream-model")
	}
	if req.MaxTokens != 2048 {
		t.Errorf("StreamRequest.MaxTokens = %v, want %v", req.MaxTokens, 2048)
	}
}
