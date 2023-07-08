package fastSVMpi

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
		mpiSkipIncoming(TAG_SH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
		tr.talk("recvd %d of %d", recvd, expect)
	}
	tr.talk("all seqs ended")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
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
			tr.talk("N3 -> T3 -> N4")
		} else if mpiCheckIncoming(TAG_SH3) {
			tag, i = TAG_SH3, 1
			tr.talk("N2 -> T2 -> N3")
		} else if mpiCheckIncoming(TAG_SH1) {
			tag, i = TAG_SH1, 1
			tr.talk("N1 -> T1 -> N2")
		} else {
			continue
		}

		if mpiCheckIncoming(tag) {
			arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, tag)
			Ni := tr.router.getSlaveRank(tr, arr[i])
			// T3 -> N4: TAG6 SEND
			mpiSendUintArray(arr, Ni, tag+1)
		}

		// !!!!! не удалять - потом сюда можно добавить оптимизации
		// N3 -> T3: TAG5 RECV
		// if mpiCheckIncoming(TAG_SH5) {
		// 	for mpiCheckIncoming(TAG_SH5) {
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH5)
		// 		tr.talk("SH5: %d", arr[0])
		// 		N4 := tr.router.getSlaveRank(tr, arr[0])
		// 		// T3 -> N4: TAG6 SEND
		// 		mpiSendUintArray(arr, N4, TAG_SH6)
		// 	}
		// 	continue
		// }
		// // N2 -> T2: TAG3 RECV
		// if mpiCheckIncoming(TAG_SH3) {
		// 	for mpiCheckIncoming(TAG_SH3) {
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH3)
		// 		tr.talk("SH3: %d", arr[1])
		// 		N3 := tr.router.getSlaveRank(tr, arr[1])
		// 		// T2 -> N3: TAG4 SEND
		// 		mpiSendUintArray(arr, N3, TAG_SH4)
		// 	}
		// 	continue
		// }
		// // N1 -> T1: TAG1 RECV
		// if mpiCheckIncoming(TAG_SH1) {
		// 	for mpiCheckIncoming(TAG_SH1) {
		// 		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH1)
		// 		tr.talk("SH1: %d", arr[1])
		// 		N2 := tr.router.getSlaveRank(tr, arr[1])
		// 		// T2 -> N3: TAG4 SEND
		// 		mpiSendUintArray(arr, N2, TAG_SH2)
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
					tr.talk("N4 -> N1(%d) -> SEQ ENDED (recv %d)", tr.rank, *confirmations)
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
					tr.talk("T3 -> N4(%d) -> N1(%d)", tr.rank, N1)

				}
				continue
			}
			// T2 -> N3: TAG4 RECV
			if mpiCheckIncoming(TAG_SH4) {
				for mpiCheckIncoming(TAG_SH4) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH4)
					pu, pv, N1 := arr[0], arr[1], arr[2]
					ppv := tr.slave.getParent(pv)
					T3 := tr.findRouter(pu)
					// N3 -> T3: TAG5 SEND
					mpiSendUintArray([]uint32{pu, ppv, N1}, T3, TAG_SH5)
					tr.talk("T2 -> N3(%d) -> T3(%d)", tr.rank, T3)
					tr.talk("TAG SH 5:%d->%d", T3, pv)
				}
				continue
			}
			// T1 -> N2: TAG2 RECV
			if mpiCheckIncoming(TAG_SH2) {
				for mpiCheckIncoming(TAG_SH2) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_SH2)
					pu, v, N1 := arr[0], arr[1], arr[2]
					pv := tr.slave.getParent(v)
					T2 := tr.findRouter(pv)
					// N2 -> T2: TAG3 SEND
					mpiSendUintArray([]uint32{pu, pv, N1}, T2, TAG_SH3)
					tr.talk("T1 -> N2(%d) -> T2(%d)", tr.rank, T2)
					tr.talk("TAG SH 3:%d->%d", T2, v)

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
			// N2 -> T2: TAG3 SEND
			N1 := uint32(tr.slave.rank)
			pu, pv := tr.slave.getParent(u), tr.slave.getParent(v)

			T2v := tr.findRouter(pv)
			T2u := tr.findRouter(pu)

			mpiSendUintArray([]uint32{pu, pv, N1}, T2v, TAG_SH3)
			mpiSendUintArray([]uint32{pv, pu, N1}, T2u, TAG_SH3)
			expectations += 2
			tr.talk("SEQ START -> N2(%d) -> T2(%d) edge(%d, %d)", tr.rank, T2v, u, v)
			tr.talk("TAG SH 2_1:%d->%d", T2v, pv)
			tr.talk("SEQ START -> N2(%d) -> T2(%d) edge(%d, %d)", tr.rank, T2u, v, u)
			tr.talk("TAG SH 2_1:%d->%d", T2u, pu)

		} else {
			// в общем случае - распределенное ребро
			// N1 -> T1: TAG1 SEND
			T1 := tr.findRouter(v)
			N1 := uint32(tr.slave.rank)
			pu := tr.slave.getParent(u)
			mpiSendUintArray([]uint32{pu, v, N1}, T1, TAG_SH1)
			expectations++
			tr.talk("SEQ START -> N1(%d) -> T1(%d) edge(%d, %d)", tr.rank, T1, u, v)
			tr.talk("TAG SH 1:%d->%d", T1, v)
		}
	}

	for confirmations < expectations {
		checkIncoming(tr, &confirmations)
		// tr.talk("%d of %d recieved", confirmations, expectations)
	}

	mpiSendTag(TAG_SH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
	tr.talk("ENDED!")
}
