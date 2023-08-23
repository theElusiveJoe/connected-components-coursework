package fastSVMpiNoDist

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import "fmt"

func runStep4ShortCutting(tr *transRole) {
	if tr.role == MASTER {
		fmt.Print("\n-----------STEP 4 STARTED-----------\n\n")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	switch tr.role {
	case MASTER:
		runStep4Master(tr)
	case SLAVE:
		runStep4Slave(tr)
	}
}

func runStep4Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiSkipIncoming(TAG_SC_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep4Slave(tr *transRole) {
	checkIncoming := func(tr *transRole, confirmations *uint32) bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N2 -> N1: TAG2 RECV
			if mpiCheckIncoming(TAG_SC2) {
				for mpiCheckIncoming(TAG_SC2) {
					(*confirmations)++
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_SC2)
					ppu, u := arr[0], arr[1]
					tr.slave.setParentIfLess(u, ppu)
				}
			}

			// N1 -> N2 TAG1 RECV
			if mpiCheckIncoming(TAG_SC1) {
				arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_SC1)
				pu, u := arr[0], arr[1]
				ppu := tr.slave.getParent(pu)
				// N2 -> N1 SEND
				mpiSendUintArray([]uint32{ppu, u}, tr.getServer(u), TAG_SC2)
			}

		}
		return true
	}

	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)

	for u, pu := range tr.slave.f {
		checkIncoming(tr, &confirmations)
		N2 := tr.getServer(pu)
		mpiSendUintArray([]uint32{pu, u}, N2, TAG_SC1)
	}

	for confirmations < uint32(len(tr.slave.f)) {
		checkIncoming(tr, &confirmations)
	}
	mpiSendTag(TAG_SC_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
}
