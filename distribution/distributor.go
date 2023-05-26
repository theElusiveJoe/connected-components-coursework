package distribution

import "connectedComponents/graph"

type multiEdge struct {
	v1 uint32
	v2 uint32
}

type Distributor struct {
	multiEdges  map[multiEdge]struct{}
	hashNum     uint32
	nodesWeight []uint32
}

func (dist *Distributor) H(node uint32) uint32 {
	return node % dist.hashNum
}

func (dist *Distributor) addEdge(v1 uint32, v2 uint32) {
	h1, h2 := dist.H(v1), dist.H(v2)
	dist.nodesWeight[h1]++
	dist.nodesWeight[h2]++
	if h2 < h1 {
		h1, h2 = h2, h1
	} else if h1 == h2 {
		return
	}
	newMEdge1 := multiEdge{h1, h2}
	if _, ok := dist.multiEdges[newMEdge1]; ok {
		return
	}
	dist.multiEdges[newMEdge1] = struct{}{}
}

func (dist *Distributor) toGraph() *graph.Graph {
	edges1, edges2 := make([]uint32, len(dist.multiEdges)), make([]uint32, len(dist.multiEdges))
	i := uint32(0)
	for multiEdge := range dist.multiEdges {
		edges1[i], edges2[i] = multiEdge.v1, multiEdge.v2
		i++
	}
	
	return &graph.Graph{
		NodesNum: dist.hashNum,
		Edges1: edges1,
		Edges2: edges2,
		Mapa: map[string]uint32{},
	}
}
