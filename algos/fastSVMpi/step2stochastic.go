package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

func runStep2Stochastic(tr *transRole) {
	switch tr.role {
	case MASTER:
		runStep2Master(tr)
	case ROUTER:
		runStep2Router(tr)
	case SLAVE:
		runStep2Slave(tr)
	}
}

func runStep2Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiCheckIncoming(TAG_SH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize-1)
}

func runStep2Router(tr *transRole) {
	for {
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}
		tag, i := C.int(-1), 0
		if mpiCheckIncoming(TAG_SH5) {
			tag, i = TAG_SH5, 0
		} else if mpiCheckIncoming(TAG_SH3) {
			tag, i = TAG_SH3, 1
		} else if mpiCheckIncoming(TAG_SH1) {
			tag, i = TAG_SH1, 1
		}

		for mpiCheckIncoming(tag) {
			arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, tag)
			Ni := tr.router.getSlaveRank(tr, arr[i])
			// T3 -> N4: TAG6 SEND
			mpiSendUintArray(arr, Ni, tag+1)
		}

		// !!!!! не удалять - потом сюда можно добавить оптимизации

		// // N3 -> T3: TAG5 RECV
		// if mpiCheckIncoming(TAG_SH5){
		// 	for mpiCheckIncoming(TAG_SH5){
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH5)
		// 		N4 = tr.router.getSlaveRank(arr[0])
		// 		// T3 -> N4: TAG6 SEND
		// 		mpiSendUintArray(arr, N4, TAG_SH6)
		// 	}
		// 	continue
		// }

		// // N2 -> T2: TAG3 RECV
		// if mpiCheckIncoming(TAG_SH3){
		// 	for mpiCheckIncoming(TAG_SH3){
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH3)
		// 		N3 = tr.router.getSlaveRank(arr[1])
		// 		// T2 -> N3: TAG4 SEND
		// 		mpiSendUintArray(arr, N3, TAG_SH4)
		// 	}
		// 	continue
		// }
		// // N1 -> T1: TAG1 RECV
		// if mpiCheckIncoming(TAG_SH1){
		// 	for mpiCheckIncoming(TAG_SH1){
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH1)
		// 		N2 = tr.router.getSlaveRank(arr[1])
		// 		// T2 -> N3: TAG4 SEND
		// 		mpiSendUintArray(arr, N3, TAG_SH4)
		// 	}
		// 	continue
		// }
	}
}

func runStep2Slave(tr *transRole) {
	checkIncoming := func(tr *transRole, confirmations *uint32) bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N4 -> N1: TAG7 RECV
			if mpiCheckIncoming(TAG_SH7) {
				for mpiCheckIncoming(TAG_SH7) {
					(*confirmations)++
					mpiSkipIncoming(TAG_SH7)
				}
				continue
			}
			// T3 -> N4: TAG6 RECV
			if mpiCheckIncoming(TAG_SH6) {
				for mpiCheckIncoming(TAG_SH6) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH6)
					pu, ppv, N1 := arr[0], arr[1], arr[2]
					tr.slave.setParentIfLess(pu, ppv)
					// N4 -> N1: TAG7 SEND
					mpiSendTag(TAG_SH7, int(N1))
				}
				continue
			}
			// T2 -> N3: TAG4 RECV
			if mpiCheckIncoming(TAG_SH4) {
				for mpiCheckIncoming(TAG_SH4) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH4)
					pu, pv, N1 := arr[0], arr[1], arr[2]
					T3 := tr.findRouter(pv)
					// N3 -> T3: TAG5 SEND
					mpiSendUintArray([]uint32{pu, pv, N1}, T3, TAG_SH5)
				}
				continue
			}
			// T1 -> N2: TAG2 RECV
			if mpiCheckIncoming(TAG_SH2) {
				for mpiCheckIncoming(TAG_SH2) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH2)
					pu, v, N1 := arr[0], arr[1], arr[2]
					T2 := tr.findRouter(v)
					// N2 -> T2: TAG3 SEND
					mpiSendUintArray([]uint32{pu, v, N1}, T2, TAG_SH3)
				}
				continue
			}
		}
		return true
	}

	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)

	for i := 0; i < tr.slave.edgesNum; i++ {
		// чем отправлять свои сообщения, лучше ответим на чужие
		checkIncoming(tr, &confirmations)

		// отправляем собственное сообщение
		u, v := tr.slave.getEdge(i)
		if tr.slave.isServerOf(v) {
			// если ребро лежит в этом слейве
			T1 := tr.findRouter(v)
			N1 := uint32(tr.slave.rank)
			pu, pv := tr.slave.f[u], tr.slave.f[v]
			mpiSendUintArray([]uint32{pu, pv, N1}, T1, TAG_SH3)
		} else {
			// в вобщем случае - распределенное ребро
			// N1 -> T1: TAG1 SEND
			T1 := tr.findRouter(v)
			N1 := uint32(tr.slave.rank)
			pu := tr.slave.f[u]
			mpiSendUintArray([]uint32{pu, v, N1}, T1, TAG_SH1)
		}
	}

	for confirmations < uint32(tr.slave.edgesNum) {
		checkIncoming(tr, &confirmations)
	}
	mpiSendTag(TAG_SH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {}
}
