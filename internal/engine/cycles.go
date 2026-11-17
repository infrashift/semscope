package engine

import (
	"sort"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
)

// dfsFrame is one entry of the explicit depth-first search stack.
type dfsFrame struct {
	node int
	next int
}

// The colors of the depth-first search.
const (
	white uint8 = iota
	gray
	black
)

// findCycles runs an iterative three-color depth-first search over g, in
// key sequence, and returns one cycle path for each back edge. The oracles
// use the same search, so the findings match. The stacks have a fixed
// capacity of one entry for each node.
func findCycles(g *callGraph) [][]int {
	n := len(g.nodes)
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return g.nodes[order[a]].key < g.nodes[order[b]].key })
	color, onPath := make([]uint8, n), make([]int, n)
	stack, pathNodes := make([]dfsFrame, 0, n), make([]int, 0, n)
	var cycles [][]int
	maxSteps := 2*n + g.edges + 1
	for _, root := range order {
		if color[root] != white {
			continue
		}
		color[root], onPath[root] = gray, 0
		stack, pathNodes = append(stack, dfsFrame{node: root}), append(pathNodes, root)
		for steps := 0; len(stack) > 0 && steps < maxSteps; steps++ {
			top := &stack[len(stack)-1]
			edges := g.nodes[top.node].edges
			if top.next >= len(edges) {
				color[top.node] = black
				stack, pathNodes = stack[:len(stack)-1], pathNodes[:len(pathNodes)-1]
				continue
			}
			next := edges[top.next]
			top.next++
			switch color[next] {
			case gray:
				cycle := append(append([]int{}, pathNodes[onPath[next]:]...), next)
				cycles = append(cycles, cycle)
			case white:
				color[next], onPath[next] = gray, len(pathNodes)
				stack, pathNodes = append(stack, dfsFrame{node: next}), append(pathNodes, next)
			}
		}
		assert.That(len(stack) == 0, "the search finishes each root within its step bound")
	}
	assert.That(cap(stack) == n, "the stack keeps its fixed capacity")
	return cycles
}

// cycleFindings reports each cycle at the function of its first node.
func cycleFindings(cr *compiledRule, g *callGraph) []Finding {
	var out []Finding
	for _, cycle := range findCycles(g) {
		assert.That(len(cycle) >= 2, "a cycle has a head and a back edge")
		head := g.nodes[cycle[0]]
		fn := head.file.Functions[head.fn]
		if cr.suppressed(head.file.Lines, fn.Pos.Line) {
			continue
		}
		names := make([]string, len(cycle))
		for i, id := range cycle {
			_, key, found := strings.Cut(g.nodes[id].key, "\x00")
			assert.That(found, "a node key holds a scope and a name")
			names[i] = key
		}
		out = append(out, Finding{
			RuleID: cr.rule.ID, Severity: cr.rule.Severity, Path: head.file.Path,
			Message: "recursion: " + strings.Join(names, " -> "),
			Line:    fn.Pos.Line, Col: fn.Pos.Col, EndLine: fn.Pos.Line, EndCol: fn.Pos.Col,
		})
	}
	assert.That(len(out) <= g.edges, "one finding for each back edge at most")
	return out
}
