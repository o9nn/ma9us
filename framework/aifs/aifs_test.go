package aifs

import (
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil)
	if s == nil { t.Fatal("NewServer returned nil") }
	if len(s.models.models) == 0 { t.Error("default models not registered") }
}

func TestSessionLifecycle(t *testing.T) {
	s := NewServer(nil)
	sess := s.NewSession("testuser")
	if sess.ID == 0 { t.Error("session ID should be > 0") }
	if sess.User != "testuser" { t.Errorf("user = %q, want %q", sess.User, "testuser") }
	sess.AddMessage("user", "hello")
	sess.AddMessage("assistant", "hi there")
	history := sess.History()
	if !strings.Contains(history, "hello") { t.Error("history missing user message") }
	if !strings.Contains(history, "hi there") { t.Error("history missing assistant message") }
	got, ok := s.GetSession(sess.ID)
	if !ok { t.Error("GetSession failed") }
	if got.ID != sess.ID { t.Error("GetSession returned wrong session") }
	if len(s.ListSessions()) != 1 { t.Errorf("ListSessions = %d, want 1", len(s.ListSessions())) }
}

func TestSessionMeta(t *testing.T) {
	s := NewServer(nil)
	sess := s.NewSession("bob")
	meta := sess.Meta()
	if !strings.Contains(meta, "bob") { t.Error("meta missing user") }
	if !strings.Contains(meta, "messages: 0") { t.Error("meta should show 0 messages") }
}

func TestModelRegistry(t *testing.T) {
	r := NewModelRegistry()
	r.Register(ModelConfig{Name: "test-model", Provider: "test", MaxTokens: 1000})
	if r.Active() != "test-model" { t.Errorf("active = %q", r.Active()) }
	m, ok := r.Get("test-model")
	if !ok { t.Error("Get failed") }
	if m.Provider != "test" { t.Errorf("provider = %q", m.Provider) }
	r.Register(ModelConfig{Name: "other-model", Provider: "test", MaxTokens: 2000})
	if err := r.SetActive("other-model"); err != nil { t.Errorf("SetActive: %v", err) }
	if r.Active() != "other-model" { t.Error("active not changed") }
	if err := r.SetActive("nonexistent"); err == nil { t.Error("should fail for nonexistent") }
}

func TestKnowledgeGraph(t *testing.T) {
	g := NewKnowledgeGraph()
	cat := g.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := g.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	g.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})
	cat2 := g.AddNode(ConceptNode, "Cat", TruthValue{0.8, 0.5})
	if cat2.ID != cat.ID { t.Error("duplicate node should return existing") }
	if len(g.ListConcepts()) != 2 { t.Errorf("concepts = %d", len(g.ListConcepts())) }
	if len(g.GetByName("Cat")) != 1 { t.Error("GetByName failed") }
	if len(g.GetIncoming(animal.ID)) != 1 { t.Error("incoming failed") }
	if len(g.Query(InheritanceLink, animal.ID, 1)) != 1 { t.Error("query failed") }
	if !strings.Contains(g.Stats(), "total_atoms: 3") { t.Error("stats wrong") }
}

func TestTruthValueMerge(t *testing.T) {
	merged := TruthValue{0.8, 0.9}.Merge(TruthValue{0.6, 0.7})
	if merged.Strength < 0.6 || merged.Strength > 0.8 { t.Errorf("strength %.3f", merged.Strength) }
	if merged.Confidence <= 0 || merged.Confidence >= 1 { t.Errorf("confidence %.3f", merged.Confidence) }
}

func TestTopology(t *testing.T) {
	top := NewTopology()
	id0 := top.AddVertex("alice", [4]float64{1, 0, 0, 0})
	id1 := top.AddVertex("bob", [4]float64{0, 1, 0, 0})
	id2 := top.AddVertex("carol", [4]float64{0, 0, 1, 0})
	top.AddEdge(id0, id1, 1.0); top.AddEdge(id1, id2, 0.5); top.AddEdge(id0, id2, 0.8)
	top.ComputeInfluence()
	top.AddCell("team-alpha", []int{id0, id1, id2})
	if !strings.Contains(top.ListVertices(), "alice") { t.Error("missing alice") }
	if !strings.Contains(top.ListEdges(), "1.000") { t.Error("missing weight") }
	if !strings.Contains(top.ListCells(), "team-alpha") { t.Error("missing cell") }
	top.SetPhase(PhaseGrowth)
	if top.Phase() != PhaseGrowth { t.Error("phase wrong") }
	if !strings.Contains(top.LifecycleStatus(), "growth") { t.Error("status wrong") }
}

func TestStereographicProject(t *testing.T) {
	p := StereographicProject([4]float64{1, 2, 3, 0})
	if p[0] != 1 || p[1] != 2 || p[2] != 3 { t.Errorf("got %v", p) }
}

func TestGetSessionNotFound(t *testing.T) {
	s := NewServer(nil)
	_, ok := s.GetSession(999)
	if ok { t.Error("GetSession should return false for non-existent ID") }
}

