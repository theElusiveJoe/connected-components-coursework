package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	_ "fmt"
)

const (
	TAG_NEXT_PHASE C.int = iota

	// STEP 1 DISTRIB
	TAG_RESPONSIBLE_FOR_V1_CONNECTED_TO_V2
	TAG_HASH_TO_SLAVE_ROW

	// STEP 2 STOCHASTIC HOOKING
	TAG_SH1
	TAG_SH2
	TAG_SH3
	TAG_SH4
	TAG_SH5
	TAG_SH6
	TAG_SH7
	TAG_SH_ALL_CONFIRMATIONS_RECIEVED

	// STEP 3 AGGRESSIVE HOOKING
	TAG_AH1
	TAG_AH2
	TAG_AH3
	TAG_AH4
	TAG_AH5
	TAG_AH_ALL_CONFIRMATIONS_RECIEVED

	// STEP 4 SHORTCUTTING
	TAG_SC1
	TAG_SC2
	TAG_SC3
	TAG_SC_ALL_CONFIRMATIONS_RECIEVED

	// STEP 5 SLAVEPOLLING
	TAG_SP1
	TAG_SP2
)

const (
	MASTER int = iota
	ROUTER
	SLAVE
)

type transRole struct {
	master *masterNode
	router *routerNode
	slave  *slaveNode
	rank   int
	role   int

	filename string
	hashNum  uint32

	worldSize  int
	routersNum int
	slavesNum  int

	SLAVES_COMM C.MPI_Comm
}

func (tr *transRole) findRouter(v uint32) int {
	h := int(v % tr.hashNum)
	routerNum := h%tr.routersNum + 1
	return routerNum
}

func Run(filename string, routersNum int) {
	tr := runStep0(filename, routersNum)
	runStep1Distrib(tr)

	for {
		runStep2Stochastic(tr)
		runStep3Aggressive(tr)
		runStep4ShortCutting(tr)
		if !runStep5SlavePolling(tr) {
			break
		}
	}

}
