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
	"connectedComponents/algos/basic"
	"connectedComponents/algos/fastSVMpi"
	"connectedComponents/utils/testing"

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
		testing.RunTestsInFolder("tests/mygraphs", fastSVMpi.Adapter, basic.Adapter)
	} else if mode > 0 {
		conf := algos.StrToConfig(strconf[1 : len(strconf)-1])
		fmt.Println(conf)

		if mode == algos.MODE_MPI_FASTSV_WITH_DIST {
			fastSVMpi.Run(conf)
		} else if mode == algos.MODE_NOMPI_BASIC {

		} else {
			panic("UNKNOWN mode: " + fmt.Sprintf("%d", mode))
		}
	} else {
		panic("UNKNOWN mode: " + fmt.Sprintf("%d", mode))

	}

}
