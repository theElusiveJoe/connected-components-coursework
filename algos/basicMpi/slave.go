package basicMpi

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
	expectedParentProposalsNum uint32
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
		fmt.Printf("INTERNODE EDGE BETWEEN %d and %d (%d-%d)\n", slave.rank, v2er, v1, v2)
		slave.distribution[v2] = v2er
		slave.foreignf[v2] = v2
	} else {
		slave.f[v2] = v2
	}
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

// первая стадия - принимаем новые ребра, пока не получим сигнал об остановке
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

// стадия 1.5 - слейвы договариваются о количестве предстоящих предложений родителей
func (slave *slaveNode) countExpectedParentProposalsNum() {
	// если слейв A отвечает за узлы {a1, a2, ... an}, n>=1
	// слейв B отвечает за узел b
	// и есть ребра {(a1,b), (a2,b) ... (an,b)}
	// то A пошлет на B один parent proposal

	slave.expectedParentProposalsNum = 0
	// в конце counter должен занулиться
	// таким образом мы удостоверимся, что все сообщения дошли
	counter := 0

	// посылаем сообщение для каждого foreign node`а
	for _, v2er := range slave.distribution {
		for mpiCheckIncoming(TAG_BRUH_I_ALREADY_KNEW) {
			mpiSkipIncoming(TAG_BRUH_I_ALREADY_KNEW)
			counter--
			if slave.rank == 1 {
				fmt.Println("HE KNew", counter)
			}
		}
		for mpiCheckIncoming(TAG_UR_INNER_NODE_IS_MY_FOREIGN) {
			mpiSkipIncomingAndResponce(TAG_UR_INNER_NODE_IS_MY_FOREIGN, TAG_BRUH_I_ALREADY_KNEW)
			slave.expectedParentProposalsNum++
		}
		mpiSendTag(TAG_UR_INNER_NODE_IS_MY_FOREIGN, int(v2er))
		counter++
	}

	for counter != 0 {
		for mpiCheckIncoming(TAG_BRUH_I_ALREADY_KNEW) {
			mpiSkipIncoming(TAG_BRUH_I_ALREADY_KNEW)
			counter--
			if slave.rank == 1 {
				fmt.Println("HE KNew", counter)
			}
		}
		for mpiCheckIncoming(TAG_UR_INNER_NODE_IS_MY_FOREIGN) {
			mpiSkipIncomingAndResponce(TAG_UR_INNER_NODE_IS_MY_FOREIGN, TAG_BRUH_I_ALREADY_KNEW)
			slave.expectedParentProposalsNum++
		}
	}
	mpiReportToMaster(TAG_ALL_MY_MESSAGES_REACHED_TARGET)
	for {
		for mpiCheckIncoming(TAG_UR_INNER_NODE_IS_MY_FOREIGN) {
			mpiSkipIncomingAndResponce(TAG_UR_INNER_NODE_IS_MY_FOREIGN, TAG_BRUH_I_ALREADY_KNEW)
			slave.expectedParentProposalsNum++
		}
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			fmt.Println("-> SLAVE: PPN next phase")
			return
		}
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
			// fmt.Printf("slave %d: should i set f[ %d ] = %d: %t\n", slave.rank, v1, parentProposition, slave.getParent(v1) > parentProposition)
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
		// fmt.Println(slave.rank, "->", foreignNoder, "set f[", foreignNode, "] =", proposedParent)
		mpiSendUintArray([]uint32{foreignNode, proposedParent}, int(foreignNoder),
			TAG_V1_PARENT_PROPOSITION)
	}

	// пока мы еще ничего не получили
	slave.recievedParentProposalsNum = 0
	// шэрим результаты, но прием приоритетнее
	for foreignNode, proposedParent := range slave.foreignf {
		checkIncomingParentProposal(slave)
		sendParentProposition(slave, foreignNode, proposedParent)
	}
	fmt.Printf("slave %d:SENT ALL (%d)\n", slave.rank, len(slave.foreignf))
	// допринимаем все, что должны
	for slave.recievedParentProposalsNum < slave.expectedParentProposalsNum {
		checkIncomingParentProposal(slave)
	}
	fmt.Printf("slave %d: RECVD ALL (%d)\n", slave.rank, slave.expectedParentProposalsNum)
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
