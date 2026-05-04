package aifs

import (
	"fmt"
	"strings"
	"sync"
)

// ModelConfig describes an available LLM model.
type ModelConfig struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	MaxTokens int    `json:"max_tokens"`
}

// ModelRegistry tracks available models.
type ModelRegistry struct {
	mu     sync.RWMutex
	models map[string]ModelConfig
	active string
}

func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{models: make(map[string]ModelConfig)}
}

func (r *ModelRegistry) Register(m ModelConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models[m.Name] = m
	if r.active == "" {
		r.active = m.Name
	}
}

func (r *ModelRegistry) Get(name string) (ModelConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.models[name]
	return m, ok
}

func (r *ModelRegistry) Active() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

func (r *ModelRegistry) SetActive(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.models[name]; !ok {
		return fmt.Errorf("model %q not found", name)
	}
	r.active = name
	return nil
}

func (r *ModelRegistry) List() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var b strings.Builder
	for name, m := range r.models {
		marker := " "
		if name == r.active {
			marker = "*"
		}
		fmt.Fprintf(&b, "%s %-30s %-12s %d\n", marker, m.Name, m.Provider, m.MaxTokens)
	}
	return b.String()
}
