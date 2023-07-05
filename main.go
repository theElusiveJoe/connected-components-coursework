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
	var mode string
	var file string
	var routersNum int
	flag.StringVar(&mode, "mode", "normal", "one of: 'normal', 'mpi'")
	flag.StringVar(&file, "f", "", "graph table or json, that we want to process with mpi")
	flag.IntVar(&routersNum, "r", 3, "routers number")
	flag.Parse()

	if mode == "mpi" {
		fmt.Println(file)
		fastSVMpi.Run(file, routersNum)
		return
	} else if mode == "normal" {

	} else {
		panic("UNKNOWN mode: " + mode)
	}

}
