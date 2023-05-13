package fastSVmpi

import (
	"connectedComponents/utils"
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

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

func (master *masterNode) getEdge(i uint32) (uint32, uint32) {
	return master.edges1[i], master.edges2[i]
}

func (master *masterNode) whoServes(v uint32) uint32 {
	return master.distribution[v]
}

func (master *masterNode) print() {
	fmt.Println(
		"Im MASTER{\n",
		" ", master.edges1, "\n",
		" ", master.edges2, "\n",
		"  distrib:", master.distribution, "\n}",
	)
}

func (master *masterNode) bcastTag(tag C.int) {
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, master.slavesNum+1)
}

func (master *masterNode) delegateEdge(i uint32) {
	a, b := master.getEdge(i)
	aer, ber := master.whoServes(a), master.whoServes(b)
	arr1 := []uint32{a, b, ber}
	mpiSendUintArray(arr1, int(aer), TAG_SEND_V1_V2_V2ER)
	if aer != ber {
		arr2 := []uint32{b, a, aer}
		mpiSendUintArray(arr2, int(ber), TAG_SEND_V1_V2_V2ER)
	}
}

func (master *masterNode) delegateAllEdges() {
	for i := uint32(0); i < (master.edgesNum); i++ {
		master.delegateEdge(i)
	}
}

func (master *masterNode) manageCCSearch() bool {
	var slavesFiishedPPNum int
	for slavesFiishedPPNum < master.slavesNum {
		mpiSkipIncoming(TAG_FINISHED_PARENT_PROPOSITION)
		slavesFiishedPPNum++
	}
	master.bcastTag(TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION)
	changed := false

	for i := 0; i < master.slavesNum; i++ {
		ch, _ := mpiRecvBool(TAG_SLAVE_WAS_CHANGED)
		changed = changed || ch
	}

	mpiBcastBoolViaSend(changed, TAG_SHALL_WE_CONTINUE, 1, master.slavesNum+1)

	return changed
}

func (master *masterNode) collectResult() map[uint32]uint32 {
	res := make(map[uint32]uint32)
	master.bcastTag(TAG_SEND_ME_RESULT)
	for i := 0; i < int(master.nodesNum); i++ {
		arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_I_SEND_RESULT)
		x, xParent := C.getArray(arr, 0), C.getArray(arr, 1)
		C.freeArray(arr)
		res[uint32(x)] = uint32(xParent)
	}
	return res
}
