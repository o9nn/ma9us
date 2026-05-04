package aifs

import (
	"fmt"
	"strings"
	"sync"
)

// TruthValue represents probabilistic belief (OpenCog-inspired).
type TruthValue struct {
	Strength   float64 `json:"strength"`
	Confidence float64 `json:"confidence"`
}

func (tv TruthValue) String() string {
	return fmt.Sprintf("<%.3f, %.3f>", tv.Strength, tv.Confidence)
}

func (tv TruthValue) Merge(other TruthValue) TruthValue {
	totalConf := tv.Confidence + other.Confidence
	if totalConf == 0 {
		return TruthValue{Strength: 0.5, Confidence: 0}
	}
	return TruthValue{
		Strength:   (tv.Strength*tv.Confidence + other.Strength*other.Confidence) / totalConf,
		Confidence: totalConf / (totalConf + 1),
	}
}

type AtomType int

const (
	ConceptNode AtomType = iota
	PredicateNode
	VariableNode
	NumberNode
	InheritanceLink
	SimilarityLink
	EvaluationLink
	MemberLink
	ListLink
)

var atomTypeNames = map[AtomType]string{
	ConceptNode: "ConceptNode", PredicateNode: "PredicateNode",
	VariableNode: "VariableNode", NumberNode: "NumberNode",
	InheritanceLink: "InheritanceLink", SimilarityLink: "SimilarityLink",
	EvaluationLink: "EvaluationLink", MemberLink: "MemberLink",
	ListLink: "ListLink",
}

func (t AtomType) String() string {
	if name, ok := atomTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("AtomType(%d)", int(t))
}

type Atom struct {
	ID       uint64     `json:"id"`
	Type     AtomType   `json:"type"`
	Name     string     `json:"name"`
	TV       TruthValue `json:"truth_value"`
	Outgoing []uint64   `json:"outgoing,omitempty"`
}

func (a *Atom) IsNode() bool { return a.Type <= NumberNode }
func (a *Atom) IsLink() bool { return a.Type > NumberNode }

func (a *Atom) String() string {
	if a.IsNode() {
		return fmt.Sprintf("(%s \"%s\" %s)", a.Type, a.Name, a.TV)
	}
	return fmt.Sprintf("(%s %v %s)", a.Type, a.Outgoing, a.TV)
}

type KnowledgeGraph struct {
	mu       sync.RWMutex
	atoms    map[uint64]*Atom
	byName   map[string][]*Atom
	byType   map[AtomType][]*Atom
	incoming map[uint64][]*Atom
	nextID   uint64
}

func NewKnowledgeGraph() *KnowledgeGraph {
	return &KnowledgeGraph{
		atoms: make(map[uint64]*Atom), byName: make(map[string][]*Atom),
		byType: make(map[AtomType][]*Atom), incoming: make(map[uint64][]*Atom),
		nextID: 1,
	}
}

func (g *KnowledgeGraph) AddNode(typ AtomType, name string, tv TruthValue) *Atom {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, existing := range g.byName[name] {
		if existing.Type == typ {
			existing.TV = existing.TV.Merge(tv)
			return existing
		}
	}
	atom := &Atom{ID: g.nextID, Type: typ, Name: name, TV: tv}
	g.nextID++
	g.atoms[atom.ID] = atom
	g.byName[name] = append(g.byName[name], atom)
	g.byType[typ] = append(g.byType[typ], atom)
	return atom
}

func (g *KnowledgeGraph) AddLink(typ AtomType, outgoing []uint64, tv TruthValue) *Atom {
	g.mu.Lock()
	defer g.mu.Unlock()
	atom := &Atom{ID: g.nextID, Type: typ, Outgoing: outgoing, TV: tv}
	g.nextID++
	g.atoms[atom.ID] = atom
	g.byType[typ] = append(g.byType[typ], atom)
	for _, targetID := range outgoing {
		g.incoming[targetID] = append(g.incoming[targetID], atom)
	}
	return atom
}

func (g *KnowledgeGraph) GetAtom(id uint64) (*Atom, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	a, ok := g.atoms[id]
	return a, ok
}

func (g *KnowledgeGraph) GetByName(name string) []*Atom {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.byName[name]
}

func (g *KnowledgeGraph) GetByType(typ AtomType) []*Atom {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.byType[typ]
}

func (g *KnowledgeGraph) GetIncoming(id uint64) []*Atom {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.incoming[id]
}

func (g *KnowledgeGraph) Query(linkType AtomType, targetID uint64, slot int) []*Atom {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var results []*Atom
	for _, link := range g.byType[linkType] {
		if slot >= 0 && slot < len(link.Outgoing) {
			if link.Outgoing[slot] == targetID {
				results = append(results, link)
			}
		} else if slot < 0 {
			for _, id := range link.Outgoing {
				if id == targetID {
					results = append(results, link)
					break
				}
			}
		}
	}
	return results
}

func (g *KnowledgeGraph) Stats() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var b strings.Builder
	fmt.Fprintf(&b, "total_atoms: %d\n", len(g.atoms))
	nodes, links := 0, 0
	for _, a := range g.atoms {
		if a.IsNode() { nodes++ } else { links++ }
	}
	fmt.Fprintf(&b, "nodes: %d\nlinks: %d\n", nodes, links)
	fmt.Fprintf(&b, "types:\n")
	for typ, atoms := range g.byType {
		fmt.Fprintf(&b, "  %s: %d\n", typ, len(atoms))
	}
	return b.String()
}

func (g *KnowledgeGraph) ListConcepts() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var names []string
	for _, a := range g.byType[ConceptNode] {
		names = append(names, a.Name)
	}
	return names
}
