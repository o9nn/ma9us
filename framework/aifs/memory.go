package aifs

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryKind identifies one of the six memory subsystems in the
// regima-cognitive-ai six-memory architecture.
type MemoryKind int

const (
	// SemanticMemory stores long-term factual and conceptual knowledge.
	SemanticMemory MemoryKind = iota
	// EpisodicMemory stores autobiographical events and experiences.
	EpisodicMemory
	// ProceduralMemory stores skills, rules, and how-to knowledge.
	ProceduralMemory
	// WorkingMemory is the short-term active buffer for ongoing reasoning.
	WorkingMemory
	// PerceptualMemory stores sensory and perceptual inputs.
	PerceptualMemory
	// ParticipatoryMemory stores social, relational, and co-created knowledge.
	ParticipatoryMemory
)

var memoryKindNames = map[MemoryKind]string{
	SemanticMemory:      "semantic",
	EpisodicMemory:      "episodic",
	ProceduralMemory:    "procedural",
	WorkingMemory:       "working",
	PerceptualMemory:    "perceptual",
	ParticipatoryMemory: "participatory",
}

// String returns the canonical lower-case name of the memory kind.
func (k MemoryKind) String() string {
	if name, ok := memoryKindNames[k]; ok {
		return name
	}
	return fmt.Sprintf("MemoryKind(%d)", int(k))
}

// AllMemoryKinds lists all six subsystems in canonical order.
var AllMemoryKinds = []MemoryKind{
	SemanticMemory, EpisodicMemory, ProceduralMemory,
	WorkingMemory, PerceptualMemory, ParticipatoryMemory,
}

// SyncEvent is the canonical broadcast event wired into the MemoryConsolidator.
// It mirrors the deltecho sync_event shape: every sync_event is a write
// opportunity; slow consolidation must not block the next perception cycle.
type SyncEvent struct {
	ID        uint64
	Kind      string
	Salience  float64
	Payload   string
	Timestamp time.Time
}

// SubsystemSalience maps each memory kind to its write-gate threshold.
// A SyncEvent is written to a subsystem only when event.Salience >= threshold.
type SubsystemSalience map[MemoryKind]float64

// DefaultSubsystemSalience returns the initial salience thresholds.
func DefaultSubsystemSalience() SubsystemSalience {
	return SubsystemSalience{
		SemanticMemory:      0.3,
		EpisodicMemory:      0.2,
		ProceduralMemory:    0.5,
		WorkingMemory:       0.1,
		PerceptualMemory:    0.1,
		ParticipatoryMemory: 0.4,
	}
}

// MemorySubsystem is a typed AtomSpace partition for one of the six memory
// kinds. It indexes atoms by their MatulaName (eternal prime name).
type MemorySubsystem struct {
	mu         sync.RWMutex
	Kind       MemoryKind
	atomIndex  map[uint64]*Atom // MatulaName → Atom
	writeCount uint64
}

func newMemorySubsystem(kind MemoryKind) *MemorySubsystem {
	return &MemorySubsystem{
		Kind:      kind,
		atomIndex: make(map[uint64]*Atom),
	}
}

// Register records an atom in this subsystem's MatulaName index.
func (s *MemorySubsystem) Register(a *Atom) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.MatulaName != 0 {
		s.atomIndex[a.MatulaName] = a
	}
	s.writeCount++
}

// LookupByMatula returns the atom with the given MatulaName, if any.
func (s *MemorySubsystem) LookupByMatula(name uint64) (*Atom, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.atomIndex[name]
	return a, ok
}

// ListAtoms returns all registered atoms as a formatted string.
func (s *MemorySubsystem) ListAtoms() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b strings.Builder
	for matulaName, atom := range s.atomIndex {
		fmt.Fprintf(&b, "%d\t%s\t%s\n", matulaName, atom.Name, atom.TV)
	}
	return b.String()
}

