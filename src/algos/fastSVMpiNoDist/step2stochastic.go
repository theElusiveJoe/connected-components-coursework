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
	}
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep2Slave(tr *transRole) {
	// должны получить столько подтверждений, колько цепочек инициировали
	confirmations := uint32(0)
	expectations := uint32(0)

	checkIncoming := func() bool {
		for mpiCheckIncoming(C.MPI_ANY_TAG) {
			if mpiCheckIncoming(TAG_NEXT_PHASE) {
				mpiSkipIncoming(TAG_NEXT_PHASE)
				return false
			}

			// N4 -> N1: TAG4 RECV
			if mpiCheckIncoming(TAG_SH4) {
				for mpiCheckIncoming(TAG_SH4) {
					confirmations++
					if confirmations%1000 == 0 {
						tr.talk("step: 2, seqs: %d of %d", confirmations, expectations)
					}
					mpiSkipIncoming(TAG_SH4)
					tr.log("recv_tag_4")
				}
				continue
			}
			// N3 -> N4: TAG3 RECV
			if mpiCheckIncoming(TAG_SH3) {
				for mpiCheckIncoming(TAG_SH3) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH3)
					// tr.log("recv_tag_3")
					pu, ppv, N1 := arr[0], arr[1], arr[2]
					tr.slave.setParentIfLess(pu, ppv)
					// N4 -> N1: TAG4 SEND
					mpiSendTag(TAG_SH4, int(N1))
					tr.log("send_tag_4")
				}
				continue
			}
			// N2 -> N3: TAG2 RECV
			if mpiCheckIncoming(TAG_SH2) {
				for mpiCheckIncoming(TAG_SH2) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH2)
					// tr.log("recv_tag_2")
					pu, pv, N1 := arr[0], arr[1], arr[2]
					ppv := tr.slave.getParent(pv)
					N4 := tr.getServer(pu)
					// N3 -> N4: TAG3 SEND
					mpiSendUintArray([]uint32{pu, ppv, N1}, N4, TAG_SH3)
					tr.log("send_tag_3")
				}
				continue
			}
			// N1 -> N2: TAG1  RECV
			if mpiCheckIncoming(TAG_SH1) {
				for mpiCheckIncoming(TAG_SH1) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH1)
					// tr.log("recv_tag_1")
					pu, v, N1 := arr[0], arr[1], arr[2]
					pv := tr.slave.getParent(v)
					N3 := tr.getServer(pv)
					// N2 -> N3: TAG2 SEND
					mpiSendUintArray([]uint32{pu, pv, N1}, N3, TAG_SH2)
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
			// N2 -> N3: TAG2 SEND
			N1 := uint32(tr.slave.rank)
			pu, pv := tr.slave.getParent(u), tr.slave.getParent(v)

			N2u, N2v := tr.getServer(pu), tr.getServer(pv)

			mpiSendUintArray([]uint32{pu, pv, N1}, N2v, TAG_SH2)
			mpiSendUintArray([]uint32{pv, pu, N1}, N2u, TAG_SH2)
			expectations += 2

			tr.log("send_tag_2")
			tr.log("send_tag_2")

		} else {
			// в общем случае - распределенное ребро
			// N1 -> N2: TAG1 SEND
			N1 := uint32(tr.slave.rank)
			pu := tr.slave.getParent(u)
			N2 := tr.getServer(v)
			mpiSendUintArray([]uint32{pu, v, N1}, N2, TAG_SH1)
			expectations++
			tr.log("send_tag_1")
		}
	}

	for confirmations < expectations {
		checkIncoming()
	}

	mpiSendTag(TAG_SH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming() {
	}
}
