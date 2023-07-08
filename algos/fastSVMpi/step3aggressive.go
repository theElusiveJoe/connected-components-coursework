package fastSVMpi

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
		mpiSkipIncoming(TAG_AH_ALL_CONFIRMATIONS_RECIEVED)
		recvd++
	}
	tr.talk("ALL SLAVES ENDED !!!")
	mpiBcastTagViaSend(TAG_NEXT_PHASE, 1, tr.worldSize)
}

func runStep3Router(tr *transRole) {
	for {
		if mpiCheckIncoming(TAG_NEXT_PHASE) {
			mpiSkipIncoming(TAG_NEXT_PHASE)
			return
		}

		tag := C.int(-1)
		if mpiCheckIncoming(TAG_AH3) {
			tag = TAG_AH3
			tr.talk("TAG_AH3")
		} else if mpiCheckIncoming(TAG_AH1) {
			tag = TAG_AH1
			tr.talk("TAG_AH1")
		} else {
			continue
		}

		arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, tag)
		Ni := tr.router.getSlaveRank(tr, arr[1])
		tr.talk("redirect TAG(%d) to slave %d", tag, Ni)
		mpiSendUintArray(arr, Ni, tag+C.int(1))
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
					tr.talk("N3 -> N1(%d) -> SEQ ENDED (recv %d)", tr.rank, *confirmations)
				}
				continue
			}
			// T2 -> N3: TAG4 RECV
			if mpiCheckIncoming(TAG_AH4) {
				for mpiCheckIncoming(TAG_AH4) {
					arr, _ := mpiRecvUintArray(3, C.MPI_ANY_SOURCE, TAG_AH4)
					u, pv, N1 := arr[0], arr[1], arr[2]
					ppv := tr.slave.getParent(pv)
					// N3 -> N1: TAG5 SEND
					mpiSendUintArray([]uint32{u, ppv}, int(N1), TAG_AH5)
					tr.talk("T2 -> N3(%d) -> N1(%d)", tr.rank, N1)
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
					mpiSendUintArray([]uint32{u, pv, N1}, T2, TAG_AH3)
					tr.talk("T1 -> N2(%d) -> T2(%d)", tr.rank, T2)
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
		if tr.slave.isServerOf(v) {
			// если ребро лежит в этом слейве
			N1 := uint32(tr.slave.rank)
			pu, pv := tr.slave.getParent(u), tr.slave.getParent(v)
			T2u := tr.findRouter(pu)
			T2v := tr.findRouter(pv)
			mpiSendUintArray([]uint32{u, pv, N1}, T2v, TAG_AH3)
			mpiSendUintArray([]uint32{v, pu, N1}, T2u, TAG_AH3)
			expectations += 2
			tr.talk("SEQ START -> N2(%d) -> T2(%d) edge(%d, %d)", tr.rank, T2v, u, v)
			tr.talk("SEQ START -> N2(%d) -> T2(%d) edge(%d, %d)", tr.rank, T2u, v, u)
		} else {
			// в вобщем случае - распределенное ребро
			// N1 -> T1: TAG1 SEND
			T1 := tr.findRouter(v)
			N1 := uint32(tr.slave.rank)
			mpiSendUintArray([]uint32{u, v, N1}, T1, TAG_AH1)
			tr.talk("SEQ START -> N1(%d) -> T1(%d) edge(%d, %d)", tr.rank, T1, u, v)
			expectations++
		}
	}

	for confirmations < expectations {
		checkIncoming(tr, &confirmations)
		// tr.talk("%d/%d", confirmations, expectations)
	}
	tr.talk("IM DONE: %d of %d", confirmations, expectations)
	mpiSendTag(TAG_AH_ALL_CONFIRMATIONS_RECIEVED, MASTER)
	for checkIncoming(tr, &confirmations) {
	}
}
