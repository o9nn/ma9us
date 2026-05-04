package aifs

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

type LifecyclePhase int

const (
	PhaseEmergence LifecyclePhase = iota
	PhaseGrowth
	PhaseMaturity
	PhaseDecay
	PhaseDissolution
	PhaseRebirth
)

var phaseNames = map[LifecyclePhase]string{
	PhaseEmergence: "emergence", PhaseGrowth: "growth",
	PhaseMaturity: "maturity", PhaseDecay: "decay",
	PhaseDissolution: "dissolution", PhaseRebirth: "rebirth",
}

func (p LifecyclePhase) String() string {
	if name, ok := phaseNames[p]; ok { return name }
	return fmt.Sprintf("phase(%d)", int(p))
}

type Vertex struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Position  [4]float64 `json:"position"`
	Influence float64    `json:"influence"`
}

type Edge struct {
	From   int     `json:"from"`
	To     int     `json:"to"`
	Weight float64 `json:"weight"`
}

type Cell struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Vertices []int  `json:"vertices"`
}

type Topology struct {
	mu        sync.RWMutex
	vertices  []Vertex
	edges     []Edge
	cells     []Cell
	phase     LifecyclePhase
	edgeIndex map[int][]int
}

func NewTopology() *Topology {
	return &Topology{
		vertices: make([]Vertex, 0), edges: make([]Edge, 0),
		cells: make([]Cell, 0), phase: PhaseEmergence,
		edgeIndex: make(map[int][]int),
	}
}

func (t *Topology) AddVertex(name string, pos [4]float64) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := len(t.vertices)
	t.vertices = append(t.vertices, Vertex{ID: id, Name: name, Position: pos})
	return id
}

func (t *Topology) AddEdge(from, to int, weight float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	idx := len(t.edges)
	t.edges = append(t.edges, Edge{From: from, To: to, Weight: weight})
	t.edgeIndex[from] = append(t.edgeIndex[from], idx)
	t.edgeIndex[to] = append(t.edgeIndex[to], idx)
}

func (t *Topology) AddCell(name string, vertices []int) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := len(t.cells)
	t.cells = append(t.cells, Cell{ID: id, Name: name, Vertices: vertices})
	return id
}

func (t *Topology) ComputeInfluence() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.vertices) == 0 { return }
	maxDeg := 0.0
	deg := make(map[int]float64)
	for _, e := range t.edges { deg[e.From] += e.Weight; deg[e.To] += e.Weight }
	for _, d := range deg { if d > maxDeg { maxDeg = d } }
	if maxDeg > 0 {
		for i := range t.vertices { t.vertices[i].Influence = deg[t.vertices[i].ID] / maxDeg }
	}
}

func (t *Topology) SetPhase(p LifecyclePhase) { t.mu.Lock(); defer t.mu.Unlock(); t.phase = p }
func (t *Topology) Phase() LifecyclePhase { t.mu.RLock(); defer t.mu.RUnlock(); return t.phase }

func StereographicProject(p [4]float64) [3]float64 {
	denom := 1.0 - p[3]
	if math.Abs(denom) < 1e-10 { denom = 1e-10 }
	return [3]float64{p[0] / denom, p[1] / denom, p[2] / denom}
}

func (t *Topology) ListVertices() string {
	t.mu.RLock(); defer t.mu.RUnlock()
	var b strings.Builder
	for _, v := range t.vertices {
		p3d := StereographicProject(v.Position)
		fmt.Fprintf(&b, "%d\t%s\t%.3f\t[%.2f, %.2f, %.2f]\n", v.ID, v.Name, v.Influence, p3d[0], p3d[1], p3d[2])
	}
	return b.String()
}

func (t *Topology) ListEdges() string {
	t.mu.RLock(); defer t.mu.RUnlock()
	var b strings.Builder
	for _, e := range t.edges { fmt.Fprintf(&b, "%d\t%d\t%.3f\n", e.From, e.To, e.Weight) }
	return b.String()
}

func (t *Topology) ListCells() string {
	t.mu.RLock(); defer t.mu.RUnlock()
	var b strings.Builder
	for _, c := range t.cells { fmt.Fprintf(&b, "%d\t%s\t%v\n", c.ID, c.Name, c.Vertices) }
	return b.String()
}

func (t *Topology) LifecycleStatus() string {
	t.mu.RLock(); defer t.mu.RUnlock()
	return fmt.Sprintf("phase: %s\nvertices: %d\nedges: %d\ncells: %d\n", t.phase, len(t.vertices), len(t.edges), len(t.cells))
}
