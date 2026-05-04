package aifs

import (
	"fmt"
	"strings"
	"testing"
)

func TestFSTree(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	if tree.Name != "ai" { t.Errorf("root = %q", tree.Name) }
	for _, name := range []string{"features", "sessions", "models", "knowledge", "topology", "config"} {
		if _, ok := tree.Children[name]; !ok { t.Errorf("missing: %s", name) }
	}
}

func TestFSTreeFeatures(t *testing.T) {
	tree := FSTree(NewServer(nil))
	for _, name := range []string{"chat", "complete", "embed", "rag"} {
		node, ok := tree.Children["features"].Children[name]
		if !ok { t.Errorf("missing feature: %s", name); continue }
		if !strings.Contains(node.ReadFn(), "name: "+name) { t.Errorf("%s missing name", name) }
	}
}

func TestFSTreeModels(t *testing.T) {
	tree := FSTree(NewServer(nil))
	models := tree.Children["models"]
	if !strings.Contains(models.Children["list"].ReadFn(), "claude") { t.Error("list missing claude") }
	if !strings.Contains(models.Children["active"].ReadFn(), "claude") { t.Error("active wrong") }
	if err := models.Children["active"].WriteFn("gpt-4o"); err != nil { t.Error(err) }
	if !strings.Contains(models.Children["active"].ReadFn(), "gpt-4o") { t.Error("not updated") }
}

func TestFSTreeSessions(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	if err := tree.Children["sessions"].Children["new"].Children["ctl"].WriteFn("alice"); err != nil {
		t.Fatal(err)
	}
	sess := srv.ListSessions()
	if len(sess) != 1 { t.Fatalf("got %d sessions", len(sess)) }
	node := SessionFSNode(sess[0])
	if err := node.Children["ctl"].WriteFn("what is 9P?"); err != nil { t.Fatal(err) }
	if !strings.Contains(node.Children["history"].ReadFn(), "what is 9P?") { t.Error("history missing") }
	if err := node.Children["context"].WriteFn("pwd=/home/user"); err != nil { t.Fatal(err) }
	if !strings.Contains(node.Children["context"].ReadFn(), "pwd=/home/user") { t.Error("context missing") }
	if !strings.Contains(node.Children["meta"].ReadFn(), "alice") { t.Error("meta missing user") }
}

func TestFSTreeKnowledge(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	kg := tree.Children["knowledge"]
	cat := srv.graph.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := srv.graph.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	srv.graph.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})
	if !strings.Contains(kg.Children["stats"].ReadFn(), "total_atoms: 3") { t.Error("stats wrong") }
	if !strings.Contains(kg.Children["concepts"].ReadFn(), "Cat") { t.Error("concepts missing") }
	if !strings.Contains(kg.Children["relations"].ReadFn(), "InheritanceLink") { t.Error("relations missing") }
}

func TestFSTreeTopology(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	top := tree.Children["topology"]
	srv.topology.AddVertex("node-a", [4]float64{1, 0, 0, 0})
	srv.topology.AddVertex("node-b", [4]float64{0, 1, 0, 0})
	srv.topology.AddEdge(0, 1, 1.0)
	srv.topology.AddCell("team-x", []int{0, 1})
	if !strings.Contains(top.Children["vertices"].ReadFn(), "node-a") { t.Error("missing vertex") }
	if !strings.Contains(top.Children["edges"].ReadFn(), "1.000") { t.Error("missing edge weight") }
	if !strings.Contains(top.Children["cells"].ReadFn(), "team-x") { t.Error("missing cell") }
	if err := top.Children["lifecycle"].WriteFn("growth"); err != nil { t.Error(err) }
	if !strings.Contains(top.Children["lifecycle"].ReadFn(), "growth") { t.Error("phase not set") }
}

func TestFSNodeStat(t *testing.T) {
	node := file("test", func() string { return "hello" })
	stat := node.Stat()
	if !strings.Contains(stat, "test") { t.Error("missing name") }
	if !strings.Contains(stat, "5") { t.Error("missing size") }
}

func TestFSNodeStatDir(t *testing.T) {
	d := dir("mydir")
	stat := d.Stat()
	if !strings.Contains(stat, "mydir") { t.Error("missing name") }
	if stat[0] != 'd' { t.Errorf("dir stat should start with 'd', got %q", string(stat[0])) }
}

