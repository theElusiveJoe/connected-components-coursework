package fastSVmpi

import (
	"fmt"
)

type slaveNode struct {
	distribution map[uint32]uint32
	rank         int
	f            map[uint32]uint32
	foreignf     map[uint32]uint32
	edges1       []uint32
	edges2       []uint32
	edgesNum     int
	changed      bool
}

func (slave *slaveNode) init(rank int) {
	slave.rank = rank
	slave.distribution = make(map[uint32]uint32)
	slave.f = make(map[uint32]uint32)
	slave.foreignf = make(map[uint32]uint32)
}

func (slave *slaveNode) addEdge(v1 uint32, v2 uint32, v2er uint32) {
	slave.edgesNum++
	slave.edges1 = append(slave.edges1, v1)
	slave.edges2 = append(slave.edges2, v2)
	slave.f[v1] = v1
	if v2er != uint32(slave.rank) {
		slave.distribution[v2] = v2er
		slave.foreignf[v2] = v2
	} else {
		slave.f[v2] = v2
	}
	// fmt.Printf("Im %d: added (%d, %d) 2nd hosted on %d\n", slave.rank, v1, v2, v2er)
}

func (slave *slaveNode) print() {
	fmt.Println(
		fmt.Sprintf("SLAVE %d {\n", slave.rank),
		"  edges:", slave.edgesNum, "\n",
		" ", slave.edges1, "\n",
		" ", slave.edges2, "\n",
		"  parents:", slave.f, "\n",
		"  distrib:", slave.distribution, "\n",
		"  fparents:", slave.foreignf, "\n}",
	)
}

func (slave *slaveNode) getEdgeAndServers(i int) (uint32, uint32, uint32, uint32) {
	return slave.edges1[i], slave.edges2[i], slave.distribution[slave.edges1[i]], slave.distribution[slave.edges2[i]]
}

func (slave *slaveNode) getEdge(i int) (uint32, uint32) {
	return slave.edges1[i], slave.edges2[i]
}

func (slave *slaveNode) getServer(v uint32) uint32 {
	return slave.distribution[v]
}

func (slave *slaveNode) getParent(v uint32) uint32 {
	return slave.f[v]
}

func (slave *slaveNode) getForeignParent(v uint32) uint32 {
	return slave.foreignf[v]
}

func (slave *slaveNode) setParent(v uint32, newParent uint32) {
	slave.f[v] = newParent
	slave.setChangedTrue()
}

func (slave *slaveNode) setForeignParent(v uint32, newParent uint32) {
	slave.foreignf[v] = newParent
}

func (slave *slaveNode) isServerOf(v uint32) bool {
	_, ok := slave.distribution[v]
	if !ok {
		return true
	}
	return false
}

func (slave *slaveNode) setChangedTrue() {
	slave.changed = true
}

func (slave *slaveNode) setChangedFalse() {
	slave.changed = false
}

func (slave *slaveNode) wasChanged() bool {
	return slave.changed
}
