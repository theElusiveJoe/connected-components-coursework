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
	TAG_RESPONSIBLE_FOR_NODE C.int = iota
	TAG_SEND_V1_V2_V2ER

	TAG_V1_PARENT_PROPOSITION
	TAG_FINISHED_PARENT_PROPOSITION
	TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION

	TAG_NEXT_PHASE
	TAG_
)

// чтобы не путать "вычислительные узлы" и "узлы графа",
// первых буду называть мастерами (masters) и слейвами (slaves),
// а вторых узлами (nodes) или вершинами (verticies).
func RunAlgo(filename string) {
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
	{
		if rank == MASTER {
			for i := 0; (uint32)(i) < (master.edgesNum); i++ {
				a, b := master.getEdge(i)
				aer, ber := master.whoServes(a), master.whoServes(b)
				arr1 := []uint32{a, b, ber}
				mpiSendUintArray(arr1, int(aer), TAG_SEND_V1_V2_V2ER)
				if aer != ber {
					arr2 := []uint32{b, a, aer}
					mpiSendUintArray(arr2, int(ber), TAG_SEND_V1_V2_V2ER)
				}
			}
			mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, worldSize)

		} else {
			for {
				arr, status := mpiRecvUintArray(3, MASTER, C.MPI_ANY_TAG)

				if status.MPI_TAG == TAG_NEXT_PHASE {
					C.freeArray(arr)
					break
				}

				v1, v2, v2er := cGetArr(arr, 0), cGetArr(arr, 1), cGetArr(arr, 2)
				C.freeArray(arr)
				slave.addEdge(v1, v2, v2er)
			}
			slave.print()
		}
		mpiBarrier(C.MPI_COMM_WORLD)
		if rank == MASTER {
			fmt.Println("------------------\nALL INITED\n------------------")
		}
		mpiBarrier(C.MPI_COMM_WORLD)
	}
	// 2 CC COMPUTING
	{
		for {
			var changed bool
			if rank == MASTER {
				var slavesFiishedPPNum int
				for slavesFiishedPPNum < master.slavesNum {
					b, _ := mpiRecvBool(TAG_FINISHED_PARENT_PROPOSITION)
					if b {
						changed = true
					}
					fmt.Println("MASTER: RECVD PROPOSITION REPORT")
					// mpiSkipIncoming(TAG_FINISHED_PARENT_PROPOSITION)
					slavesFiishedPPNum++
				}
				fmt.Println("MASTER: ALL sLAVE FINIDHED PARENT PROPOSITION")
				// mpiBcastTagViaSend(TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION, 1, worldSize)
				fmt.Println("MASTER BROADCASTS:", changed)
				mpiBcastBoolViaSend(changed, TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION, 1, worldSize)
			} else {
				slave.setChangedFalse()
				// 2.1 слейвы работают каждый внутри себя
				for i := 0; i < slave.edgesNum; i++ {
					v1, v2 := slave.getEdge(i)
					v1parent := slave.getParent(v1)

					// если это внутренний узел
					if slave.isServerOf(v2) {
						v2parent := slave.getParent(v2)
						if v1parent > v2parent {
							slave.setParent(v1, v2parent)
						} else if v1parent < v2parent {
							slave.setParent(v2, v1parent)
						}
					} else {
						v2parent := slave.getForeignParent(v2)
						if v1parent > v2parent {
							slave.setParent(v1, v2parent)
						} else if v1parent < v2parent {
							slave.setForeignParent(v2, v1parent)
						}
					}
				}
				fmt.Println("->-> slave", slave.rank, "ENDED INNER TRANSFORMS")
				mpiBarrier(SLAVES_COMM)
				slave.print()
				mpiBarrier(SLAVES_COMM)

				// 2.2 слейвы обмениваются информацией (принятие приоритетнее)
				fmt.Println("->->-> slave", slave.rank, "START TO PROPOSE")
				mpiBarrier(SLAVES_COMM)
				for foreignNode, proposedParent := range slave.foreignf {
					for mpiCheckIncoming(TAG_V1_PARENT_PROPOSITION) {
						arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_V1_PARENT_PROPOSITION)
						v1, parentProposition := cGetArr(arr, 0), cGetArr(arr, 1)
						// fmt.Println("->->-> slave", slave.rank, "THERE IS INCOMING PROPOSAL", v1, parentProposition)
						C.freeArray(arr)
						if slave.f[v1] > parentProposition {
							slave.f[v1] = parentProposition
						}
					}
					foreignNoder := slave.getServer(foreignNode)

					mpiSendUintArray([]uint32{foreignNode, proposedParent}, int(foreignNoder),
						TAG_V1_PARENT_PROPOSITION)
					fmt.Println("->->-> slave", slave.rank,
						"I SENT PROPOSAL TO SLAVE", foreignNoder,
						"TO SET", proposedParent, "AS A PARENT TO NODE", foreignNode)
				}

				fmt.Println("->->!! slave", slave.rank, "ENDED OUTER PROPOSITIONS")
				mpiBarrier(SLAVES_COMM)

				// слейв отчитывается о том, что закончил предлагать
				// родителей другим слейвам,
				// но продолжает слушать другие предлоение или команду мастера
				// mpiReportToMaster(TAG_FINISHED_PARENT_PROPOSITION)
				mpiSendBool(slave.wasChanged(), MASTER, TAG_FINISHED_PARENT_PROPOSITION)
				fmt.Println("->->->-> slave", slave.rank, "REPORTED TO MASTER")

				for {
					if mpiCheckIncoming(TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION) {
						changed, _ = mpiRecvBool(TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION)
						fmt.Println("SLAVE", rank, "RECVD", changed)
						break
					}
					if mpiCheckIncoming(TAG_V1_PARENT_PROPOSITION) {
						arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_V1_PARENT_PROPOSITION)
						v1, parentProposition := cGetArr(arr, 0), cGetArr(arr, 1)
						C.freeArray(arr)
						if slave.f[v1] > parentProposition {
							slave.f[v1] = parentProposition
						}
					}
				}

				mpiBarrier(SLAVES_COMM)
				fmt.Println(slave.rank, "ENDED ITER")
				mpiBarrier(SLAVES_COMM)
			}
			mpiBarrier(C.MPI_COMM_WORLD)
			// break
			if !changed {
				break
			}
		}
		mpiBarrier(C.MPI_COMM_WORLD)
	}

	if rank == MASTER {
		fmt.Println("------------------\nALL STOPPED\n------------------")
	} else {
		slave.print()
	}

	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
}
