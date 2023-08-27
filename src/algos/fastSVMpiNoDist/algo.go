package fastSVMpiNoDist

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	"connectedComponents/src/algos"
	"fmt"
	"os"
	"time"
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
	SLAVE
)

type transRole struct {
	slave *slaveNode
	rank  int
	role  int

	filename string

	worldSize int
	slavesNum int

	curIter int
	curStep string

	logstring string

	SLAVES_COMM C.MPI_Comm
}

func (tr *transRole) String() string {
	var rs string
	if tr.role == MASTER {
		rs = "MASTER"
	} else {
		rs = "SLAVE"
	}

	s := ""
	s += "tr {\n"
	s += fmt.Sprintf("    rank: %d\n", tr.rank)
	s += fmt.Sprintf("    role: %s\n", rs)

	if tr.role == MASTER {

	} else {
		s += "    me: " + tr.slave.String() + "\n"
	}

	s += "}"
	return s
}

func (tr *transRole) talk(format string, args ...any) {
	// return
	var label string
	if tr.role == 0 {
		label = "MASTER"
	} else {
		label = fmt.Sprintf("SLAVE %d", tr.rank)
	}

	fmt.Print(
		fmt.Sprintf("-> {%s}: ", label),
		fmt.Sprintf(format, args...),
		"\n",
	)
}

func (tr *transRole) log(format string, args ...any) {
	msg := fmt.Sprintf("%d ", time.Now().UnixMilli())
	msg += fmt.Sprintf("%d ", tr.rank)
	msg += fmt.Sprintf("%d ", tr.curIter)
	msg += fmt.Sprintf("%s ", tr.curStep)
	msg += fmt.Sprintf(format, args...)
	msg += "\n"
	s := []byte(msg)

	file, err := os.OpenFile("log.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprintf(file, "%s", s)
}

func (tr *transRole) getServer(v uint32) int {
	return int(v%uint32(tr.slavesNum)) + 1
}

func Run(conf *algos.RunConfig) {
	tr := runStep0(conf.TestFile, uint32(conf.HashNum))

	runStep1Distrib(tr)

	mpiBarrier(C.MPI_COMM_WORLD)
	// fmt.Println(tr)
	mpiBarrier(C.MPI_COMM_WORLD)
	tr.curIter = 0

	for {
		tr.curIter++
		if tr.role == MASTER {
			fmt.Println("-> iteration", tr.curIter)
		}
		tr.curStep = "2"
		runStep2Stochastic(tr)
		tr.curStep = "3"
		runStep3Aggressive(tr)
		tr.curStep = "4"
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
		fmt.Printf("\n----------- ENDED in %d iters -----------\n\n", tr.curIter)
	}
	C.MPI_Finalize()
}