// Stats returns a one-line summary of this subsystem.
func (s *MemorySubsystem) Stats() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("kind: %s\natoms: %d\nwrites: %d\n",
		s.Kind, len(s.atomIndex), s.writeCount)
}

// MemoryConsolidator wires the six memory subsystems to a broadcast channel.
// Every SyncEvent becomes a write opportunity, gated by SubsystemSalience.
// The drop-on-full channel ensures slow consolidation never blocks the
// perception cycle (mirroring the tickInProgress guard in the agent loop).
type MemoryConsolidator struct {
	mu          sync.RWMutex
	graph       *KnowledgeGraph
	subsystems  map[MemoryKind]*MemorySubsystem
	saliences   SubsystemSalience
	eventCount  uint64
	broadcastCh chan SyncEvent
	stopCh      chan struct{}
}

// NewMemoryConsolidator creates a consolidator backed by the given graph and
// starts its internal consolidation goroutine.
func NewMemoryConsolidator(graph *KnowledgeGraph) *MemoryConsolidator {
	subs := make(map[MemoryKind]*MemorySubsystem, len(AllMemoryKinds))
	for _, kind := range AllMemoryKinds {
		subs[kind] = newMemorySubsystem(kind)
	}
	mc := &MemoryConsolidator{
		graph:       graph,
		subsystems:  subs,
		saliences:   DefaultSubsystemSalience(),
		broadcastCh: make(chan SyncEvent, 64),
		stopCh:      make(chan struct{}),
	}
	go mc.consolidationLoop()
	return mc
}

// Broadcast enqueues a SyncEvent for consolidation. If the channel is full,
// the event is dropped rather than blocking — slow consolidation must not
// block the next perception cycle.
func (mc *MemoryConsolidator) Broadcast(e SyncEvent) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	select {
	case mc.broadcastCh <- e:
	default:
		// drop: overrun; consolidation is behind
	}
}

// consolidationLoop is the cooperative event consumer goroutine.
func (mc *MemoryConsolidator) consolidationLoop() {
	for {
		select {
		case e := <-mc.broadcastCh:
			mc.consolidate(e)
		case <-mc.stopCh:
			return
		}
	}
}

// consolidate writes event e to every subsystem whose salience threshold is
// met by e.Salience. The event is materialised as a ConceptNode in the shared
// graph so it becomes part of the persistent AtomSpace.
func (mc *MemoryConsolidator) consolidate(e SyncEvent) {
	mc.mu.Lock()
	mc.eventCount++
	id := mc.eventCount
	sals := mc.saliences
	mc.mu.Unlock()

	node := mc.graph.AddNode(
		ConceptNode,
		fmt.Sprintf("sync:%s:%d", e.Kind, id),
		TruthValue{Strength: e.Salience, Confidence: 0.5},
	)
	for _, kind := range AllMemoryKinds {
		if e.Salience >= sals[kind] {
			mc.subsystems[kind].Register(node)
		}
	}
}

// Stop shuts down the consolidation goroutine. It is safe to call once.
func (mc *MemoryConsolidator) Stop() {
	close(mc.stopCh)
}

// Subsystem returns the named memory subsystem.
func (mc *MemoryConsolidator) Subsystem(kind MemoryKind) *MemorySubsystem {
	return mc.subsystems[kind]
}

// SetSalience updates the write-gate threshold for a subsystem.
func (mc *MemoryConsolidator) SetSalience(kind MemoryKind, threshold float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.saliences[kind] = threshold
}

// Stats returns an overview of the consolidator and all subsystems.
func (mc *MemoryConsolidator) Stats() string {
	mc.mu.RLock()
	ec := mc.eventCount
	mc.mu.RUnlock()
	var b strings.Builder
	fmt.Fprintf(&b, "events_consolidated: %d\n", ec)
	for _, kind := range AllMemoryKinds {
		fmt.Fprintf(&b, "\n[%s]\n%s", kind, mc.subsystems[kind].Stats())
	}
	return b.String()
}
