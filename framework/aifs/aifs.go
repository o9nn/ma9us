// Package aifs implements the AI Filesystem — a 9P server that exposes
// AI services, knowledge graphs, and organizational topology as files.
//
// This is the core of the 9cog framework, bridging:
//   - go9p (9P protocol)
//   - aichat (LLM providers)
//   - cogpwsh (knowledge graph)
//   - 120c (organizational topology)
//   - airc (rc shell integration)
package aifs

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Server is the AI filesystem server that synthesizes all framework components.
type Server struct {
	mu       sync.RWMutex
	sessions map[uint64]*Session
	models   *ModelRegistry
	graph    *KnowledgeGraph
	topology *Topology
	config   *Config
	nextID   atomic.Uint64
}

// NewServer creates a new AI filesystem server with the given configuration.
func NewServer(cfg *Config) *Server {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	s := &Server{
		sessions: make(map[uint64]*Session),
		models:   NewModelRegistry(),
		graph:    NewKnowledgeGraph(),
		topology: NewTopology(),
		config:   cfg,
	}
	for _, m := range cfg.Models {
		s.models.Register(m)
	}
	return s
}

// NewSession creates a new AI conversation session.
func (s *Server) NewSession(user string) *Session {
	id := s.nextID.Add(1)
	sess := &Session{
		ID:        id,
		User:      user,
		CreatedAt: time.Now(),
		Model:     s.config.DefaultModel,
		Messages:  make([]Message, 0),
		Context:   make(map[string]string),
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	return sess
}

// GetSession retrieves a session by ID.
func (s *Server) GetSession(id uint64) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

// ListSessions returns all active sessions.
func (s *Server) ListSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		result = append(result, sess)
	}
	return result
}

// Session represents an AI conversation with state.
type Session struct {
	mu        sync.RWMutex
	ID        uint64
	User      string
	Model     string
	CreatedAt time.Time
	Messages  []Message
	Context   map[string]string
}

// Message is a single message in a conversation.
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// AddMessage appends a message to the session history.
func (s *Session) AddMessage(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// History returns the conversation history as a formatted string.
func (s *Session) History() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b strings.Builder
	for _, m := range s.Messages {
		fmt.Fprintf(&b, "[%s] %s: %s\n", m.Timestamp.Format(time.RFC3339), m.Role, m.Content)
	}
	return b.String()
}

// SetContext updates session context (pwd, shell, etc.)
func (s *Session) SetContext(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Context[key] = value
}

// Meta returns session metadata as a formatted string.
func (s *Session) Meta() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("id: %d\nuser: %s\nmodel: %s\ncreated: %s\nmessages: %d\n",
		s.ID, s.User, s.Model, s.CreatedAt.Format(time.RFC3339), len(s.Messages))
}
