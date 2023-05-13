package fastSVmpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	"fmt"
)

const MASTER = 0

const (
	TAG_NEXT_PHASE C.int = iota

	TAG_RESPONSIBLE_FOR_NODE
	TAG_SEND_V1_V2_V2ER

	TAG_V1_PARENT_PROPOSITION
	TAG_FINISHED_PARENT_PROPOSITION

	TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION
	TAG_SLAVE_WAS_CHANGED
	TAG_SHALL_WE_CONTINUE

	TAG_SEND_ME_RESULT
	TAG_I_SEND_RESULT
	TAG_I_SENT_EVERYTING
)

// чтобы не путать "вычислительные узлы" и "узлы графа",
// первых буду называть мастерами (masters) и слейвами (slaves),
// а вторых узлами (nodes) или вершинами (verticies).
func RunAlgo(filename string) map[uint32]uint32 {
	// 0 INIT WORLD
	C.MPI_Init(nil, nil)
	var rank, worldSize int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intPtr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intPtr(&worldSize))

	var master masterNode
	var slave slaveNode

	if rank == MASTER {
		master.init(filename, worldSize-1)
		master.print()
	} else {
		slave.init(rank)
	}

	// настраивае коммуникатор для слейвов
	var SLAVES_COMM C.MPI_Comm
	var color C.int
	if rank == MASTER {
		color = C.MPI_UNDEFINED
	} else {
		color = 1
	}
	C.MPI_Comm_split(
		C.MPI_COMM_WORLD,
		color,
		(C.int)(rank),
		&SLAVES_COMM,
	)

	// 1 MASTER DISTRIBUTES NODES
	if rank == MASTER {
		master.delegateAllEdges()
		master.bcastTag(TAG_NEXT_PHASE)
	} else {
		slave.recvEdgesTillGetBreakTag(TAG_NEXT_PHASE)
		slave.print()
	}

	mpiBarrier(C.MPI_COMM_WORLD)
	if rank == MASTER {
		fmt.Println("------------------\nALL INITED\n------------------")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	// 2 CC COMPUTING

	if rank == MASTER {
		for master.manageCCSearch() {
		}
	} else {
		slave.runInnerHooking()
		cont := slave.runParentProposals()
		for cont {
			slave.runInnerHooking()
			cont = slave.runParentProposals()
		}
	}

	mpiBarrier(C.MPI_COMM_WORLD)

	if rank == MASTER {
		fmt.Println("------------------\nALL STOPPED\n------------------")
	} else {
		slave.print()
	}

	var result map[uint32]uint32
	if rank == MASTER {
		result = master.collectResult()
	} else {
		slave.sendResult()
	}

	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
	return result
}
