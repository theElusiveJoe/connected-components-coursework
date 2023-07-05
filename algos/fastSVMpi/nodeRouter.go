package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import "fmt"

type routerNode struct {
	hashToSlave map[uint32]uint32
}

func (router *routerNode) init() {
	router.hashToSlave = make(map[uint32]uint32)
}

func (router *routerNode) addRecord(hash uint32, slave uint32) {
	router.hashToSlave[hash] = slave
}

func (router *routerNode) getSlaveRank(tr *transRole, v uint32) int {
	hash := v % tr.hashNum
	return int(router.hashToSlave[hash])
}

func (router *routerNode) String() string {
	return fmt.Sprintf("%v", router.hashToSlave)
}
