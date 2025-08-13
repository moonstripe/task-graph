package graph

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"slices"

	"github.com/google/uuid"
)

// -- WHAT I WANT TO IMPLEMENT --

type Node interface {
	Id() uuid.UUID
	Label() string
}

type GraphOperable interface {
	// Node operations
	AddNode(n Node)
	GetNode(id uuid.UUID) (Node, error)
	GetNodesCount() int
	GetAllNodes() []Node

	// Edge operations
	AddEdge(from, to Node)
	RemoveEdge(from, to Node)
	HasEdge(from, to Node) bool

	// Access adjacency
	GetAdjacency() map[Node][]Node
	GetAdjacencyFromNode(u Node) []Node

	// Graph properties
	IsDirected() bool
}

// simple node
type SimpleNode struct {
	ID uuid.UUID `json:"id"`
}

func NewSimpleNode() SimpleNode {
	return SimpleNode{ID: uuid.New()}
}

func (sN SimpleNode) Id() uuid.UUID {
	return sN.ID
}

func (sN SimpleNode) Label() string {
	return sN.ID.String()[:8]
}

// graph structure

type SimpleDigraph struct {
	nodes []Node
	adj   map[Node][]Node
}

func NewSimpleDigraph() *SimpleDigraph {
	return &SimpleDigraph{
		nodes: []Node{},
		adj:   make(map[Node][]Node, 0),
	}
}

func NewSimpleDigraphWithSize(n int) *SimpleDigraph {
	sd := &SimpleDigraph{
		nodes: make([]Node, 0),
		adj:   make(map[Node][]Node),
	}
	for range n {
		sd.AddNode(NewSimpleNode())
	}
	return sd
}

func (g *SimpleDigraph) AddNode(n Node) {
	g.nodes = append(g.nodes, n)
	g.adj[n] = make([]Node, 0)
}

func (g *SimpleDigraph) GetNode(id uuid.UUID) (Node, error) {
	for _, node := range g.GetAllNodes() {
		if id == node.Id() {
			return node, nil
		}
	}

	return nil, errors.New("could not find node here")
}

func (g *SimpleDigraph) GetNodesCount() int {
	return len(g.nodes)
}

func (g *SimpleDigraph) GetAllNodes() []Node {
	return g.nodes
}

func (g *SimpleDigraph) AddEdge(from, to Node) {
	g.adj[from] = append(g.adj[from], to)
}

func (g *SimpleDigraph) RemoveEdge(from, to Node) {
	newAdj := make([]Node, 0)
	for _, node := range g.adj[from] {
		if node != to {
			newAdj = append(newAdj, node)
		}
	}
	g.adj[from] = newAdj
}

func (g *SimpleDigraph) HasEdge(from, to Node) bool {
	return slices.Contains(g.adj[from], to)
}

func (g *SimpleDigraph) GetAdjacency() map[Node][]Node {
	return g.adj
}

func (g *SimpleDigraph) GetAdjacencyFromNode(u Node) []Node {
	return g.adj[u]
}

func (g *SimpleDigraph) IsDirected() bool {
	return true
}

