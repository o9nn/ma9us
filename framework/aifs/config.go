package aifs

// Config holds the AI filesystem server configuration.
type Config struct {
	ListenAddr      string                    `json:"listen_addr"`
	DefaultModel    string                    `json:"default_model"`
	Models          []ModelConfig             `json:"models"`
	MaxSessions     int                       `json:"max_sessions"`
	EnableKnowledge bool                      `json:"enable_knowledge"`
	EnableTopology  bool                      `json:"enable_topology"`
	ProviderConfigs map[string]ProviderConfig `json:"providers"`
}

// ProviderConfig holds configuration for a single LLM provider.
type ProviderConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	OrgID   string `json:"org_id"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:      ":5641",
		DefaultModel:    "claude-sonnet-4-6",
		Models:          defaultModels(),
		MaxSessions:     0,
		EnableKnowledge: true,
		EnableTopology:  true,
		ProviderConfigs: make(map[string]ProviderConfig),
	}
}

func defaultModels() []ModelConfig {
	return []ModelConfig{
		{Name: "claude-opus-4-6", Provider: "anthropic", MaxTokens: 200000},
		{Name: "claude-sonnet-4-6", Provider: "anthropic", MaxTokens: 200000},
		{Name: "claude-haiku-4-5", Provider: "anthropic", MaxTokens: 200000},
		{Name: "gpt-4o", Provider: "openai", MaxTokens: 128000},
		{Name: "gpt-4o-mini", Provider: "openai", MaxTokens: 128000},
		{Name: "gemini-2.5-pro", Provider: "google", MaxTokens: 1000000},
		{Name: "gemini-2.5-flash", Provider: "google", MaxTokens: 1000000},
	}
}
