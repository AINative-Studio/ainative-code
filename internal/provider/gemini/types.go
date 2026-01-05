package gemini

// geminiRequest represents a request to the Gemini API
type geminiRequest struct {
	Contents         []geminiContent       `json:"contents"`
	SystemInstruction *geminiContent       `json:"systemInstruction,omitempty"`
	GenerationConfig *generationConfig    `json:"generationConfig,omitempty"`
	SafetySettings   []safetySetting      `json:"safetySettings,omitempty"`
	Tools            []geminiTool         `json:"tools,omitempty"`
}

// geminiContent represents content in the Gemini API format
type geminiContent struct {
	Role  string        `json:"role,omitempty"` // "user" or "model"
	Parts []geminiPart  `json:"parts"`
}

// geminiPart represents a part of content (text, image, etc.)
type geminiPart struct {
	Text         string        `json:"text,omitempty"`
	InlineData   *inlineData   `json:"inlineData,omitempty"`
	FileData     *fileData     `json:"fileData,omitempty"`
	FunctionCall *functionCall `json:"functionCall,omitempty"`
	FunctionResponse *functionResponse `json:"functionResponse,omitempty"`
}

// inlineData represents inline binary data (e.g., base64 encoded image)
type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

// fileData represents a file reference
type fileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"`
}

// generationConfig contains generation parameters
type generationConfig struct {
	StopSequences   []string `json:"stopSequences,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
}

// safetySetting represents a safety configuration
type safetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// geminiTool represents a function tool
type geminiTool struct {
	FunctionDeclarations []functionDeclaration `json:"functionDeclarations,omitempty"`
}

// functionDeclaration represents a function that can be called
type functionDeclaration struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON schema
}

// functionCall represents a function call from the model
type functionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// functionResponse represents a response to a function call
type functionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// geminiResponse represents a response from the Gemini API
type geminiResponse struct {
	Candidates     []candidate    `json:"candidates"`
	PromptFeedback *promptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *usageMetadata  `json:"usageMetadata,omitempty"`
}

// candidate represents a generated candidate
type candidate struct {
	Content       geminiContent   `json:"content"`
	FinishReason  string          `json:"finishReason,omitempty"`
	Index         int             `json:"index"`
	SafetyRatings []safetyRating  `json:"safetyRatings,omitempty"`
	CitationMetadata *citationMetadata `json:"citationMetadata,omitempty"`
}

// safetyRating represents a safety evaluation
type safetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
	Blocked     bool   `json:"blocked,omitempty"`
}

// promptFeedback contains feedback about the prompt
type promptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`
	SafetyRatings []safetyRating `json:"safetyRatings,omitempty"`
}

// citationMetadata contains citation information
type citationMetadata struct {
	CitationSources []citationSource `json:"citationSources,omitempty"`
}

// citationSource represents a citation source
type citationSource struct {
	StartIndex int    `json:"startIndex,omitempty"`
	EndIndex   int    `json:"endIndex,omitempty"`
	URI        string `json:"uri,omitempty"`
	License    string `json:"license,omitempty"`
}

// usageMetadata contains token usage information
type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// geminiError represents an error response from the Gemini API
type geminiError struct {
	Error errorDetails `json:"error"`
}

// errorDetails contains error details
type errorDetails struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Status  string      `json:"status"`
	Details []errorDetail `json:"details,omitempty"`
}

// errorDetail contains additional error information
type errorDetail struct {
	Type     string `json:"@type"`
	Reason   string `json:"reason,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// streamResponse represents a streaming response chunk
type streamResponse struct {
	Candidates     []candidate    `json:"candidates,omitempty"`
	PromptFeedback *promptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *usageMetadata  `json:"usageMetadata,omitempty"`
}
