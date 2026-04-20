package osexit

import (
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

type forwardVisitor struct {
	graph     *callgraph.Graph
	seen      map[*callgraph.Node]bool
	reachable map[*ssa.Function]bool
}

type backwardVisitor struct {
	graph *callgraph.Graph
	seen  map[*callgraph.Edge]bool
	roots map[*ssa.Function][]ssa.CallInstruction
}

// findReachableFunctions traverses a provided call graph forwards (from caller to callee),
// starting from a given function.
// It records all functions that are reachable from the starting function, including any
// intermediate functions.
func findReachableFunctions(graph *callgraph.Graph, start *ssa.Function) map[*ssa.Function]bool {
	v := &forwardVisitor{
		graph:     graph,
		seen:      make(map[*callgraph.Node]bool),
		reachable: make(map[*ssa.Function]bool),
	}

	v.visit(graph.Nodes[start])

	return v.reachable
}

func (v *forwardVisitor) visit(node *callgraph.Node) {
	if node == nil {
		return
	}

	if v.seen[node] {
		return
	}

	v.seen[node] = true

	if node.Func != nil {
		v.reachable[node.Func] = true
	}

	for _, out := range node.Out {
		v.visit(out.Callee)
	}
}

// findFunctionRoots traverses a given call graph backwards (from callees to callers),
// starting from a given "end" function.
// For any callee, it follows all incoming edges (caller -> callee) exactly once, terminating
// at callees that have no incoming edges - these are considered "root" callees.
// For any "root" callee, a list of call instructions is recorded - these represent places in the "root" function
// where the first call of a call chain is made (the call chain that ends with the provided "end" function).
func findFunctionRoots(graph *callgraph.Graph, start *ssa.Function) map[*ssa.Function][]ssa.CallInstruction {
	v := &backwardVisitor{
		graph: graph,
		seen:  make(map[*callgraph.Edge]bool),
		roots: make(map[*ssa.Function][]ssa.CallInstruction),
	}

	v.visit(graph.Nodes[start], nil)

	return v.roots
}

func (v *backwardVisitor) visit(node *callgraph.Node, call ssa.CallInstruction) {
	if node == nil {
		return
	}

	v.maybeAddRoot(node, call)

	for _, in := range node.In {
		v.handleEdge(in)
	}
}

func (v *backwardVisitor) maybeAddRoot(node *callgraph.Node, call ssa.CallInstruction) {
	if len(node.In) > 0 {
		// not a root node yet
		return
	}

	fn := node.Func
	if fn == nil {
		return
	}

	if call == nil {
		return
	}

	v.roots[fn] = append(v.roots[fn], call)
}

func (v *backwardVisitor) handleEdge(edge *callgraph.Edge) {
	if edge.Site == nil {
		// as per [callgraph.Edge] documentation:
		// > Site is nil for edges originating in synthetic or intrinsic
		// > functions, e.g. reflect.Value.Call or the root of the call graph.
		return
	}

	if v.seen[edge] {
		return
	}

	v.seen[edge] = true

	v.visit(edge.Caller, edge.Site)
}
