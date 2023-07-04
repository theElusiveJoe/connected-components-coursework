package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"


// true, если нужно прогрнать еще один круг, иначе false
func runStep5SlavePolling(tr *transRole) bool {
	switch tr.role {
	case MASTER:
		return runStep5Master(tr)
	case ROUTER:
		return runStep5Router(tr)
	case SLAVE:
		return runStep5Slave(tr)
	}
	return true
}

func runStep5Master(tr *transRole) bool {
	i:= 0
	changed := false
	for i < tr.slavesNum{
		ch, _ := mpiRecvBool(TAG_SP1)
		changed = changed || ch
		i++
	}
	mpiBcastBoolViaSend(changed, TAG_SP2, 1, tr.worldSize-1)
	return changed
}

func runStep5Router(tr *transRole) bool {
	cont, _ := mpiRecvBool(TAG_SP2)
	return cont
}

func runStep5Slave(tr *transRole) bool {
	mpiSendBool(tr.slave.changed, MASTER, TAG_SP1)
	cont, _ := mpiRecvBool(TAG_SP2)
	return cont
}