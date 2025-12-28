package provider

// ChatOptions contains configuration options for chat requests
type ChatOptions struct {
	Model          string
	MaxTokens      int
	Temperature    float64
	TopP           float64
	StopSequences  []string
	SystemPrompt   string
	Stream         bool
	Metadata       map[string]string
}

// ChatOption is a function that modifies ChatOptions
type ChatOption func(*ChatOptions)

// StreamOption is a function that modifies ChatOptions for streaming requests
type StreamOption func(*ChatOptions)

// WithModel sets the model to use for the request
func WithModel(model string) ChatOption {
	return func(opts *ChatOptions) {
		opts.Model = model
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(maxTokens int) ChatOption {
	return func(opts *ChatOptions) {
		opts.MaxTokens = maxTokens
	}
}

// WithTemperature sets the sampling temperature (0.0 to 1.0)
func WithTemperature(temperature float64) ChatOption {
	return func(opts *ChatOptions) {
		opts.Temperature = temperature
	}
}

// WithTopP sets the nucleus sampling parameter (0.0 to 1.0)
func WithTopP(topP float64) ChatOption {
	return func(opts *ChatOptions) {
		opts.TopP = topP
	}
}

// WithStopSequences sets the sequences that will stop generation
func WithStopSequences(sequences ...string) ChatOption {
	return func(opts *ChatOptions) {
		opts.StopSequences = sequences
	}
}

// WithSystemPrompt sets the system prompt for the conversation
func WithSystemPrompt(prompt string) ChatOption {
	return func(opts *ChatOptions) {
		opts.SystemPrompt = prompt
	}
}

// WithMetadata adds custom metadata to the request
func WithMetadata(key, value string) ChatOption {
	return func(opts *ChatOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
	}
}

// StreamWithModel sets the model to use for streaming requests
func StreamWithModel(model string) StreamOption {
	return func(opts *ChatOptions) {
		opts.Model = model
		opts.Stream = true
	}
}

// StreamWithMaxTokens sets the maximum number of tokens for streaming requests
func StreamWithMaxTokens(maxTokens int) StreamOption {
	return func(opts *ChatOptions) {
		opts.MaxTokens = maxTokens
		opts.Stream = true
	}
}

// StreamWithTemperature sets the temperature for streaming requests
func StreamWithTemperature(temperature float64) StreamOption {
	return func(opts *ChatOptions) {
		opts.Temperature = temperature
		opts.Stream = true
	}
}

// StreamWithTopP sets the top_p for streaming requests
func StreamWithTopP(topP float64) StreamOption {
	return func(opts *ChatOptions) {
		opts.TopP = topP
		opts.Stream = true
	}
}

// StreamWithStopSequences sets stop sequences for streaming requests
func StreamWithStopSequences(sequences ...string) StreamOption {
	return func(opts *ChatOptions) {
		opts.StopSequences = sequences
		opts.Stream = true
	}
}

// StreamWithSystemPrompt sets the system prompt for streaming requests
func StreamWithSystemPrompt(prompt string) StreamOption {
	return func(opts *ChatOptions) {
		opts.SystemPrompt = prompt
		opts.Stream = true
	}
}

// StreamWithMetadata adds metadata for streaming requests
func StreamWithMetadata(key, value string) StreamOption {
	return func(opts *ChatOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
		opts.Stream = true
	}
}

// ApplyChatOptions applies a list of ChatOption functions to ChatOptions
func ApplyChatOptions(opts *ChatOptions, options ...ChatOption) {
	for _, opt := range options {
		opt(opts)
	}
}

// ApplyStreamOptions applies a list of StreamOption functions to ChatOptions
func ApplyStreamOptions(opts *ChatOptions, options ...StreamOption) {
	opts.Stream = true
	for _, opt := range options {
		opt(opts)
	}
}

// DefaultChatOptions returns ChatOptions with sensible defaults
func DefaultChatOptions() *ChatOptions {
	return &ChatOptions{
		MaxTokens:   1024,
		Temperature: 0.7,
		TopP:        1.0,
		Stream:      false,
		Metadata:    make(map[string]string),
	}
}
