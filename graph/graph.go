package graph

type Graph struct {
	NodesNum uint32
	Edges1 []uint32
	Edges2 []uint32
	Mapa map[string]uint32
}

func (g *Graph) GetEdge(i uint32) (uint32, uint32) {
	return g.Edges1[i], g.Edges2[i]
}