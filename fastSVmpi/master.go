package fastSVmpi

import (
	"connectedComponents/utils"
	"fmt"
)

type masterNode struct {
	edgesNum     uint32
	edges1       []uint32
	edges2       []uint32
	distribution []uint32
	slavesNum    int
	nodesNum     uint32
}

func (master *masterNode) init(filename string, slavesNum int) {
	nodesNum, _, _, edges1, edges2 := utils.GetEdges(filename)
	master.nodesNum = (uint32)(nodesNum)
	master.edgesNum = (uint32)(len(edges1))
	master.edges1, master.edges2 = edges1, edges2
	master.distribution = make([]uint32, master.nodesNum)
	master.slavesNum = slavesNum

	for i := 0; i < int(master.nodesNum); i++ {
		var d uint32
		if i < 4 {
			d = uint32(1)
		} else {
			d = (uint32)(i%(master.slavesNum) + 1)
		}
		master.distribution[i] = d // (uint32)(i%(master.slavesNum) + 1) //uint8(rand.Intn(worldSize-1) + 1)
	}
}

func (master *masterNode) getEdge(i int) (uint32, uint32) {
	return master.edges1[i], master.edges2[i]
}

func (master *masterNode) whoServes(v uint32) uint32 {
	return master.distribution[v]
}

func (master *masterNode) print() {
	fmt.Println(
		fmt.Sprintf("Im MASTER{\n"),
		" ", master.edges1, "\n",
		" ", master.edges2, "\n",
		"  distrib:", master.distribution, "\n}",
	)
}
