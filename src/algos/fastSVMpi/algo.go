package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	"connectedComponents/src/algos"
	"fmt"
)

const (
	TAG_NEXT_PHASE C.int = iota

	// STEP 1 DISTRIB
	TAG_DISTRIBUTED_EDGE
	TAG_INNER_EDGE
	TAG_GOT_EDGE
	TAG_HASH_TO_SLAVE_ROW
	TAG_GOT_ROW

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

func (tr *transRole) String() string {
	var rs string
	if tr.role == 0 {
		rs = "MASTER"
	} else if tr.role == 1 {
		rs = "ROUTER"
	} else {
		rs = "SLAVE"
	}

	s := ""
	s += "tr {\n"
	s += fmt.Sprintf("    rank: %d\n", tr.rank)
	s += fmt.Sprintf("    role: %s\n", rs)
	// s += fmt.Sprintf("    hashnum: %d\n", tr.hashNum)
	// s += fmt.Sprintf("    slaves: %d\n", tr.slavesNum)
	// s += fmt.Sprintf("    routers: %d\n", tr.routersNum)

	if tr.role == 0 {

	} else if tr.role == 1 {
		s += "    me: " + tr.router.String() + "\n"
	} else {
		s += "    me: " + tr.slave.String() + "\n"
	}

	s += "}"
	return s
}

func (tr *transRole) talk(format string, args ...any) {
	// var label string
	// if tr.role == 0 {
	// 	label = "MASTER"
	// } else if tr.role == 1 {
	// 	label = fmt.Sprintf("ROUTER %d", tr.rank)
	// } else {
	// 	label = fmt.Sprintf("SLAVE %d", tr.rank)
	// }

	// fmt.Print(
	// 	fmt.Sprintf("-> {%s}: ", label),
	// 	fmt.Sprintf(format, args...),
	// 	"\n",
	// )
}

func (tr *transRole) findRouter(v uint32) int {
	h := int(v % tr.hashNum)
	return h%tr.routersNum + 1
}

func Run(conf *algos.RunConfig) {
	tr := runStep0(conf.TestFile, conf.RoutersNum, uint32(conf.HashNum))

	runStep1Distrib(tr)

	mpiBarrier(C.MPI_COMM_WORLD)
	fmt.Println(tr)
	mpiBarrier(C.MPI_COMM_WORLD)
	iters := 0

	for {
		iters++
		if tr.role == MASTER {
			fmt.Println("-> iteration", iters)
		}
		runStep2Stochastic(tr)
		runStep3Aggressive(tr)
		runStep4ShortCutting(tr)
		if !runStep5SlavePolling(tr) {
			break
		}
	}
	runStep6SaveResult(tr, conf)

	mpiBarrier(C.MPI_COMM_WORLD)

	mpiBarrier(C.MPI_COMM_WORLD)
	if tr.role == SLAVE {
		fmt.Println(tr.slave.f)
	}
	mpiBarrier(C.MPI_COMM_WORLD)
	if tr.role == MASTER {
		fmt.Printf("\n----------- ENDED in %d iters -----------\n\n", iters)
	}
	C.MPI_Finalize()
}
