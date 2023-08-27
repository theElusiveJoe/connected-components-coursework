package fastSVMpi

import (
	"connectedComponents/src/distribution"
	"connectedComponents/src/graph"
	"connectedComponents/src/utils/ioEdges"
	"fmt"
)

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

// стараемся максимально не захломлять структуры master, slave и router
// чтобы не следить за памятью
// например, distribution нужен только на стадии 0
// поэтому пусть distributor будет локальной переменной в runStep1Master

func runStep1Distrib(tr *transRole) {
	switch tr.role {
	case MASTER:
		runStep1Master(tr)
	case ROUTER:
		runStep1Router(tr)
	case SLAVE:
		runStep1Slave(tr)
	}
}

func runStep1Master(tr *transRole) {
	var submits, expects uint32
	checkSubmits := func(submits *uint32) {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			mpiSkipIncoming(C.MPI_ANY_TAG)
			(*submits)++
		}
	}

	fmt.Print("\n-----------STEP 1 STARTED-----------\n\n")
	// читаем граф и загружаем его в итератор
	g := ioEdges.LoadGraph(tr.filename)
	var iterator graph.GraphIterator
	iterator.Init(g)

	// находим распределение [хеш ноды]:[номер слейва]
	hashToSlave := distribution.FindDistribution(
		&iterator, uint32(tr.slavesNum),
		tr.hashNum,
		// пока тут поставим тождественное отображение {индекс ноды} -> {хеш ноды}
		// uint32(iterator.NodesNum()-20), // uint32(tr.hashNum)*20,
	)

	// распределяем ребра по слейвам
	iterator.StartIter()
	tr.talk("start distribution around slaves")
	for iterator.HasEdges() {
		checkSubmits(&submits)
		v1, v2 := iterator.GetNextEdge()
		h1, h2 := v1%tr.hashNum, v2%tr.hashNum
		v1er, v2er := hashToSlave[h1], hashToSlave[h2]
		v1er += uint32(tr.routersNum) + 1
		v2er += uint32(tr.routersNum) + 1
		// tr.talk("send edge (%d, %d)", v1, v2)
		if v1er == v2er {
			mpiSendUintArray(
				[]uint32{v1, v2},
				int(v1er),
				TAG_INNER_EDGE,
			)
			expects++
		} else {
			mpiSendUintArray(
				[]uint32{v1, v2},
				int(v1er),
				TAG_DISTRIBUTED_EDGE,
			)
			mpiSendUintArray(
				[]uint32{v2, v1},
				int(v2er),
				TAG_DISTRIBUTED_EDGE,
			)
			expects += 2
		}
	}

	tr.talk("start distribution around routers")
	for hash, slave := range hashToSlave {
		checkSubmits(&submits)
		routerNum := hash%int(tr.routersNum) + 1
		mpiSendUintArray(
			[]uint32{uint32(hash), slave + 1 + uint32(tr.routersNum)},
			routerNum,
			TAG_HASH_TO_SLAVE_ROW,
		)
		expects++
	}

	// убедимся, что все дошло
	for submits < expects {
		checkSubmits(&submits)
	}
	// закончили этап
	tr.talk("next phase")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep1Router(tr *transRole) {
	tr.router.init()

	for {
		// принимаем входящие сообщения о ребрах
		for mpiCheckIncoming(TAG_HASH_TO_SLAVE_ROW) {
			mpiSendTag(TAG_GOT_ROW, MASTER)
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_HASH_TO_SLAVE_ROW)
			hash, slave := arr[0], arr[1]
			tr.router.addRecord(hash, slave)
			// tr.talk("recvd hash %d: slave %d", hash, slave)
		}
		// проверяем на остановку
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}
	}
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
			// tr.talk("recvd edge %d <-> %d", v1, v2)
		}

		for mpiCheckIncoming(TAG_INNER_EDGE) {
			mpiSendTag(TAG_GOT_EDGE, MASTER)
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_INNER_EDGE)
			v1, v2 := arr[0], arr[1]
			tr.slave.addEdge(v1, v2, true)
			// tr.talk("recvd edge %d <-> %d", v1, v2)
		}

		// проверяем на остановку
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}
	}
}
