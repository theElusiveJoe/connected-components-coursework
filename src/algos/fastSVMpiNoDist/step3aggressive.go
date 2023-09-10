package fastSVMpiNoDist

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"
import "fmt"

func runStep3Aggressive(tr *transRole) {
	if tr.role == MASTER {
		fmt.Print("\n-----------STEP 3 STARTED-----------\n\n")
	}
	mpiBarrier(C.MPI_COMM_WORLD)

	switch tr.role {
	case MASTER:
		runStep3Master(tr)
	case SLAVE:
		runStep3Slave(tr)
	}
}

func runStep3Master(tr *transRole) {
	expect := tr.slavesNum
	recvd := 0
	for recvd < expect {
		mpiSkipIncoming(TAG_AH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	// tr.talk("ALL SLAVES ENDED !!!")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep3Slave(tr *transRole) {
	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)
	expectations := uint32(0)

	checkIncoming := func() bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N3 -> N1: TAG3 RECV
			if mpiCheckIncoming(TAG_AH3) {
				for mpiCheckIncoming(TAG_AH3) {
					confirmations++
					if confirmations%1000 == 0 {
						tr.talk("step: 3, seqs: %d of %d", confirmations, expectations)
					}
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_AH3)
					tr.log("recv_tag_3")
					u, ppv := arr[0], arr[1]
					tr.slave.setParentIfLess(u, ppv)
				}
				continue
			}
			// N2 -> N3: TAG2 RECV
			if mpiCheckIncoming(TAG_AH2) {
				for mpiCheckIncoming(TAG_AH2) {
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_AH2)
					// tr.log("recv_tag_2")
					u, pv := arr[0], arr[1]
					ppv := tr.slave.getParent(pv)
					// N3 -> N1: TAG3 SEND
					mpiSendUintArray([]uint32{u, ppv}, tr.getServer(u), TAG_AH3)
					tr.log("send_tag_3")
				}
				continue
			}
			// N1 -> N2: TAG1 RECV
			if mpiCheckIncoming(TAG_AH1) {
				for mpiCheckIncoming(TAG_AH1) {
					arr, _ := mpiRecvUintArray(2, C.MPI_ANY_SOURCE, TAG_AH1)
					// tr.log("recv_tag_1")

					u, v := arr[0], arr[1]
					pv := tr.slave.getParent(v)
					N3 := tr.getServer(pv)
					// N2 -> N3: TAG2 SEND
					mpiSendUintArray([]uint32{u, pv}, N3, TAG_AH2)
					tr.log("send_tag_2")
				}
				continue
			}
		}
		return true
	}

	for i := 0; i < tr.slave.edgesNum; i++ {
		// чем отправлять свои сообщения, лучше ответим на чужие
		if expectations-confirmations > 5000 {
			for expectations-confirmations > 5000 {
				checkIncoming()
			}
		} else {
			checkIncoming()
		}

		// отправляем собственное сообщение
		u, v := tr.slave.getEdge(i)
		if tr.slave.isServerOf(v) {
			// если ребро лежит в этом слейве
			pu, pv := tr.slave.getParent(u), tr.slave.getParent(v)
			N2u := tr.getServer(pu)
			N2v := tr.getServer(pv)
			mpiSendUintArray([]uint32{u, pv}, N2v, TAG_AH2)
			mpiSendUintArray([]uint32{v, pu}, N2u, TAG_AH2)
			tr.log("send_tag_2")
			tr.log("send_tag_2")

			expectations += 2
			// tr.talk("SEQ START -> N2(%d) -> N3(%d) edge(%d, %d)", tr.rank, N2v, u, v)
			// tr.talk("SEQ START -> N2(%d) -> N3(%d) edge(%d, %d)", tr.rank, N2u, v, u)
		} else {
			// в вобщем случае - распределенное ребро
			// N1 -> T1: TAG1 SEND
			N2 := tr.getServer(v)
			mpiSendUintArray([]uint32{u, v}, N2, TAG_AH1)
			tr.log("send_tag_1")

			// tr.talk("SEQ START -> N1(%d) -> N2(%d) edge(%d, %d)", tr.rank, N2, u, v)
			expectations++
		}
	}

	for confirmations < expectations {
		checkIncoming()
		// // tr.talk("%d/%d", confirmations, expectations)
	}
	// tr.talk("IM DONE: %d of %d", confirmations, expectations)
	mpiSendTag(TAG_AH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming() {
	}
}