func TestFSTreeConfigReadWrite(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	cfg := tree.Children["config"]
	// ReadFn includes listen addr and default model
	content := cfg.ReadFn()
	if !strings.Contains(content, ":5641") { t.Error("config missing listen addr") }
	if !strings.Contains(content, "claude") { t.Error("config missing default model") }
	// WriteFn can update default_model
	if err := cfg.WriteFn("default_model=gpt-4o\n"); err != nil { t.Fatal(err) }
	if srv.config.DefaultModel != "gpt-4o" { t.Errorf("default_model = %q, want gpt-4o", srv.config.DefaultModel) }
	// Lines without '=' are skipped without error
	if err := cfg.WriteFn("no-equals-sign\ndefault_model=gemini-2.5-pro\n"); err != nil { t.Fatal(err) }
	if srv.config.DefaultModel != "gemini-2.5-pro" { t.Errorf("default_model = %q", srv.config.DefaultModel) }
	// A key=value pair for an unknown key is silently ignored
	if err := cfg.WriteFn("listen=:9999\ndefault_model=gpt-4o-mini\n"); err != nil { t.Fatal(err) }
	if srv.config.DefaultModel != "gpt-4o-mini" { t.Errorf("default_model = %q", srv.config.DefaultModel) }
	if srv.config.ListenAddr != ":5641" { t.Errorf("listen addr unexpectedly changed to %q", srv.config.ListenAddr) }
}

func TestFSTreeSessionsAnonymous(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	// Initial ctl prompt before any session is created
	initialPrompt := tree.Children["sessions"].Children["new"].Children["ctl"].ReadFn()
	if !strings.Contains(initialPrompt, "username") { t.Errorf("initial ctl = %q", initialPrompt) }
	// Writing empty string creates an anonymous session
	if err := tree.Children["sessions"].Children["new"].Children["ctl"].WriteFn("   "); err != nil {
		t.Fatal(err)
	}
	sessions := srv.ListSessions()
	if len(sessions) != 1 { t.Fatalf("got %d sessions", len(sessions)) }
	if sessions[0].User != "anonymous" { t.Errorf("user = %q, want anonymous", sessions[0].User) }
	// After creation, ctl ReadFn returns the session path
	ctlRead := tree.Children["sessions"].Children["new"].Children["ctl"].ReadFn()
	if !strings.Contains(ctlRead, "/ai/sessions/") { t.Errorf("ctl = %q", ctlRead) }
}

func TestFSTreeSessionsListReadFn(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	srv.NewSession("user1")
	srv.NewSession("user2")
	list := tree.Children["sessions"].ReadFn()
	// Should contain at least two lines (one per session ID)
	lines := strings.Split(strings.TrimSpace(list), "\n")
	if len(lines) < 2 { t.Errorf("sessions list = %q", list) }
}

func TestFSTreeKnowledgeQuery(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	kg := tree.Children["knowledge"]
	cat := srv.graph.AddNode(ConceptNode, "Cat", TruthValue{1.0, 0.9})
	animal := srv.graph.AddNode(ConceptNode, "Animal", TruthValue{1.0, 0.9})
	srv.graph.AddLink(InheritanceLink, []uint64{cat.ID, animal.ID}, TruthValue{1.0, 0.95})

	query := kg.Children["query"]
	// Successful query with explicit slot
	if err := query.WriteFn(fmt.Sprintf("InheritanceLink %d 1", animal.ID)); err != nil { t.Fatal(err) }
	if !strings.Contains(query.ReadFn(), "InheritanceLink") { t.Error("query result missing") }

	// Successful query without slot (slot = -1), with an existing target
	if err := query.WriteFn(fmt.Sprintf("InheritanceLink %d", cat.ID)); err != nil { t.Fatal(err) }
	if !strings.Contains(query.ReadFn(), "InheritanceLink") { t.Error("slot -1 query result missing") }

	// MemberLink and EvaluationLink paths (no matching atoms, result is empty but no error)
	if err := query.WriteFn(fmt.Sprintf("MemberLink %d", cat.ID)); err != nil { t.Fatal(err) }
	if err := query.WriteFn(fmt.Sprintf("EvaluationLink %d", cat.ID)); err != nil { t.Fatal(err) }
	if err := query.WriteFn(fmt.Sprintf("SimilarityLink %d", cat.ID)); err != nil { t.Fatal(err) }
}

