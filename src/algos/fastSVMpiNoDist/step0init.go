package fastSVMpiNoDist

/*
#include "help.c"
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
*/
import "C"

// На этой стадии настраивается только MPI и т.н. "глобальный контекст"
func runStep0(filename string, hashnum uint32) *transRole {
	C.MPI_Init(nil, nil)
	var rank, worldSize int
	C.MPI_Comm_rank(C.MPI_COMM_WORLD, intPtr(&rank))
	C.MPI_Comm_size(C.MPI_COMM_WORLD, intPtr(&worldSize))

	var role int
	if rank == 0 {
		role = MASTER
	} else {
		role = SLAVE
	}

	var slave slaveNode

	// настраивае коммуникатор для слейвов
	var SLAVES_COMM C.MPI_Comm
	var color C.int
	if role == SLAVE {
		color = 1
	} else {
		color = C.MPI_UNDEFINED
	}
	C.MPI_Comm_split(
		C.MPI_COMM_WORLD,
		color,
		(C.int)(rank),
		&SLAVES_COMM,
	)

	tr := transRole{
		filename: filename,

		slave: &slave,

		rank:      rank,
		worldSize: worldSize,
		role:      role,
		slavesNum: worldSize - 1,

		SLAVES_COMM: SLAVES_COMM,
	}
	return &tr
}
