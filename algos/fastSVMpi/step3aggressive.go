package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

func runStep3Aggressive(tr *transRole) {
	switch tr.role {
	case MASTER:
		runStep3Master(tr)
	case ROUTER:
		runStep3Router(tr)
	case SLAVE:
		runStep3Slave(tr)
	}
}

func runStep3Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiCheckIncoming(TAG_AH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep3Router(tr *transRole) {
	for {
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}

		tag := C.int(-1)
		if mpiCheckIncoming(TAG_SH3) {
			tag = TAG_SH3
		} else if mpiCheckIncoming(TAG_SH1) {
			tag = TAG_SH1
		}
		for mpiCheckIncoming(tag) {
			arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, tag)
			Ni := tr.router.getSlaveRank(tr, arr[1])
			mpiSendUintArray(arr, Ni, tag+1)
		}
	}
}

func runStep3Slave(tr *transRole) {
	checkIncoming := func(tr *transRole, confirmations *uint32) bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N3 -> N1: TAG5 RECV
			if mpiCheckIncoming(TAG_AH5) {
				for mpiCheckIncoming(TAG_AH5) {
					(*confirmations)++
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_AH5)
					u, ppv := arr[0], arr[1]
					tr.slave.setParentIfLess(u, ppv)
				}
				continue
			}
			// T2 -> N3: TAG4 RECV
			if mpiCheckIncoming(TAG_AH4) {
				for mpiCheckIncoming(TAG_AH4) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_AH4)
					u, ppv, N1 := arr[0], arr[1], arr[2]
					// N3 -> N1: TAG5 SEND
					mpiSendUintArray([]uint32{u, ppv}, int(N1), TAG_AH5)
				}
				continue
			}
			// T1 -> N2: TAG2 RECV
			if mpiCheckIncoming(TAG_AH2) {
				for mpiCheckIncoming(TAG_AH2) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_AH2)
					u, v, N1 := arr[0], arr[1], arr[2]
					pv := tr.slave.getParent(v)
					T2 := tr.findRouter(pv)
					// N2 -> T2: TAG3 SEND
					mpiSendUintArray([]uint32{u, pv, N1}, T2, TAG_SH3)
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
			N1 := uint32(tr.slave.rank)
			pv := tr.slave.getParent(v)
			T2 := tr.findRouter(pv)
			mpiSendUintArray([]uint32{u, pv, N1}, T2, TAG_AH3)
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
	mpiSendTag(TAG_AH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
}
