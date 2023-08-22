package basicMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import (
	"connectedComponents/algos"
	"connectedComponents/utils/ioEdges"
	"fmt"
)

const MASTER = 0

const (
	TAG_NEXT_PHASE C.int = iota

	TAG_RESPONSIBLE_FOR_NODE
	TAG_SEND_V1_V2_V2ER

	TAG_UR_INNER_NODE_IS_MY_FOREIGN
	TAG_BRUH_I_ALREADY_KNEW
	TAG_ALL_MY_MESSAGES_REACHED_TARGET

	TAG_V1_PARENT_PROPOSITION
	TAG_FINISHED_PARENT_PROPOSITION

	TAG_SLAVE_WAS_CHANGED
	TAG_SHALL_WE_CONTINUE

	TAG_SEND_ME_RESULT
	TAG_I_SEND_RESULT
)

// чтобы не путать "вычислительные узлы" и "узлы графа",
// первых буду называть мастерами (masters) и слейвами (slaves),
// а вторых узлами (nodes) или вершинами (verticies).
func basicMpiCCSearch(nodesNum uint32, edges1 []uint32, edges2 []uint32, conf *algos.RunConfig) {
	// 0 INIT WORLD
	C.MPI_Init(nil, nil)
	var rank, worldSize int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intPtr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intPtr(&worldSize))
	fmt.Println("####RANK", rank)
	var master masterNode
	var slave slaveNode

	if rank == MASTER {
		master.init(nodesNum, edges1, edges2, worldSize-1, conf.HashNum)
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
		// slave.print()
	}

	mpiBarrier(C.MPI_COMM_WORLD)
	if rank == MASTER {
		fmt.Println("------------------\nSTART INITING\n------------------")
	}
	slave.print()
	mpiBarrier(C.MPI_COMM_WORLD)
	if rank == MASTER {
		fmt.Println("------------------\nALL INITED\n------------------")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	// 1.5 count Expected Parent Proposals Num
	if rank == MASTER {
		master.manageExpectedPPNCounting()
	} else {
		slave.countExpectedParentProposalsNum()
	}

	// 2 CC COMPUTING

	if rank == MASTER {
		for master.manageCCSearch() {
			fmt.Println("=>=> DONE AGAIN")
		}
	} else {
		slave.runInnerHooking()
		cont := slave.runParentProposals()
		for cont {
			mpiBarrier(SLAVES_COMM)
			fmt.Println("----------")
			mpiBarrier(SLAVES_COMM)
			slave.runInnerHooking()
			cont = slave.runParentProposals()
		}
	}

	mpiBarrier(C.MPI_COMM_WORLD)
	if rank == MASTER {
		fmt.Println("------------------\nALL STOPPED\n------------------")
	}
	mpiBarrier(C.MPI_COMM_WORLD)
	if rank != MASTER {
		slave.print()
	}
	mpiBarrier(C.MPI_COMM_WORLD)
	if rank == MASTER {
		fmt.Println("------------------\nALL STOPPED\n------------------")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	if rank == MASTER {
		master.collectResultToTable(conf)
	} else {
		slave.sendResult()
	}
	C.MPI_Finalize()
}

func Run(conf *algos.RunConfig) {
	g := ioEdges.LoadGraph(conf.TestFile)
	basicMpiCCSearch(g.NodesNum, g.Edges1, g.Edges2, conf)
}
