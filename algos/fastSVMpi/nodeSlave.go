package fastSVMpi

import (
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

type slaveNode struct {
	rank     int
	f        map[uint32]uint32
	edges1   []uint32
	edges2   []uint32
	edgesNum int
	changed  bool
}

func (slave *slaveNode) init(rank int) {
	slave.rank = rank
	slave.f = make(map[uint32]uint32)
	slave.changed = true
}

func (slave *slaveNode) print() {
	fmt.Println(
		fmt.Sprintf("SLAVE %d {\n", slave.rank),
		"  edges:", slave.edgesNum, "\n",
		" ", slave.edges1, "\n",
		" ", slave.edges2, "\n",
		"  parents:", slave.f,
	)
}

func (slave *slaveNode) addEdge(v1 uint32, v2 uint32) {
	slave.edgesNum++
	slave.edges1 = append(slave.edges1, v1)
	slave.edges2 = append(slave.edges2, v2)
}
func (slave *slaveNode) getEdge(i int) (uint32, uint32) {
	return slave.edges1[i], slave.edges2[i]
}

func (slave *slaveNode) getParent(v uint32) uint32 {
	return slave.f[v]
}
func (slave *slaveNode) setParentIfLess(v uint32, newParent uint32) {
	if slave.f[v] > newParent {
		slave.setParent(v, newParent)
	}
}
func (slave *slaveNode) setParent(v uint32, newParent uint32) {
	slave.f[v] = newParent
	slave.setChangedTrue()
}

func (slave *slaveNode) setChangedTrue()  { slave.changed = true }
func (slave *slaveNode) setChangedFalse() { slave.changed = false }
func (slave *slaveNode) wasChanged() bool { return slave.changed }

func (slave *slaveNode) isServerOf(v uint32) bool {
	_, ok := slave.f[v]
	return ok
}

// func (slave *slaveNode) findRouter(v uint32, tr *transRole) uint32 {
// 	return
// }
