package basic

import (
	"connectedComponents/graph"
	"connectedComponents/utils/ioEdges"
)

// максимально базовый алгоритм
func basicCCSearch(nodesNum uint32, edges1 []uint32, edges2 []uint32) []uint32 {
	f := make([]uint32, nodesNum)
	for i := 0; i < len(f); i++ {
		f[i] = uint32(i)
	}

	changed := true
	for changed {
		changed = false

		for i := 0; i < len(edges1); i++ {
			u, v := edges1[i], edges2[i]
			pu, pv := f[u], f[v]
			if pu < pv {
				f[v] = pu
				changed = true
			} else if pu > pv {
				f[u] = pv
				changed = true
			}
		}
	}

	return f
}

func BasicCCSearchFromFile(filename string) []uint32 {
	g := ioEdges.GetGraph(filename)
	return basicCCSearch(g.NodesNum, g.Edges1, g.Edges2)
}

func BasicCCSearch(g *graph.Graph) []uint32 {
	return basicCCSearch(g.NodesNum, g.Edges1, g.Edges2)
}