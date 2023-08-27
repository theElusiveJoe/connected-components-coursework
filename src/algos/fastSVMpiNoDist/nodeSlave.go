package fastSVMpiNoDist

import (
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

type slaveNode struct {
	rank int

	f map[uint32]uint32

	edges1   []uint32
	edges2   []uint32
	edgesNum int

	changed bool
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
		// " ", slave.edges1, "\n",
		// " ", slave.edges2, "\n",
		"  parents:", slave.f,
	)
}

func (slave *slaveNode) addEdge(v1 uint32, v2 uint32, isInner bool) {
	// добавлем ребро
	slave.edgesNum++
	slave.edges1 = append(slave.edges1, v1)
	slave.edges2 = append(slave.edges2, v2)

	// сразу ставим минимальный parent
	var min uint32
	if v2 < v1 {
		min = v2
	} else {
		min = v1
	}
	slave.setParentIfLess(v1, min)

	// если это внутренне ребро, то обрабатываем его симметрично
	if isInner {
		slave.setParentIfLess(v2, min)
	}
}

func (slave *slaveNode) getEdge(i int) (uint32, uint32) {
	return slave.edges1[i], slave.edges2[i]
}

func (slave *slaveNode) getParent(v uint32) uint32 {
	if p, ok := slave.f[v]; !ok {
		panic(fmt.Sprintf("i dont manage node %d\n", v))
	} else {
		return p
	}
}
func (slave *slaveNode) setParentIfLess(v uint32, newParent uint32) {
	// если parent еще не определен
	if _, ok := slave.f[v]; !ok {
		slave.setParent(v, newParent)
		return
	}

	// иначе проверяем
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

func (slave *slaveNode) String() string {
	return fmt.Sprintf(
		"{\n"+
			"        edges:\n"+
			"            %v\n"+
			"            %v\n"+
			"        parents:\n"+
			"            %v\n"+
			"        changed: %v\n"+
			"    }\n",
		slave.edges1, slave.edges2, slave.f, slave.changed)
}
