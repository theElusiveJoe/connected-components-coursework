package utils

import (
	"connectedComponents/graph"
	"connectedComponents/utils/ioEdges"
)

type EdgesIteratorInMemory struct {
	g   *graph.Graph
	len uint32
	i   uint32
}

func (iterator *EdgesIteratorInMemory) init(filename string) {
	iterator.g = ioEdges.GetGraph(filename)
	iterator.len = uint32(len(iterator.g.Edges1))
}

func (iterator *EdgesIteratorInMemory) hasEdges() bool {
	return iterator.i != iterator.len
}

func (iterator *EdgesIteratorInMemory) nextEdge() (uint32, uint32) {
	iterator.i++
	return iterator.g.GetEdge(iterator.i - 1)
}
