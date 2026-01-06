package azure

// azureRequest represents a request to the Azure OpenAI Chat Completions API
// Azure OpenAI uses the same request format as OpenAI
type azureRequest struct {
	Model            string          `json:"model,omitempty"` // Optional in Azure, deployment is in URL path
	Messages         []azureMessage  `json:"messages"`
	MaxTokens        int             `json:"max_tokens,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"top_p,omitempty"`
	N                int             `json:"n,omitempty"`
	Stream           bool            `json:"stream"`
	Stop             []string        `json:"stop,omitempty"`
	PresencePenalty  *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64        `json:"frequency_penalty,omitempty"`
	User             string          `json:"user,omitempty"`
	Tools            []azureTool     `json:"tools,omitempty"`
	ToolChoice       interface{}     `json:"tool_choice,omitempty"`
	ResponseFormat   *responseFormat `json:"response_format,omitempty"`
}

// azureMessage represents a message in the Azure OpenAI API format
type azureMessage struct {
	Role       string        `json:"role"`    // "system", "user", "assistant", "tool"
	Content    interface{}   `json:"content"` // string or array of content parts
	Name       string        `json:"name,omitempty"`
	ToolCalls  []toolCall    `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

// contentPart represents a part of multi-modal content
type contentPart struct {
	Type     string    `json:"type"` // "text" or "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

// imageURL represents an image URL in content
type imageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// azureTool represents a function tool definition
type azureTool struct {
	Type     string      `json:"type"` // "function"
	Function functionDef `json:"function"`
}

// functionDef represents a function definition
type functionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON schema
}

// toolCall represents a tool/function call from the assistant
type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function functionCall `json:"function"`
}

// functionCall represents a function call
type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// responseFormat controls the output format
type responseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// azureResponse represents a response from the Azure OpenAI Chat Completions API
type azureResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

// choice represents a completion choice
type choice struct {
	Index        int          `json:"index"`
	Message      azureMessage `json:"message"`
	FinishReason string       `json:"finish_reason"` // "stop", "length", "function_call", "content_filter", "tool_calls"
}

// usage represents token usage information
type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// azureError represents an error response from the Azure OpenAI API
type azureError struct {
	Error errorDetails `json:"error"`
}

// errorDetails contains error details
type errorDetails struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param,omitempty"`
	Code    interface{} `json:"code,omitempty"`
}

// streamEvent represents a Server-Sent Event from streaming
type streamEvent struct {
	eventType string
	data      string
}

// azureStreamResponse represents a streaming response chunk
type azureStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []streamChoice `json:"choices"`
}

// streamChoice represents a streaming completion choice
type streamChoice struct {
	Index        int          `json:"index"`
	Delta        messageDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"` // nil during streaming, set on completion
}

// messageDelta represents incremental message content
type messageDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}
