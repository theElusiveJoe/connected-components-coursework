package fastSVmpi

import (
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

type slaveNode struct {
	distribution               map[uint32]uint32
	rank                       int
	f                          map[uint32]uint32
	foreignf                   map[uint32]uint32
	edges1                     []uint32
	edges2                     []uint32
	edgesNum                   int
	changed                    bool
	expextedParentProposalsNum uint32
	recievedParentProposalsNum uint32
}

func (slave *slaveNode) init(rank int) {
	slave.rank = rank
	slave.distribution = make(map[uint32]uint32)
	slave.f = make(map[uint32]uint32)
	slave.foreignf = make(map[uint32]uint32)
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
		slave.expextedParentProposalsNum++
	}
	fmt.Printf("-> slave %d: got edge (%d, %d) with v2 on %d\n", slave.rank, v1, v2, v2er)
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
func (slave *slaveNode) setParent(v uint32, newParent uint32) {
	slave.f[v] = newParent
	slave.setChangedTrue()
}

func (slave *slaveNode) getForeignParent(v uint32) uint32 {
	return slave.foreignf[v]
}
func (slave *slaveNode) setForeignParent(v uint32, newParent uint32) {
	slave.foreignf[v] = newParent
}

func (slave *slaveNode) isServerOf(v uint32) bool {
	_, ok := slave.distribution[v]
	return !ok
}

func (slave *slaveNode) setChangedTrue()  { slave.changed = true }
func (slave *slaveNode) setChangedFalse() { slave.changed = false }
func (slave *slaveNode) wasChanged() bool { return slave.changed }

// нулевая стадия - принимаем новые ребра, пока не получим сигнал об остановке
func (slave *slaveNode) recvEdgesTillGetBreakTag(breakTag C.int) {
	recvEdgeOrBreak := func(slave *slaveNode, breakTag C.int) bool {
		arr, status := mpiRecvUintArray(3, MASTER, C.MPI_ANY_TAG)
		if status.MPI_TAG == breakTag {
			C.freeArray(arr)
			return false
		}
		if status.MPI_TAG != TAG_SEND_V1_V2_V2ER {
			panic(fmt.Sprintf("Expected TAG_SEND_V1_V2_V2ER got %d", status.MPI_TAG))
		}
		slave.addEdge(cGetArr(arr, 0), cGetArr(arr, 1), cGetArr(arr, 2))
		C.freeArray(arr)
		return true
	}

	for recvEdgeOrBreak(slave, breakTag) {
	}
}

// перевешиваем родителей внутри слейва, не обмениваясь информацией
func (slave *slaveNode) runInnerHooking() {
	innerHookingOfIthEdge := func(slave *slaveNode, i int) {
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

	slave.setChangedFalse()
	for i := 0; i < slave.edgesNum; i++ {
		innerHookingOfIthEdge(slave, i)
	}
	// fmt.Printf("-> slave %d: Ended inner hooking\n", slave.rank)
}

// обмениваемся результатами друг с другом
func (slave *slaveNode) runParentProposals() bool {
	checkIncomingParentProposal := func(slave *slaveNode) {
		// пока есть входящие предложения для foreign_parent,
		// принимаем их и записываем
		for mpiCheckIncoming(TAG_V1_PARENT_PROPOSITION) {
			slave.recievedParentProposalsNum++
			arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_V1_PARENT_PROPOSITION)
			v1, parentProposition := cGetArr(arr, 0), cGetArr(arr, 1)
			fmt.Printf("slave %d: should i set f[ %d ] = %d: %t\n", slave.rank, v1, parentProposition, slave.getParent(v1) > parentProposition)
			C.freeArray(arr)
			if slave.getParent(v1) > parentProposition {
				slave.setParent(v1, parentProposition)
			}
		}
	}
	sendParentProposition := func(slave *slaveNode, foreignNode uint32, proposedParent uint32) {
		// foreignNode - узлел графа A
		// proposedParent - узлел графа B. Мы предлагаем поставить B в качестве родителя узла A
		// foreignNoder - слейв, отвечающий за узел A. Туда мы и отправим запрос
		foreignNoder := slave.getServer(foreignNode)
		fmt.Println(slave.rank, "->", foreignNoder, "set f[", foreignNode, "] =", proposedParent)
		mpiSendUintArray([]uint32{foreignNode, proposedParent}, int(foreignNoder),
			TAG_V1_PARENT_PROPOSITION)
	}

	slave.recievedParentProposalsNum = 0
	for foreignNode, proposedParent := range slave.foreignf {
		checkIncomingParentProposal(slave)
		sendParentProposition(slave, foreignNode, proposedParent)
	}
	// fmt.Printf("-> slave %d: Ended parent proposals\n", slave.rank)

	// отчитываемся о том, что закончили шэрить результаты
	mpiReportToMaster(TAG_FINISHED_PARENT_PROPOSITION)
	for {
		// fmt.Printf("-> slave %d: Listening to incoming parent proposals\n", slave.rank)
		// пока не получим сигнал о том, что все закончили
		if mpiCheckIncoming(TAG_ALL_SLAVES_FINISHED_PARENT_PROPOSITION) {
			// fmt.Printf("-> slave %d: Ended proposition listening\n", slave.rank)
			break
		}
		// слущаем входящие предложения
		checkIncomingParentProposal(slave)
	}

	// все закончили добычу и обмен информацией
	// пора ли нам завершится?

	// отправляем своё мнение
	mpiSendBool(slave.changed, MASTER, TAG_SLAVE_WAS_CHANGED)
	// и ждем команды - продолжать или нет
	shallWeContinue, _ := mpiRecvBool(TAG_SHALL_WE_CONTINUE)
	return shallWeContinue
}

func (slave *slaveNode) sendResult() {
	mpiSkipIncoming(TAG_SEND_ME_RESULT)
	for x, xParent := range slave.f {
		mpiSendUintArray(
			[]uint32{x, xParent},
			MASTER,
			TAG_I_SEND_RESULT,
		)
	}
}
