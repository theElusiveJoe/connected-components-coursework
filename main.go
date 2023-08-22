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
	"connectedComponents/algos/basicMpi"
	"connectedComponents/algos/fastSVMpi"
	"connectedComponents/utils/testing"

	"flag"
	"fmt"
)

const MODE_TESTS = "compare"

func main() {
	var strconf string
	var mode string
	flag.StringVar(&mode, "mode", "", "launch mode")
	flag.StringVar(&strconf, "conf", "", "json dumps of RunConfig")
	flag.Parse()

	if mode == MODE_TESTS {
		testing.RunTestsInFolder("tests/mygraphs", basicMpi.Adapter, basic.Adapter)
	} else {
		fmt.Println("CONFIG:", strconf)
		conf := algos.StrToConfig(strconf[1 : len(strconf)-1])
		fmt.Println(conf)

		if mode == algos.MODE_MPI_FASTSV_WITH_DIST {
			fastSVMpi.Run(conf)
		} else if mode == algos.MODE_MPI_BASIC {
			basicMpi.Run(conf)
		} else {
			panic("UNKNOWN mode: " + mode)
		}
	}
}
