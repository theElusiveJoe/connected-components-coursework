package basicMpi

import (
	"bytes"
	"connectedComponents/src/algos"
	"connectedComponents/src/graph"
	"encoding/gob"
	"fmt"
	"os"
	"path"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

type masterNode struct {
	edgesNum uint32
	edges1   []uint32
	edges2   []uint32
	// distribution []uint32
	slavesNum int
	nodesNum  uint32
	hashNum   uint32
}

func (master *masterNode) init(nodesNum uint32, edges1 []uint32, edges2 []uint32, slavesNum int, hashNum int) {
	master.nodesNum = (uint32)(nodesNum)
	master.edgesNum = (uint32)(len(edges1))
	master.edges1, master.edges2 = edges1, edges2
	// master.distribution = make([]uint32, master.nodesNum)
	master.slavesNum = slavesNum
	master.hashNum = uint32(hashNum)

	g := graph.Graph{
		NodesNum: master.nodesNum,
		Edges1:   master.edges1,
		Edges2:   master.edges2,
		Mapa:     map[string]uint32{}, //mapa,
	}
	var iter graph.GraphIterator
	iter.Init(&g)

	fmt.Println("ДЛЯ поиска распределения используется граф:")

	master.print()
}

func (master *masterNode) getEdge(i uint32) (uint32, uint32) {
	return master.edges1[i], master.edges2[i]
}

func (master *masterNode) whoServes(v uint32) uint32 {
	// return master.distribution[v%master.hashNum]
	return v%uint32(master.slavesNum) + 1
}

func (master *masterNode) print() {
	fmt.Println(
		"Im MASTER{\n",
		" ", master.edges1, "\n",
		" ", master.edges2, "\n",
		"\n}",
	)
}

func (master *masterNode) bcastTag(tag C.int) {
	if tag == TAG_NEXT_PHASE {
		fmt.Println("->-> MASTER BROADCASTS NEXT PHASE")
	}
	mpiBcastTagViaSend(tag, 1, master.slavesNum+1)
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

func (master *masterNode) manageExpectedPPNCounting() {
	for i := 0; i < master.slavesNum; i++ {
		// fmt.Println("-> MASTER: slave counted PPN")
		mpiSkipIncoming(TAG_ALL_MY_MESSAGES_REACHED_TARGET)
	}
	master.bcastTag(TAG_NEXT_PHASE)
}

func (master *masterNode) manageCCSearch() bool {
	changed := false
	for i := 0; i < master.slavesNum; i++ {
		ch, _ := mpiRecvBool(TAG_SLAVE_WAS_CHANGED)
		// fmt.Println("slave wants to continue:", ch)
		changed = changed || ch
	}

	mpiBcastBoolViaSend(changed, TAG_SHALL_WE_CONTINUE, 1, master.slavesNum+1)

	return changed
}

func (master *masterNode) collectResult() []uint32 {
	res := make([]uint32, master.nodesNum)
	master.bcastTag(TAG_SEND_ME_RESULT)

	for i := 0; i < int(master.slavesNum); i++ {
		fmt.Println("COLLECT FROM NODE", i+1)
		arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_I_SEND_RESULT)
		x, xParent := C.getArray(arr, 0), C.getArray(arr, 1)
		C.freeArray(arr)
		res[uint32(x)] = uint32(xParent)
	}
	return res
}

func (master *masterNode) collectResultToTable(conf *algos.RunConfig) {
	master.bcastTag(TAG_SEND_ME_RESULT)
	total_res := make(map[uint32]uint32)
	for i := 0; i < int(master.slavesNum); i++ {
		arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_I_SEND_RESULT)
		x, xParent := uint32(C.getArray(arr, 0)), uint32(C.getArray(arr, 1))
		C.freeArray(arr)
		total_res[x] = xParent
	}

	name := conf.Id + "_" + fmt.Sprintf("%d", 1) + ".mapbin"
	p := path.Join(conf.ResultDir, name)
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(total_res); err != nil {
		panic("result encoding failed")
	}
	os.Create(p)
	if err := os.WriteFile(p, b.Bytes(), 0777); err != nil {
		panic(err)
	}
}
