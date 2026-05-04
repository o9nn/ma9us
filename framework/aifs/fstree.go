package aifs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FSNode represents a node in the synthetic AI filesystem tree.
type FSNode struct {
	Name     string
	Mode     uint32
	Children map[string]*FSNode
	ReadFn   func() string
	WriteFn  func(data string) error
	parent   *FSNode
}

const (
	ModeDir  = 0040000 | 0555
	ModeFile = 0444
	ModeRW   = 0666
)

// FSTree builds the complete AI filesystem tree from a server instance.
func FSTree(srv *Server) *FSNode {
	root := dir("ai")

	features := dir("features")
	features.Children["chat"] = file("chat", func() string {
		return "name: chat\ntype: conversational\ndescription: Multi-turn conversational AI with session state\nproviders: anthropic, openai, google, cohere\n"
	})
	features.Children["complete"] = file("complete", func() string {
		return "name: complete\ntype: code-completion\ndescription: Code completion and generation\nproviders: anthropic, openai\n"
	})
	features.Children["embed"] = file("embed", func() string {
		return "name: embed\ntype: embeddings\ndescription: Text embeddings for semantic search\nproviders: openai, cohere, google\n"
	})
	features.Children["rag"] = file("rag", func() string {
		return "name: rag\ntype: retrieval-augmented-generation\ndescription: Query documents with embeddings + BM25 ranking\nproviders: all (via embedding provider)\n"
	})
	root.Children["features"] = features

	sessions := dir("sessions")
	newDir := dir("new")
	newDir.Children["ctl"] = rwfile("ctl",
		func() string { return "write a username to create a new session\n" },
		func(data string) error {
			user := strings.TrimSpace(data)
			if user == "" { user = "anonymous" }
			sess := srv.NewSession(user)
			newDir.Children["ctl"].ReadFn = func() string {
				return fmt.Sprintf("/ai/sessions/%d\n", sess.ID)
			}
			return nil
		},
	)
	sessions.Children["new"] = newDir
	sessions.ReadFn = func() string {
		var b strings.Builder
		for _, sess := range srv.ListSessions() {
			fmt.Fprintf(&b, "%d\n", sess.ID)
		}
		return b.String()
	}
	root.Children["sessions"] = sessions

	models := dir("models")
	models.Children["list"] = file("list", func() string { return srv.models.List() })
	models.Children["active"] = rwfile("active",
		func() string { return srv.models.Active() + "\n" },
		func(data string) error { return srv.models.SetActive(strings.TrimSpace(data)) },
	)
	root.Children["models"] = models

	if srv.config.EnableKnowledge {
		knowledge := dir("knowledge")
		knowledge.Children["stats"] = file("stats", func() string { return srv.graph.Stats() })
		concepts := dir("concepts")
		concepts.ReadFn = func() string { return strings.Join(srv.graph.ListConcepts(), "\n") + "\n" }
		knowledge.Children["concepts"] = concepts
		relations := dir("relations")
		relations.ReadFn = func() string {
			var b strings.Builder
			for _, link := range srv.graph.GetByType(InheritanceLink) { fmt.Fprintf(&b, "%s\n", link) }
			for _, link := range srv.graph.GetByType(SimilarityLink) { fmt.Fprintf(&b, "%s\n", link) }
			for _, link := range srv.graph.GetByType(EvaluationLink) { fmt.Fprintf(&b, "%s\n", link) }
			return b.String()
		}
		knowledge.Children["relations"] = relations
		var lastQueryResult string
		knowledge.Children["query"] = rwfile("query",
			func() string { return lastQueryResult },
			func(data string) error {
				parts := strings.Fields(data)
				if len(parts) < 2 { return fmt.Errorf("query format: TypeName target_id [slot]") }
				targetID, err := strconv.ParseUint(parts[1], 10, 64)
				if err != nil { return err }
				slot := -1
				if len(parts) >= 3 { slot, err = strconv.Atoi(parts[2]); if err != nil { return err } }
				var linkType AtomType
				switch parts[0] {
				case "InheritanceLink": linkType = InheritanceLink
				case "SimilarityLink": linkType = SimilarityLink
				case "EvaluationLink": linkType = EvaluationLink
				case "MemberLink": linkType = MemberLink
				default: return fmt.Errorf("unknown link type: %s", parts[0])
				}
				results := srv.graph.Query(linkType, targetID, slot)
				var b strings.Builder
				for _, r := range results { fmt.Fprintf(&b, "%s\n", r) }
				lastQueryResult = b.String()
				return nil
			},
		)
		root.Children["knowledge"] = knowledge
	}

	if srv.config.EnableTopology {
		topology := dir("topology")
		topology.Children["vertices"] = file("vertices", func() string { return srv.topology.ListVertices() })
		topology.Children["edges"] = file("edges", func() string { return srv.topology.ListEdges() })
		topology.Children["cells"] = file("cells", func() string { return srv.topology.ListCells() })
		topology.Children["lifecycle"] = rwfile("lifecycle",
			func() string { return srv.topology.LifecycleStatus() },
			func(data string) error {
				phase := strings.TrimSpace(data)
				for p, name := range phaseNames {
					if name == phase { srv.topology.SetPhase(p); return nil }
				}
				return fmt.Errorf("unknown phase: %s", phase)
			},
		)
		root.Children["topology"] = topology
	}

	root.Children["config"] = rwfile("config",
		func() string {
			return fmt.Sprintf("listen: %s\ndefault_model: %s\nknowledge: %v\ntopology: %v\n",
				srv.config.ListenAddr, srv.config.DefaultModel,
				srv.config.EnableKnowledge, srv.config.EnableTopology)
		},
		func(data string) error {
			for _, line := range strings.Split(data, "\n") {
				parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
				if len(parts) != 2 { continue }
				if strings.TrimSpace(parts[0]) == "default_model" {
					srv.config.DefaultModel = strings.TrimSpace(parts[1])
				}
			}
			return nil
		},
	)
	return root
}