func TestListSessionsMultiple(t *testing.T) {
	s := NewServer(nil)
	s.NewSession("alice")
	s.NewSession("bob")
	s.NewSession("carol")
	sessions := s.ListSessions()
	if len(sessions) != 3 { t.Errorf("ListSessions = %d, want 3", len(sessions)) }
}

func TestSessionEmptyHistory(t *testing.T) {
	s := NewServer(nil)
	sess := s.NewSession("user")
	if sess.History() != "" { t.Error("empty history should return empty string") }
}

func TestAtomIsLink(t *testing.T) {
	node := &Atom{Type: ConceptNode}
	link := &Atom{Type: InheritanceLink}
	if node.IsLink() { t.Error("ConceptNode should not be a link") }
	if !link.IsLink() { t.Error("InheritanceLink should be a link") }
}

func TestAtomStringNode(t *testing.T) {
	node := &Atom{Type: ConceptNode, Name: "Dog", TV: TruthValue{1.0, 0.9}}
	s := node.String()
	if !strings.Contains(s, "ConceptNode") { t.Errorf("node string = %q", s) }
	if !strings.Contains(s, "Dog") { t.Errorf("node string missing name: %q", s) }
}

func TestAtomStringLink(t *testing.T) {
	link := &Atom{Type: InheritanceLink, Outgoing: []uint64{1, 2}, TV: TruthValue{0.9, 0.8}}
	s := link.String()
	if !strings.Contains(s, "InheritanceLink") { t.Errorf("link string = %q", s) }
	if !strings.Contains(s, "0.900") { t.Errorf("link string missing strength: %q", s) }
}

func TestAtomTypeStringUnknown(t *testing.T) {
	unknown := AtomType(999)
	s := unknown.String()
	if !strings.Contains(s, "999") { t.Errorf("unknown atom type = %q", s) }
}

func TestKnowledgeGraphGetAtom(t *testing.T) {
	g := NewKnowledgeGraph()
	node := g.AddNode(ConceptNode, "Dog", TruthValue{1.0, 0.9})
	got, ok := g.GetAtom(node.ID)
	if !ok { t.Fatal("GetAtom should find existing atom") }
	if got.Name != "Dog" { t.Errorf("name = %q, want Dog", got.Name) }
	_, ok = g.GetAtom(9999)
	if ok { t.Error("GetAtom should return false for non-existent ID") }
}

func TestTruthValueMergeZeroConfidence(t *testing.T) {
	merged := TruthValue{0.8, 0}.Merge(TruthValue{0.6, 0})
	if merged.Strength != 0.5 { t.Errorf("strength = %.3f, want 0.5", merged.Strength) }
	if merged.Confidence != 0 { t.Errorf("confidence = %.3f, want 0", merged.Confidence) }
}

func TestKnowledgeGraphQueryBySlot(t *testing.T) {
	g := NewKnowledgeGraph()
	cat := g.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := g.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	g.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})
	// Query by slot 0 (subject)
	results := g.Query(InheritanceLink, cat.ID, 0)
	if len(results) != 1 { t.Errorf("query slot 0 = %d, want 1", len(results)) }
	// Query by slot 1 (object)
	results = g.Query(InheritanceLink, animal.ID, 1)
	if len(results) != 1 { t.Errorf("query slot 1 = %d, want 1", len(results)) }
	// Query by slot out of range returns nothing
	results = g.Query(InheritanceLink, cat.ID, 5)
	if len(results) != 0 { t.Errorf("query slot 5 = %d, want 0", len(results)) }
	// Query with slot=-1 (any position): should find link containing cat.ID
	results = g.Query(InheritanceLink, cat.ID, -1)
	if len(results) != 1 { t.Errorf("query slot -1 = %d, want 1", len(results)) }
	// Query with slot=-1 for a target not in the link should return nothing
	notExist := g.AddNode(ConceptNode, "Phantom", TruthValue{1.0, 0.5})
	results = g.Query(InheritanceLink, notExist.ID, -1)
	if len(results) != 0 { t.Errorf("query slot -1 miss = %d, want 0", len(results)) }
}

func TestLifecyclePhaseStringUnknown(t *testing.T) {
	unknown := LifecyclePhase(999)
	s := unknown.String()
	if !strings.Contains(s, "999") { t.Errorf("unknown phase = %q", s) }
}

func TestStereographicProjectPole(t *testing.T) {
	// w = 1 is exactly on the pole; the guard clamps denom to 1e-10 to avoid division by zero.
	p := StereographicProject([4]float64{1, 2, 3, 1})
	// Should not panic; values will be very large but finite
	if p[0] == 0 && p[1] == 0 && p[2] == 0 { t.Error("near-pole projection should be non-zero") }
}

func TestComputeInfluenceEmpty(t *testing.T) {
	top := NewTopology()
	// Should not panic on empty topology
	top.ComputeInfluence()
}

func TestComputeInfluenceNoEdges(t *testing.T) {
	top := NewTopology()
	top.AddVertex("solo", [4]float64{1, 0, 0, 0})
	// maxDeg == 0, influence should remain 0
	top.ComputeInfluence()
	if !strings.Contains(top.ListVertices(), "solo") { t.Error("missing vertex") }
}
