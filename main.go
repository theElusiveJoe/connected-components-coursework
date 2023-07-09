package main

/*
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
#include "mpi.h"
#include "stdlib.h"
#include <stdint.h>
*/
import "C"

import (
	"connectedComponents/algos"
	"connectedComponents/algos/fastSVMpi"
	"flag"
	"fmt"
)

func main() {
	var strconf string
	var mode int
	flag.IntVar(&mode, "mode", -1, "-1 if not algo run else 0+")
	flag.StringVar(&strconf, "conf", "", "json dumps of RunConfig")
	flag.Parse()

	if mode == -1 {
		conf := algos.RunConfig{
			TestFile:   "tests/graphs2/synthGraph-1l-90e.csv",
			ResultDir:  "outputs/",
			RoutersNum: 3,
			Slavesnum:  5,
			HashNum:    1000000,
			Id:         "",
		}
		res := fastSVMpi.Adapter(conf)
		fmt.Println(res)
	} else {
		conf := algos.StrToConfig(strconf[1 : len(strconf)-1])
		fmt.Println(conf)

		if mode == algos.MODE_MPI_FASTSV_WITH_DIST {
			fastSVMpi.Run(conf)
		} else if mode == algos.MODE_MPI_FASTSV_NO_DIST {

		} else {
			panic("UNKNOWN mode: " + fmt.Sprintf("%d", mode))
		}
	}

}
