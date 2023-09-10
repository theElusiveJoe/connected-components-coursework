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
	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)
	expectations := uint32(len(tr.slave.f))

	checkIncoming := func() bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N2 -> N1: TAG2 RECV
			if mpiCheckIncoming(TAG_SC2) {
				for mpiCheckIncoming(TAG_SC2) {
					confirmations++
					if confirmations%1000 == 0 {
						tr.talk("step: 4, seqs: %d of %d", confirmations, expectations)
					}
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_SC2)
					// tr.log("recv_tag_2")
					ppu, u := arr[0], arr[1]
					tr.slave.setParentIfLess(u, ppu)
				}
			}

			// N1 -> N2 TAG1 RECV
			if mpiCheckIncoming(TAG_SC1) {
				for mpiCheckIncoming(TAG_SC1) {
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_SC1)
					// tr.log("recv_tag_1")

					pu, u := arr[0], arr[1]
					ppu := tr.slave.getParent(pu)
					// N2 -> N1 SEND
					mpiSendUintArray([]uint32{ppu, u}, tr.getServer(u), TAG_SC2)
					tr.log("send_tag_2")
				}
			}

		}
		return true
	}

	for u, pu := range tr.slave.f {
		checkIncoming()
		N2 := tr.getServer(pu)
		mpiSendUintArray([]uint32{pu, u}, N2, TAG_SC1)
		tr.log("send_tag_1")
	}

	for confirmations < expectations {
		checkIncoming()
	}
	mpiSendTag(TAG_SC_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming() {
	}
}
