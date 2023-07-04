package graph

type Graph struct {
	NodesNum uint32
	Edges1   []uint32
	Edges2   []uint32
	Mapa     map[string]uint32
}

func (g *Graph) GetEdge(i uint32) (uint32, uint32) {
	return g.Edges1[i], g.Edges2[i]
}

func (g *Graph) Len() uint32 {
	return uint32(len(g.Edges1))
}

func (g *Graph) ToIterator() *GraphIterator {
	var gi GraphIterator
	gi.Init(g)
	return &gi
}
