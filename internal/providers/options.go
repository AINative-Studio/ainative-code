package providers

// Option is a functional option for configuring requests
type Option func(*RequestOptions)

// RequestOptions holds optional parameters for Chat and Stream requests
type RequestOptions struct {
	MaxTokens     *int
	Temperature   *float64
	TopP          *float64
	StopSequences []string
	Metadata      map[string]interface{}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(tokens int) Option {
	return func(opts *RequestOptions) {
		opts.MaxTokens = &tokens
	}
}

// WithTemperature sets the sampling temperature (typically 0.0 to 1.0)
func WithTemperature(temp float64) Option {
	return func(opts *RequestOptions) {
		opts.Temperature = &temp
	}
}

// WithTopP sets the nucleus sampling parameter (typically 0.0 to 1.0)
func WithTopP(topP float64) Option {
	return func(opts *RequestOptions) {
		opts.TopP = &topP
	}
}

// WithStopSequences sets sequences where the model will stop generating
func WithStopSequences(sequences ...string) Option {
	return func(opts *RequestOptions) {
		opts.StopSequences = sequences
	}
}

// WithMetadata adds custom metadata to the request
func WithMetadata(key string, value interface{}) Option {
	return func(opts *RequestOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata[key] = value
	}
}

// ApplyOptions applies all functional options to a ChatRequest
func ApplyOptions(req *ChatRequest, opts ...Option) {
	options := &RequestOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}
	if options.Temperature != nil {
		req.Temperature = *options.Temperature
	}
	if options.TopP != nil {
		req.TopP = *options.TopP
	}
	if len(options.StopSequences) > 0 {
		req.StopSequences = options.StopSequences
	}
	if options.Metadata != nil {
		if req.Metadata == nil {
			req.Metadata = make(map[string]interface{})
		}
		for k, v := range options.Metadata {
			req.Metadata[k] = v
		}
	}
}

// ApplyStreamOptions applies all functional options to a StreamRequest
func ApplyStreamOptions(req *StreamRequest, opts ...Option) {
	options := &RequestOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}
	if options.Temperature != nil {
		req.Temperature = *options.Temperature
	}
	if options.TopP != nil {
		req.TopP = *options.TopP
	}
	if len(options.StopSequences) > 0 {
		req.StopSequences = options.StopSequences
	}
	if options.Metadata != nil {
		if req.Metadata == nil {
			req.Metadata = make(map[string]interface{})
		}
		for k, v := range options.Metadata {
			req.Metadata[k] = v
		}
	}
}
