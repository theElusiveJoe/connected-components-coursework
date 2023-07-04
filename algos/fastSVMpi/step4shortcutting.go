package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

func runStep4ShortCutting(tr *transRole) {
	switch tr.role {
	case MASTER:
		runStep4Master(tr)
	case ROUTER:
		runStep4Router(tr)
	case SLAVE:
		runStep4Slave(tr)
	}
}

func runStep4Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiCheckIncoming(TAG_SC_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize-1)
}

func runStep4Router(tr *transRole) {
	for {
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}

		// N1 -> T1 RECV
		for mpiCheckIncoming(TAG_SC1) {
			arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SC1)
			N2 := tr.router.getSlaveRank(tr, arr[0])
			// T1 -> N2 SEND
			mpiSendUintArray(arr, N2, TAG_SC2)
		}
	}
}

func runStep4Slave(tr *transRole) {
	checkIncoming := func(tr *transRole, confirmations *uint32) bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N2 -> N1: TAG3 RECV
			if mpiCheckIncoming(TAG_SC3) {
				for mpiCheckIncoming(TAG_SC3) {
					(*confirmations)++
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_SC3)
					ppu, u := arr[0], arr[1]
					tr.slave.setParentIfLess(u, ppu)
				}
			}

			// T1 -> N2 RECV
			if mpiCheckIncoming(TAG_SC2) {
				arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SC2)
				pu, u, N1 := arr[0], arr[1], arr[2]
				ppu := tr.slave.getParent(pu)
				// N2 -> N1 SEND
				mpiSendUintArray([]uint32{ppu, u}, int(N1), TAG_SC3)
			}

		}
		return true
	}

	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)

	for u, pu := range tr.slave.f {
		checkIncoming(tr, &confirmations)
		N1 := uint32(tr.slave.rank)
		T1 := tr.findRouter(pu)
		mpiSendUintArray([]uint32{pu, u, N1}, T1, TAG_SC1)
	}

	for confirmations < uint32(len(tr.slave.f)) {
		checkIncoming(tr, &confirmations)
	}
	mpiSendTag(TAG_SC_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
}
