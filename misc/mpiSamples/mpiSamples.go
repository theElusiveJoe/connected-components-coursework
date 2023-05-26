package mpiSamples

/*
#include "mpi.h"
#include "stdlib.h"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi

static int* makeArray(int size) {
	return (int*)malloc(sizeof(int) * size);
}

static void setArray(int *a, int s, int n) {
	a[n] = s;
}
static int getArray(int *a, int n) {
	return a[n];
}

static void freeArray(int *a) {
	free(a);
}

static MPI_Status createStatus(){
	MPI_Status status;
	return status;
}

*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

func intptr(v *int) *C.int {
	return (*C.int)(unsafe.Pointer(v))
}
func intprost(v int) C.int {
	return (C.int)(v)
}

func toGoArr(arr *C.int, l int) []int {
	list := make([]int, l)
	for i := 0; i < l; i++ {
		list[i] = (int)(C.getArray(arr, (C.int)(i)))
	}
	return list
}

func BcastExample() {
	C.MPI_Init(nil, nil)

	var rank, size int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intptr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intptr(&size))

	// fmt.Printf("1 im %d of %d\n", rank, size)

	var value int
	if rank == 0 {
		value = 100500
	}
	// fmt.Printf("2 im %d of %d: %d\n", rank, size, value)
	fmt.Printf("1 --- im %d of %d: %d\n", rank, size, value)

	C.MPI_Bcast(
		unsafe.Pointer(&value),
		1,
		C.MPI_INT,
		0,
		C.MPI_COMM_WORLD,
	)
	C.MPI_Barrier(C.MPI_COMM_WORLD)
	if rank == 0 {
		println("#######barrier##########")
	}
	C.MPI_Barrier(C.MPI_COMM_WORLD)

	fmt.Printf("2 --- im %d of %d: %d\n", rank, size, value)

	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
}

func ScatterExample() {
	root := 0

	C.MPI_Init(nil, nil)
	var rank, size int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intptr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intptr(&size))

	lTotal, lRecv, lSend := 15, 3, 3
	fmt.Println(lTotal, lRecv, lSend)
	var l int
	if rank == root {
		l = lTotal
	} else {
		l = lRecv
	}

	buf := C.makeArray((C.int)(0))
	if rank == root {
		buf = C.makeArray((C.int)(lTotal))
		for i := 0; i < lTotal; i++ {
			C.setArray(buf, (C.int)(lTotal-i), (C.int)(i))
		}
	} else {
		buf = C.makeArray((C.int)(lRecv))
	}

	fmt.Printf("1 --- im %d of %d: %v\n", rank, size, toGoArr(buf, l))

	C.MPI_Scatter(
		unsafe.Pointer(buf), // send_data
		intprost(lSend),     // send_count
		C.MPI_INT,           // send_datatype
		unsafe.Pointer(buf), // recv_data
		intprost(lRecv),     // recv_size
		C.MPI_INT,           // recv_datatype
		intprost(root),      // root
		C.MPI_COMM_WORLD,    // communicator
	)

	C.MPI_Barrier(C.MPI_COMM_WORLD)
	if rank == root {
		println("#######barrier##########")
	}
	C.MPI_Barrier(C.MPI_COMM_WORLD)

	fmt.Printf("2 --- im %d of %d: %v\n", rank, size, toGoArr(buf, l))

	C.freeArray(buf)
	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
}

func SendRecvExample() {
	C.MPI_Init(nil, nil)
	var rank, size int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intptr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intptr(&size))

	arr := C.makeArray(C.int(9999))
	var n int

	if rank == 0 {
		for {
			C.MPI_Send(
				unsafe.Pointer(arr), // посылаем секретное число
				9999,
				C.MPI_INT,
				intprost(1), // нашей цели
				0,
				C.MPI_COMM_WORLD,
			)
			n++
			fmt.Println("SENT", n)
		}
	}

	if rank == 1 {
		for {
			C.MPI_Recv(
				unsafe.Pointer(arr), // получаем чужое секретное число
				9999,
				C.MPI_INT,
				C.MPI_ANY_SOURCE,
				C.MPI_ANY_TAG,
				C.MPI_COMM_WORLD,
				C.MPI_STATUS_IGNORE,
			)
			n++
			fmt.Println("RECV", n)
		}
	}

	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
}

func RecvIfExistsExample() {
	C.MPI_Init(nil, nil)
	var rank, size int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intptr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intptr(&size))
	fmt.Println("Hello, im", rank)
	var secretNum int

	if rank == 0 {
		fmt.Println("0 initalized")
		for i := 0; i < 15; i++ {
			fmt.Println("iter")
			secretNum++
			C.MPI_Send(
				unsafe.Pointer(&secretNum), // посылаем рандомное число
				1,
				C.MPI_INT,
				1, // нашей цели
				0, // tag
				C.MPI_COMM_WORLD,
			)
			fmt.Println("SENT", secretNum)
			time.Sleep(100 * time.Microsecond)
		}
	} else if rank == 1 {
		fmt.Println("1 initalized")
		var iternum int32
		for true {
			fmt.Println("		iter", iternum, secretNum)
			iternum++

			var flag C.int
			var ammount C.int
			var status C.MPI_Status
			// var req C.MPI_Request

			C.MPI_Iprobe(C.MPI_ANY_SOURCE, C.MPI_ANY_TAG, C.MPI_COMM_WORLD, &flag, C.MPI_STATUS_IGNORE)
			// fmt.Println("FLAG:", flag, "RES:", res)
			if flag == 1 {
				C.MPI_Recv(
					unsafe.Pointer(&secretNum), // получаем чужое секретное число
					1,                          // 1 штука
					C.MPI_INT,                  //
					C.MPI_ANY_SOURCE,           // откуда угодно
					0,                          // tag
					C.MPI_COMM_WORLD,
					&status,
					// &req,
				)

				C.MPI_Get_count(&status, C.MPI_INT, &ammount)
				// fmt.Println("			RECIEVED", secretNum)
				fmt.Println("			RECIEVED", secretNum, "FROM", status.MPI_SOURCE)

			}
			if secretNum == 10 {
				break
			}
			// time.Sleep(3 * time.Second)
		}
	}

	C.MPI_Barrier(C.MPI_COMM_WORLD)
	C.MPI_Abort(C.MPI_COMM_WORLD, 0)
}
