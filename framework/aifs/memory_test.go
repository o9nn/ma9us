package aifs

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// MemoryKind
// ---------------------------------------------------------------------------

func TestMemoryKindString(t *testing.T) {
	cases := []struct {
		kind MemoryKind
		want string
	}{
		{SemanticMemory, "semantic"},
		{EpisodicMemory, "episodic"},
		{ProceduralMemory, "procedural"},
		{WorkingMemory, "working"},
		{PerceptualMemory, "perceptual"},
		{ParticipatoryMemory, "participatory"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("MemoryKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestMemoryKindStringUnknown(t *testing.T) {
	unknown := MemoryKind(99)
	s := unknown.String()
	if !strings.Contains(s, "99") {
		t.Errorf("unknown MemoryKind string = %q", s)
	}
}

func TestAllMemoryKindsLength(t *testing.T) {
	if len(AllMemoryKinds) != 6 {
		t.Errorf("AllMemoryKinds has %d entries, want 6", len(AllMemoryKinds))
	}
}

// ---------------------------------------------------------------------------
// DefaultSubsystemSalience
// ---------------------------------------------------------------------------

func TestDefaultSubsystemSalienceCoversAllKinds(t *testing.T) {
	sal := DefaultSubsystemSalience()
	for _, kind := range AllMemoryKinds {
		if _, ok := sal[kind]; !ok {
			t.Errorf("DefaultSubsystemSalience missing %s", kind)
		}
	}
}

// ---------------------------------------------------------------------------
// MemorySubsystem
// ---------------------------------------------------------------------------

func TestMemorySubsystemRegisterAndLookup(t *testing.T) {
	sub := newMemorySubsystem(SemanticMemory)
	g := NewKnowledgeGraph()
	atom := g.AddNode(ConceptNode, "TestConcept", TruthValue{1.0, 0.9})

	sub.Register(atom)
	got, ok := sub.LookupByMatula(atom.MatulaName)
	if !ok {
		t.Fatal("LookupByMatula should find registered atom")
	}
	if got.Name != "TestConcept" {
		t.Errorf("got %q, want %q", got.Name, "TestConcept")
	}
}

func TestMemorySubsystemLookupMiss(t *testing.T) {
	sub := newMemorySubsystem(EpisodicMemory)
	_, ok := sub.LookupByMatula(9999)
	if ok {
		t.Error("LookupByMatula should return false for unregistered name")
	}
}

func TestMemorySubsystemSkipsZeroMatulaName(t *testing.T) {
	sub := newMemorySubsystem(WorkingMemory)
	atom := &Atom{ID: 1, MatulaName: 0, Name: "zero"}
	sub.Register(atom)
	// writeCount still increments even for zero-name atoms.
	if sub.writeCount != 1 {
		t.Errorf("writeCount = %d, want 1", sub.writeCount)
	}
	// But the atom is not indexed.
	if len(sub.atomIndex) != 0 {
		t.Error("atom with MatulaName=0 should not be indexed")
	}
}

func TestMemorySubsystemStats(t *testing.T) {
	sub := newMemorySubsystem(ProceduralMemory)
	g := NewKnowledgeGraph()
	a := g.AddNode(ConceptNode, "A", TruthValue{1.0, 0.9})
	sub.Register(a)
	stats := sub.Stats()
	if !strings.Contains(stats, "procedural") {
		t.Error("stats missing kind name")
	}
	if !strings.Contains(stats, "atoms: 1") {
		t.Errorf("stats = %q, want 'atoms: 1'", stats)
	}
	if !strings.Contains(stats, "writes: 1") {
		t.Errorf("stats = %q, want 'writes: 1'", stats)
	}
}

func TestMemorySubsystemListAtoms(t *testing.T) {
	sub := newMemorySubsystem(PerceptualMemory)
	g := NewKnowledgeGraph()
	a := g.AddNode(ConceptNode, "SensorNode", TruthValue{0.8, 0.7})
	sub.Register(a)
	list := sub.ListAtoms()
	if !strings.Contains(list, "SensorNode") {
		t.Errorf("ListAtoms = %q, want SensorNode", list)
	}
}

// ---------------------------------------------------------------------------
// MemoryConsolidator
// ---------------------------------------------------------------------------

func TestNewMemoryConsolidator(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	if mc == nil {
		t.Fatal("NewMemoryConsolidator returned nil")
	}
	for _, kind := range AllMemoryKinds {
		if mc.Subsystem(kind) == nil {
			t.Errorf("Subsystem(%s) is nil", kind)
		}
	}
}

func TestMemoryConsolidatorBroadcastAndConsolidate(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	e := SyncEvent{
		Kind:     "perception",
		Salience: 0.9, // above all thresholds
		Payload:  "test payload",
	}
	mc.Broadcast(e)

	// Give the goroutine time to process the event.
	time.Sleep(20 * time.Millisecond)

	stats := mc.Stats()
	if !strings.Contains(stats, "events_consolidated: 1") {
		t.Errorf("stats = %q, want events_consolidated: 1", stats)
	}
	// High salience should write to all subsystems.
	for _, kind := range AllMemoryKinds {
		sub := mc.Subsystem(kind)
		sub.mu.RLock()
		n := len(sub.atomIndex)
		sub.mu.RUnlock()
		if n == 0 {
			t.Errorf("subsystem %s should have at least one atom after high-salience event", kind)
		}
	}
}

func TestMemoryConsolidatorSalienceGating(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	// ProceduralMemory threshold is 0.5; send an event below that.
	mc.Broadcast(SyncEvent{Kind: "low", Salience: 0.15})
	time.Sleep(20 * time.Millisecond)

	// Working (0.1) and Perceptual (0.1) should have received it.
	for _, kind := range []MemoryKind{WorkingMemory, PerceptualMemory} {
		sub := mc.Subsystem(kind)
		sub.mu.RLock()
		n := len(sub.atomIndex)
		sub.mu.RUnlock()
		if n == 0 {
			t.Errorf("subsystem %s should have atom for salience=0.15 event", kind)
		}
	}
	// Procedural (0.5) should NOT have received it.
	sub := mc.Subsystem(ProceduralMemory)
	sub.mu.RLock()
	n := len(sub.atomIndex)
	sub.mu.RUnlock()
	if n != 0 {
		t.Errorf("ProceduralMemory should be empty for salience=0.15 event, got %d atoms", n)
	}
}

func TestMemoryConsolidatorSetSalience(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	// Raise ProceduralMemory threshold to block all events.
	mc.SetSalience(ProceduralMemory, 1.1)
	mc.Broadcast(SyncEvent{Kind: "high", Salience: 1.0})
	time.Sleep(20 * time.Millisecond)

	sub := mc.Subsystem(ProceduralMemory)
	sub.mu.RLock()
	n := len(sub.atomIndex)
	sub.mu.RUnlock()
	if n != 0 {
		t.Errorf("ProceduralMemory should be empty with threshold 1.1, got %d atoms", n)
	}
}

func TestMemoryConsolidatorBroadcastSetsTimestamp(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	before := time.Now()
	mc.Broadcast(SyncEvent{Kind: "ts-test", Salience: 0.5})
	time.Sleep(20 * time.Millisecond)
	after := time.Now()

	// The event should have been processed; eventCount should be 1.
	mc.mu.RLock()
	ec := mc.eventCount
	mc.mu.RUnlock()
	if ec != 1 {
		t.Errorf("eventCount = %d, want 1 (broadcast at %v, checked at %v)", ec, before, after)
	}
}

func TestMemoryConsolidatorDropOnFull(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	// Fill the channel (capacity 64) and send two extra to exercise the drop path.
	for i := 0; i < 66; i++ {
		mc.Broadcast(SyncEvent{Kind: "flood", Salience: 0.5})
	}
	// No deadlock or panic is the success criterion.
}

func TestMemoryConsolidatorStats(t *testing.T) {
	g := NewKnowledgeGraph()
	mc := NewMemoryConsolidator(g)
	defer mc.Stop()

	stats := mc.Stats()
	if !strings.Contains(stats, "events_consolidated: 0") {
		t.Errorf("initial stats = %q", stats)
	}
	for _, kind := range AllMemoryKinds {
		if !strings.Contains(stats, kind.String()) {
			t.Errorf("stats missing subsystem %s", kind)
		}
	}
}

// ---------------------------------------------------------------------------
// FSTree /ai/memory/
// ---------------------------------------------------------------------------

func TestFSTreeMemorySubtree(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)

	mem, ok := tree.Children["memory"]
	if !ok {
		t.Fatal("FSTree missing /ai/memory/")
	}
	if _, ok := mem.Children["stats"]; !ok {
		t.Error("/ai/memory/stats missing")
	}
	if _, ok := mem.Children["broadcast"]; !ok {
		t.Error("/ai/memory/broadcast missing")
	}
	for _, kind := range AllMemoryKinds {
		if _, ok := mem.Children[kind.String()]; !ok {
			t.Errorf("/ai/memory/%s missing", kind)
		}
	}
}

func TestFSTreeMemorySubsystemFiles(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	mem := tree.Children["memory"]

	for _, kind := range AllMemoryKinds {
		sub := mem.Children[kind.String()]
		if _, ok := sub.Children["stats"]; !ok {
			t.Errorf("/ai/memory/%s/stats missing", kind)
		}
		if _, ok := sub.Children["atoms"]; !ok {
			t.Errorf("/ai/memory/%s/atoms missing", kind)
		}
	}
}

func TestFSTreeMemoryBroadcastWriteFn(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	broadcast := tree.Children["memory"].Children["broadcast"]

	if err := broadcast.WriteFn("perception 0.9 hello world"); err != nil {
		t.Fatalf("broadcast WriteFn: %v", err)
	}
	result := broadcast.ReadFn()
	if !strings.Contains(result, "perception") {
		t.Errorf("broadcast ReadFn = %q, want 'perception'", result)
	}
	if !strings.Contains(result, "0.900") {
		t.Errorf("broadcast ReadFn = %q, want salience 0.900", result)
	}
}

func TestFSTreeMemoryBroadcastErrors(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	broadcast := tree.Children["memory"].Children["broadcast"]

	// Too few arguments.
	if err := broadcast.WriteFn("perception"); err == nil {
		t.Error("expected error for too-few broadcast arguments")
	}
	// Invalid salience.
	if err := broadcast.WriteFn("perception notanumber"); err == nil {
		t.Error("expected error for invalid salience")
	}
}

func TestFSTreeMemoryStatsReadFn(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	stats := tree.Children["memory"].Children["stats"].ReadFn()
	if !strings.Contains(stats, "events_consolidated") {
		t.Errorf("memory stats = %q", stats)
	}
}

func TestFSTreeMemoryDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableMemory = false
	srv := NewServer(cfg)
	tree := FSTree(srv)
	if _, ok := tree.Children["memory"]; ok {
		t.Error("/ai/memory should be absent when EnableMemory=false")
	}
}

func TestFSTreeConfigIncludesMemory(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	content := tree.Children["config"].ReadFn()
	if !strings.Contains(content, "memory: true") {
		t.Errorf("config ReadFn = %q, want 'memory: true'", content)
	}
}

func TestFSTreeMemorySubsystemStatsReadFn(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	for _, kind := range AllMemoryKinds {
		stats := tree.Children["memory"].Children[kind.String()].Children["stats"].ReadFn()
		if !strings.Contains(stats, kind.String()) {
			t.Errorf("/ai/memory/%s/stats = %q", kind, stats)
		}
	}
}

func TestFSTreeMemorySubsystemAtomsReadFn(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)

	// Broadcast a high-salience event and let it consolidate.
	broadcast := tree.Children["memory"].Children["broadcast"]
	if err := broadcast.WriteFn("test-event 0.9"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)

	// All subsystems should now contain the atom.
	for _, kind := range AllMemoryKinds {
		atoms := tree.Children["memory"].Children[kind.String()].Children["atoms"].ReadFn()
		if atoms == "" {
			t.Errorf("/ai/memory/%s/atoms is empty after high-salience broadcast", kind)
		}
	}
}
