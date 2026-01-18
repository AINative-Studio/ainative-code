package provider

// ProviderCapabilities defines the capabilities of each supported LLM provider
var ProviderCapabilities = map[string]ProviderInfo{
	"anthropic": {
		Name:                    "anthropic",
		DisplayName:             "Anthropic Claude",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		MaxTokens:               200000,
	},
	"openai": {
		Name:                    "openai",
		DisplayName:             "OpenAI GPT",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		MaxTokens:               128000,
	},
	"google": {
		Name:                    "google",
		DisplayName:             "Google Gemini",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		MaxTokens:               1000000,
	},
}
