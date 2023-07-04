package graph

type GraphIterator struct {
	G   *Graph
	len uint32
	i   uint32
}

func (iterator *GraphIterator) Init(g *Graph) {
	iterator.G = g
	iterator.len = g.Len()
}

func (iterator *GraphIterator) StartIter() {
	iterator.i = 0
}

func (iterator *GraphIterator) HasEdges() bool {
	return iterator.i != iterator.len
}

func (iterator *GraphIterator) GetNextEdge() (uint32, uint32) {
	iterator.i++
	return iterator.G.GetEdge(iterator.i - 1)
}

func (iterator *GraphIterator) Len() uint32 {
	return iterator.G.Len()
}

func (iterator *GraphIterator) NodesNum() uint32 {
	return iterator.G.NodesNum
}
