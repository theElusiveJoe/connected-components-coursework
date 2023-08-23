package fastSVMpiNoDist

import (
	"connectedComponents/graph"
	"connectedComponents/utils/ioEdges"
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

func runStep1Distrib(tr *transRole) {
	switch tr.role {
	case MASTER:
		runStep1Master(tr)
	case SLAVE:
		runStep1Slave(tr)
	}
}

func runStep1Master(tr *transRole) {
	var submits, expects uint32
	checkSubmits := func(submits *uint32) {
		for mpiCheckIncoming(TAG_GOT_EDGE) {
			mpiSkipIncoming(TAG_GOT_EDGE)
			(*submits)++
		}
	}

	fmt.Print("\n-----------STEP 1 STARTED-----------\n\n")

	// читаем граф и загружаем его в итератор
	g := ioEdges.LoadGraph(tr.filename)
	var iterator graph.GraphIterator
	iterator.Init(g)

	// распределяем ребра по слейвам
	iterator.StartIter()
	tr.talk("start distribution around slaves")
	for iterator.HasEdges() {
		checkSubmits(&submits)
		v1, v2 := iterator.GetNextEdge()
		v1er, v2er := tr.getServer(v1), tr.getServer(v2)
		tr.talk("send edge (%d, %d)", v1, v2)

		if v1er == v2er {
			mpiSendUintArray(
				[]uint32{v1, v2},
				v1er,
				TAG_INNER_EDGE,
			)
			expects++
		} else {
			mpiSendUintArray(
				[]uint32{v1, v2},
				v1er,
				TAG_DISTRIBUTED_EDGE,
			)
			mpiSendUintArray(
				[]uint32{v2, v1},
				v2er,
				TAG_DISTRIBUTED_EDGE,
			)
			expects += 2
		}
	}

	// убедимся, что все дошло
	for submits < expects {
		checkSubmits(&submits)
	}

	// закончили этап
	tr.talk("next phase")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep1Slave(tr *transRole) {
	tr.slave.init(tr.rank)

	for {
		// принимаем входящие сообщения о ребрах
		for mpiCheckIncoming(TAG_DISTRIBUTED_EDGE) {
			mpiSendTag(TAG_GOT_EDGE, MASTER)
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_DISTRIBUTED_EDGE)
			v1, v2 := arr[0], arr[1]
			tr.slave.addEdge(v1, v2, false)
			tr.talk("recvd outer edge %d <-> %d", v1, v2)
		}
		for mpiCheckIncoming(TAG_INNER_EDGE) {
			mpiSendTag(TAG_GOT_EDGE, MASTER)
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_INNER_EDGE)
			v1, v2 := arr[0], arr[1]
			tr.slave.addEdge(v1, v2, true)
			tr.talk("recvd inner edge %d <-> %d", v1, v2)
		}
		// проверяем на остановку
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}
	}
}