func TestFSTreeKnowledgeQueryErrors(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	query := tree.Children["knowledge"].Children["query"]

	// Too few parts
	if err := query.WriteFn("InheritanceLink"); err == nil { t.Error("expected error for too few parts") }
	// Invalid target ID
	if err := query.WriteFn("InheritanceLink notanumber"); err == nil { t.Error("expected error for bad ID") }
	// Invalid slot
	if err := query.WriteFn("InheritanceLink 1 notanumber"); err == nil { t.Error("expected error for bad slot") }
	// Unknown link type
	if err := query.WriteFn("FooLink 1"); err == nil { t.Error("expected error for unknown link type") }
}

func TestFSTreeLifecycleUnknownPhase(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	lifecycle := tree.Children["topology"].Children["lifecycle"]
	if err := lifecycle.WriteFn("nonexistent-phase"); err == nil {
		t.Error("expected error for unknown lifecycle phase")
	}
}

func TestFSTreeDisabledFeatures(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableKnowledge = false
	cfg.EnableTopology = false
	srv := NewServer(cfg)
	tree := FSTree(srv)
	if _, ok := tree.Children["knowledge"]; ok { t.Error("knowledge should be absent") }
	if _, ok := tree.Children["topology"]; ok { t.Error("topology should be absent") }
}

func TestFSTreeKnowledgeRelationsAllTypes(t *testing.T) {
	srv := NewServer(nil)
	tree := FSTree(srv)
	kg := tree.Children["knowledge"]
	a := srv.graph.AddNode(ConceptNode, "A", TruthValue{1.0, 0.9})
	b := srv.graph.AddNode(ConceptNode, "B", TruthValue{1.0, 0.9})
	srv.graph.AddLink(InheritanceLink, []uint64{a.ID, b.ID}, TruthValue{1.0, 0.9})
	srv.graph.AddLink(SimilarityLink, []uint64{a.ID, b.ID}, TruthValue{0.8, 0.7})
	srv.graph.AddLink(EvaluationLink, []uint64{a.ID, b.ID}, TruthValue{0.7, 0.6})
	rel := kg.Children["relations"].ReadFn()
	if !strings.Contains(rel, "InheritanceLink") { t.Error("missing InheritanceLink") }
	if !strings.Contains(rel, "SimilarityLink") { t.Error("missing SimilarityLink") }
	if !strings.Contains(rel, "EvaluationLink") { t.Error("missing EvaluationLink") }
}

func TestFSNodeStatNoReadPermission(t *testing.T) {
	// A node with mode 0 (no permissions) exercises the "no read" branch in Stat.
	n := &FSNode{Name: "secret", Mode: 0}
	stat := n.Stat()
	if !strings.Contains(stat, "secret") { t.Error("missing name in stat") }
	if stat[1] != '-' { t.Errorf("expected no-read '-', got %q", string(stat[1])) }
}

func TestFSNodeStatRWFile(t *testing.T) {
	// An rwfile (ModeRW = 0666) exercises the write-permission branch in Stat.
	n := rwfile("rw", func() string { return "data" }, func(string) error { return nil })
	stat := n.Stat()
	if !strings.Contains(stat, "rw") { t.Error("missing name") }
	if stat[2] != 'w' { t.Errorf("expected write 'w', got %q", string(stat[2])) }
}

func TestSessionFSNodeContextInvalidLine(t *testing.T) {
	srv := NewServer(nil)
	sess := srv.NewSession("user")
	node := SessionFSNode(sess)
	// A line without '=' should be silently skipped
	if err := node.Children["context"].WriteFn("no-equals\nkey=value\n"); err != nil { t.Fatal(err) }
	if sess.Context["key"] != "value" { t.Errorf("context[key] = %q", sess.Context["key"]) }
}

func TestSessionFSNodeEmptyCtl(t *testing.T) {
	srv := NewServer(nil)
	sess := srv.NewSession("alice")
	node := SessionFSNode(sess)
	// Writing empty string to ctl should be a no-op
	if err := node.Children["ctl"].WriteFn("   "); err != nil { t.Fatal(err) }
	if len(sess.Messages) != 0 { t.Error("empty write should not add messages") }
	// ReadFn with no messages returns empty string
	if node.Children["ctl"].ReadFn() != "" { t.Error("ctl ReadFn should be empty with no messages") }
	// After adding a message, ReadFn returns the last message content
	if err := node.Children["ctl"].WriteFn("hello"); err != nil { t.Fatal(err) }
	ctlContent := node.Children["ctl"].ReadFn()
	if !strings.Contains(ctlContent, "hello") { t.Errorf("ctl ReadFn = %q, want content with 'hello'", ctlContent) }
}
