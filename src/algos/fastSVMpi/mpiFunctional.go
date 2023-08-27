package fastSVMpi

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

import (
	"unsafe"
)

// C stuff
func intPtr(v *int) *C.int {
	return (*C.int)(unsafe.Pointer(v))
}
func cGetArr(arr *C.uint, i int) uint32 {
	return (uint32)(C.getArray(arr, (C.int)(i)))
}

// MPI stuff
func mpiBarrier(communicator C.MPI_Comm) {
	C.MPI_Barrier(communicator)
}

// COMMUNOCATION WITH MASTER
func mpiBcastTagViaSend(tag C.int, a int, b int) {
	for i := a; i < b; i++ {
		mpiSendTag(tag, i)
	}
}
func mpiBcastBoolViaSend(bol bool, tag C.int, a int, b int) {
	for i := a; i < b; i++ {
		mpiSendBool(bol, i, tag)
	}
}

// PROSTO
func mpiCheckIncoming(tag C.int) bool {
	var flag C.int
	C.MPI_Iprobe(C.MPI_ANY_SOURCE, tag, C.MPI_COMM_WORLD, &flag, C.MPI_STATUS_IGNORE)
	return flag == 1
}
func mpiSkipIncoming(tag C.int) {
	C.MPI_Recv(
		unsafe.Pointer(nil),
		0,
		C.MPI_UNSIGNED,
		C.MPI_ANY_SOURCE,
		tag,
		C.MPI_COMM_WORLD,
		C.MPI_STATUS_IGNORE,
	)
}

// TAG
func mpiSkipIncomingAndResponce(intag C.int, outtag C.int) {
	var status C.MPI_Status
	C.MPI_Recv(
		unsafe.Pointer(nil),
		0,
		C.MPI_UNSIGNED,
		C.MPI_ANY_SOURCE,
		intag,
		C.MPI_COMM_WORLD,
		&status,
	)
	mpiSendTag(outtag, int(status.MPI_SOURCE))
}
func mpiSendTag(tag C.int, recipient int) {
	C.MPI_Send(
		unsafe.Pointer(nil),
		0,
		C.MPI_UNSIGNED,
		(C.int)(recipient),
		tag,
		C.MPI_COMM_WORLD,
	)
}

// UINT
func mpiSendUint(num uint32, recipient int, tag C.int) {
	C.MPI_Send(
		unsafe.Pointer(&num), // что посылаем
		1,                    // сколько
		C.MPI_UNSIGNED,       // какого типа
		(C.int)(recipient),   // куда посылаем
		tag,                  // тэг
		C.MPI_COMM_WORLD,     // коммуникатор
	)
}
func mpiRecvUint(tag C.int) (uint32, C.MPI_Status) {
	var num uint32
	var status C.MPI_Status
	C.MPI_Recv(
		unsafe.Pointer(&num),
		1,
		C.MPI_UNSIGNED,
		C.MPI_ANY_SOURCE,
		tag,
		C.MPI_COMM_WORLD,
		&status,
	)
	return num, status
}

// BOOL
func mpiSendBool(b bool, recipient int, tag C.int) {
	C.MPI_Send(
		unsafe.Pointer(&b), // что посылаем
		1,                  // сколько
		C.MPI_C_BOOL,       // какого типа
		(C.int)(recipient), // куда посылаем
		tag,                // тэг
		C.MPI_COMM_WORLD,   // коммуникатор
	)
}
func mpiRecvBool(tag C.int) (bool, C.MPI_Status) {
	var b bool
	var status C.MPI_Status
	C.MPI_Recv(
		unsafe.Pointer(&b),
		1,
		C.MPI_C_BOOL,
		C.MPI_ANY_SOURCE,
		tag,
		C.MPI_COMM_WORLD,
		&status,
	)
	return b, status
}

// UINT ARRAY
func mpiSendUintArray(source []uint32, recipient int, tag C.int) {
	arr := C.createArray((C.int)(len(source)))
	for i := 0; i < len(source); i++ {
		C.setArray(arr, (C.uint)(source[i]), (C.int)(i))
	}
	C.MPI_Send(
		unsafe.Pointer(arr),  // что посылаем
		(C.int)(len(source)), // сколько
		C.MPI_UNSIGNED,       // какого типа
		(C.int)(recipient),   // куда посылаем
		tag,                  // тэг
		C.MPI_COMM_WORLD,     // коммуникатор
	)
	C.freeArray(arr)
}
func mpiRecvUintArray(msgLen int, source int, tag C.int) ([]uint32, C.MPI_Status) {
	arr := C.createArray((C.int)(msgLen))
	var status C.MPI_Status
	C.MPI_Recv(
		unsafe.Pointer(arr),
		(C.int)(msgLen),
		C.MPI_UNSIGNED,
		(C.int)(source),
		tag,
		C.MPI_COMM_WORLD,
		&status,
	)
	x := make([]uint32, msgLen)
	for i := 0; i < msgLen; i++ {
		x[i] = cGetArr(arr, i)
	}
	C.freeArray(arr)
	return x, status
}