// SessionFSNode creates a filesystem subtree for a specific session.
func SessionFSNode(sess *Session) *FSNode {
	d := dir(strconv.FormatUint(sess.ID, 10))
	d.Children["ctl"] = rwfile("ctl",
		func() string {
			if len(sess.Messages) == 0 { return "" }
			return sess.Messages[len(sess.Messages)-1].Content + "\n"
		},
		func(data string) error {
			query := strings.TrimSpace(data)
			if query == "" { return nil }
			sess.AddMessage("user", query)
			sess.AddMessage("assistant", fmt.Sprintf("[%s] Response to: %s", sess.Model, query))
			return nil
		},
	)
	d.Children["history"] = file("history", func() string { return sess.History() })
	d.Children["context"] = rwfile("context",
		func() string {
			sess.mu.RLock(); defer sess.mu.RUnlock()
			var b strings.Builder
			for k, v := range sess.Context { fmt.Fprintf(&b, "%s=%s\n", k, v) }
			return b.String()
		},
		func(data string) error {
			for _, line := range strings.Split(data, "\n") {
				parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
				if len(parts) == 2 { sess.SetContext(parts[0], parts[1]) }
			}
			return nil
		},
	)
	d.Children["meta"] = file("meta", func() string { return sess.Meta() })
	return d
}

func dir(name string) *FSNode {
	return &FSNode{Name: name, Mode: ModeDir, Children: make(map[string]*FSNode)}
}

func file(name string, readFn func() string) *FSNode {
	return &FSNode{Name: name, Mode: ModeFile, ReadFn: readFn}
}

func rwfile(name string, readFn func() string, writeFn func(string) error) *FSNode {
	return &FSNode{Name: name, Mode: ModeRW, ReadFn: readFn, WriteFn: writeFn}
}

func (n *FSNode) Stat() string {
	modeStr := "-"
	if n.Mode&0040000 != 0 { modeStr = "d" }
	if n.Mode&0400 != 0 { modeStr += "r" } else { modeStr += "-" }
	if n.Mode&0200 != 0 { modeStr += "w" } else { modeStr += "-" }
	size := 0
	if n.ReadFn != nil { size = len(n.ReadFn()) }
	return fmt.Sprintf("%s %8d %s %s", modeStr, size, time.Now().Format("Jan _2 15:04"), n.Name)
}
