package main

/*
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
#include "mpi.h"
#include "stdlib.h"
#include <stdint.h>
*/
import "C"

import (
	"connectedComponents/algos/fastSVMpi"
	"flag"
	"fmt"
)

func main() {
	var mode, file string
	var routersNum, hashNum int
	flag.StringVar(&mode, "mode", "normal", "one of: 'normal', 'mpi-with-dist', 'mpi-no-dist'")
	flag.StringVar(&file, "file", "", "graph table or json, that we want to process with mpi")
	flag.IntVar(&routersNum, "routers", 3, "routers number")
	flag.IntVar(&hashNum, "hash", 1000000000, "hash for pre distribution")
	flag.Parse()

	if mode == "normal" {

	} else if mode == "mpi-with-dist" {
		fmt.Println(file)
		fastSVMpi.Run(file, routersNum, hashNum)
	} else if mode == "mpi-with-dist" {

	} else {
		panic("UNKNOWN mode: " + mode)
	}

}
