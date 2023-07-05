package fastSVMpi

// "strconv"

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
	hashNum      uint32
}

func (master *masterNode) init() {}
