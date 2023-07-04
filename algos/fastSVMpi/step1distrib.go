package fastSVMpi

import (
	"connectedComponents/distribution"
	"connectedComponents/graph"
	"connectedComponents/utils/ioEdges"
)

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
	// читаем граф и загружаем его в итератор
	g := ioEdges.LoadGraph(tr.filename)
	var iterator graph.GraphIterator
	iterator.Init(g)

	// находим распределение [хеш ноды]:[номер слейва]
	hashToSlave := distribution.FindDistribution(
		&iterator, uint32(tr.slavesNum), uint32(tr.hashNum),
	)

	// распределяем ребра по слейвам
	iterator.StartIter()
	for iterator.HasEdges() {
		v1, v2 := iterator.GetNextEdge()
		h1, h2 := v1%tr.hashNum, v2%tr.hashNum
		v1er, v2er := hashToSlave[h1], hashToSlave[h2]
		mpiSendUintArray([]uint32{v1, v2}, int(v1er), TAG_RESPONSIBLE_FOR_V1_CONNECTED_TO_V2)
		if v1er != v2er {
			mpiSendUintArray([]uint32{v2, v1}, int(v2er), TAG_RESPONSIBLE_FOR_V1_CONNECTED_TO_V2)
		}
	}
	for hash, slave := range hashToSlave {
		routerNum := hash%int(tr.hashNum) + 1
		mpiSendUintArray([]uint32{uint32(hash), slave+1+uint32(tr.routersNum)}, routerNum, TAG_HASH_TO_SLAVE_ROW)
	}

	// закончили этап
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize-1)
}

func runStep1Router(tr *transRole) {
	tr.router.init()

	for {
		// принимаем входящие сообщения о ребрах
		for mpiCheckIncoming(TAG_HASH_TO_SLAVE_ROW) {
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_HASH_TO_SLAVE_ROW)
			hash, slave := arr[0], arr[1]
			tr.router.addRecord(hash, slave)
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
		for mpiCheckIncoming(TAG_RESPONSIBLE_FOR_V1_CONNECTED_TO_V2) {
			arr, _ := mpiRecvUintArray(2, MASTER, TAG_RESPONSIBLE_FOR_V1_CONNECTED_TO_V2)
			v1, v2 := arr[0], arr[1]
			tr.slave.addEdge(v1, v2)
		}
		// проверяем на остановку
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}
	}
}
