package fastSVMpiNoDist

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import "fmt"

func runStep2Stochastic(tr *transRole) {
	if tr.role == MASTER {
		fmt.Print("\n-----------STEP 2 STARTED-----------\n\n")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	switch tr.role {
	case MASTER:
		runStep2Master(tr)
	case SLAVE:
		runStep2Slave(tr)
	}
}

func runStep2Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiSkipIncoming(TAG_SH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
		tr.talk("recvd %d of %d", recvd, expect)
	}
	tr.talk("all seqs ended")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep2Slave(tr *transRole) {
	checkIncoming := func(tr *transRole, confirmations *uint32) bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N4 -> N1: TAG4 RECV
			if mpiCheckIncoming(TAG_SH4) {
				for mpiCheckIncoming(TAG_SH4) {
					tr.talk("N4 -> N1(%d) -> SEQ ENDED (recv %d)", tr.rank, *confirmations)
					(*confirmations)++
					mpiSkipIncoming(TAG_SH4)
				}
				continue
			}
			// N3 -> N4: TAG3 RECV
			if mpiCheckIncoming(TAG_SH3) {
				for mpiCheckIncoming(TAG_SH3) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH3)
					pu, ppv, N1 := arr[0], arr[1], arr[2]
					tr.slave.setParentIfLess(pu, ppv)
					// N4 -> N1: TAG4 SEND
					mpiSendTag(TAG_SH4, int(N1))
					tr.talk("N3 -> N4(%d) -> N1(%d)", tr.rank, N1)
				}
				continue
			}
			// N2 -> N3: TAG2 RECV
			if mpiCheckIncoming(TAG_SH2) {
				for mpiCheckIncoming(TAG_SH2) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH2)
					pu, pv, N1 := arr[0], arr[1], arr[2]
					ppv := tr.slave.getParent(pv)
					N4 := tr.getServer(pu)
					// N3 -> N4: TAG3 SEND
					mpiSendUintArray([]uint32{pu, ppv, N1}, N4, TAG_SH3)
					tr.talk("N2 -> N3(%d) -> N4(%d)", tr.rank, N4)
				}
				continue
			}
			// N1 -> N2: TAG1  RECV
			if mpiCheckIncoming(TAG_SH1) {
				for mpiCheckIncoming(TAG_SH1) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH1)
					pu, v, N1 := arr[0], arr[1], arr[2]
					pv := tr.slave.getParent(v)
					N3 := tr.getServer(pv)
					// N2 -> N3: TAG2 SEND
					mpiSendUintArray([]uint32{pu, pv, N1}, N3, TAG_SH2)
					tr.talk("N1 -> N2(%d) -> N3(%d)", tr.rank, N3)
				}
				continue
			}
		}
		return true
	}

	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)
	expectations := uint32(0)

	for i := 0; i < tr.slave.edgesNum; i++ {
		// чем отправлять свои сообщения, лучше ответим на чужие
		checkIncoming(tr, &confirmations)

		// отправляем собственное сообщение
		u, v := tr.slave.getEdge(i)
		tr.talk("i init seq for edge (%d, %d)", u, v)
		if tr.slave.isServerOf(v) {
			// если ребро лежит в этом слейве
			// N2 -> N3: TAG2 SEND
			N1 := uint32(tr.slave.rank)
			pu, pv := tr.slave.getParent(u), tr.slave.getParent(v)

			N2u, N2v := tr.getServer(pu), tr.getServer(pv)

			mpiSendUintArray([]uint32{pu, pv, N1}, N2v, TAG_SH2)
			mpiSendUintArray([]uint32{pv, pu, N1}, N2u, TAG_SH2)
			expectations += 2

			tr.talk("SEQ START -> N2(%d) -> N3(%d) edge(%d, %d)", tr.rank, N2v, u, v)
			tr.talk("SEQ START -> N2(%d) -> N3(%d) edge(%d, %d)", tr.rank, N2u, v, u)

		} else {
			// в общем случае - распределенное ребро
			// N1 -> N2: TAG1 SEND
			N1 := uint32(tr.slave.rank)
			pu := tr.slave.getParent(u)
			N2 := tr.getServer(v)
			mpiSendUintArray([]uint32{pu, v, N1}, N2, TAG_SH1)
			expectations++
			tr.talk("SEQ START -> N1(%d) -> N2(%d) edge(%d, %d)", tr.rank, N2, u, v)
		}
	}

	for confirmations < expectations {
		checkIncoming(tr, &confirmations)
		tr.talk("%d of %d recieved", confirmations, expectations)
	}

	mpiSendTag(TAG_SH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
	tr.talk("ENDED!")
}