func SaveDAGDotAndPNG(sD SimpleDigraph, filenameBase string, useNames bool) error {
	dotFilename := filenameBase + ".dot"
	pngFilename := filenameBase + ".png"

	f, err := os.Create(dotFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// index map: Node UUID -> 0..n-1
	idx := make(map[uuid.UUID]int)
	nodes := sD.GetAllNodes()
	for i, node := range nodes {
		idx[node.Id()] = i
	}

	fmt.Fprintln(f, "digraph G {")
	fmt.Fprintln(f, "  rankdir=LR;")
	fmt.Fprintln(f, "  node [shape=box, style=rounded, fontsize=12];")

	// node headers
	for i, node := range nodes {
		fmt.Fprintf(f, "  %d [label=\"%s\"];\n", i, node.Label())
	}

	// edges
	for i, u := range nodes {
		for _, v := range sD.GetAdjacencyFromNode(u) {
			j := idx[v.Id()]
			fmt.Fprintf(f, "  %d -> %d;\n", i, j)
		}
	}
	fmt.Fprintln(f, "}")

	if err := exec.Command("dot", "-Tpng", dotFilename, "-o", pngFilename).Run(); err != nil {
		return fmt.Errorf("failed to run dot: %w", err)
	}
	fmt.Printf("Saved DAG as %s and %s\n", dotFilename, pngFilename)
	return nil
}

// Like SaveDAGDotAndPNG, but lets you pass "ranks" (layers) to show nodes on the same rank.
func SaveDAGDotAndPNGWithRanks(sD SimpleDigraph, filenameBase string, useNames bool, ranks [][]Node) error {
	dotFilename := filenameBase + ".dot"
	pngFilename := filenameBase + ".png"

	f, err := os.Create(dotFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// index map: Node UUID -> 0..n-1
	idx := make(map[uuid.UUID]int)
	nodes := sD.GetAllNodes()
	for i, node := range nodes {
		idx[node.Id()] = i
	}

	fmt.Fprintln(f, "digraph G {")
	fmt.Fprintln(f, "  rankdir=LR;")
	fmt.Fprintln(f, "  node [shape=box, style=rounded, fontsize=12];")

	// node headers

	for i, node := range nodes {
		fmt.Fprintf(f, "  %d [label=\"%s\"];\n", i, node.Label())
	}

	// optional rank constraints (layers)
	for ri, layer := range ranks {
		fmt.Fprintf(f, "  { rank=same; // layer %d\n", ri)
		for _, n := range layer {
			if j, ok := idx[n.Id()]; ok {
				fmt.Fprintf(f, "    %d;\n", j)
			}
		}
		fmt.Fprintln(f, "  }")
	}

	// edges
	for i, u := range nodes {
		for _, v := range sD.GetAdjacencyFromNode(u) {
			j := idx[v.Id()]
			fmt.Fprintf(f, "  %d -> %d;\n", i, j)
		}
	}
	fmt.Fprintln(f, "}")

	if err := exec.Command("dot", "-Tpng", dotFilename, "-o", pngFilename).Run(); err != nil {
		return fmt.Errorf("failed to run dot: %w", err)
	}
	fmt.Printf("Saved DAG as %s and %s\n", dotFilename, pngFilename)
	return nil
}

// Build a chain that follows the topological order exactly: u0->u1->...->uk
func BuildLinearChain(order []Node) *SimpleDigraph {
	g := NewSimpleDigraph()
	// Add nodes in the order they should appear
	for _, n := range order {
		g.AddNode(n)
	}
	// Connect as a straight line
	for i := 0; i+1 < len(order); i++ {
		g.AddEdge(order[i], order[i+1])
	}
	return g
}

// Build a layered DAG: all edges go from layer i to layer i+1 (complete bipartite).
// This makes the parallel structure explicit and easy to render.
func BuildLayeredDag(layers [][]Node) *SimpleDigraph {
	g := NewSimpleDigraph()
	// Preserve node ordering by layers
	for _, layer := range layers {
		for _, n := range layer {
			g.AddNode(n)
		}
	}
	// Connect every node in layer i to every node in layer i+1
	if len(layers) >= 2 {
		// Build a quick index of where nodes ended up in g.nodes
		// (Not strictly required for edges; we can use the Node references directly.)
		for li := 0; li+1 < len(layers); li++ {
			fromLayer := layers[li]
			toLayer := layers[li+1]
			for _, u := range fromLayer {
				for _, v := range toLayer {
					g.AddEdge(u, v)
				}
			}
		}
	}
	return g
}

// picks random permutation to avoid cycles
func GenerateRandomSDAG(n int, p float64, rnd *rand.Rand) *SimpleDigraph {
	g := NewSimpleDigraph()

	// add n unnamed nodes
	for range n {
		g.AddNode(NewSimpleNode())
	}

	// permutation to impose a topological order
	perm := rnd.Perm(n)
	pos := make([]int, n)
	for i, v := range perm {
		pos[v] = i
	}

	// helper to grab node by index quickly
	nodeAt := func(i int) Node { return g.nodes[i] }

	for u := range n {
		for v := range n {
			if u != v && pos[u] < pos[v] {
				if rnd.Float64() < p {
					g.AddEdge(nodeAt(u), nodeAt(v))
				}
			}
		}
	}

	return g
}

// topological sort (Kahn)
func TopologicalSortKahn(g GraphOperable) ([]Node, bool) {
	indeg := make(map[uuid.UUID]int)

	// 1) init indegree for ALL nodes to 0
	for _, u := range g.GetAllNodes() {
		indeg[u.Id()] = 0
	}
	// 2) count incoming edges
	for _, u := range g.GetAllNodes() {
		for _, v := range g.GetAdjacencyFromNode(u) {
			indeg[v.Id()]++
		}
	}

	// 3) queue = all nodes with indeg 0
	queue := make([]Node, 0)
	for _, u := range g.GetAllNodes() {
		if indeg[u.Id()] == 0 {
			queue = append(queue, u)
		}
	}

	order := make([]Node, 0, g.GetNodesCount())
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)
		for _, v := range g.GetAdjacencyFromNode(u) {
			indeg[v.Id()]--
			if indeg[v.Id()] == 0 {
				queue = append(queue, v)
			}
		}
	}
	if len(order) != g.GetNodesCount() {
		return nil, false // a cycle exists
	}
	return order, true
}

// execution layers (parallelizable sets)
func KahnLayers(g GraphOperable) [][]Node {
	indeg := make(map[uuid.UUID]int)
	for _, u := range g.GetAllNodes() {
		indeg[u.Id()] = 0
	}
	for _, u := range g.GetAllNodes() {
		for _, v := range g.GetAdjacencyFromNode(u) {
			indeg[v.Id()]++
		}
	}

	zero := make([]Node, 0)
	for _, u := range g.GetAllNodes() {
		if indeg[u.Id()] == 0 {
			zero = append(zero, u)
		}
	}

	layers := make([][]Node, 0)
	for len(zero) > 0 {
		layers = append(layers, zero)
		next := make([]Node, 0)
		for _, u := range zero {
			for _, v := range g.GetAdjacencyFromNode(u) {
				indeg[v.Id()]--
				if indeg[v.Id()] == 0 {
					next = append(next, v)
				}
			}
		}
		zero = next
	}
	return layers
}
