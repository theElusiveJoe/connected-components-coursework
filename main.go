package main

import (
	"connectedComponents/fastSV"
	"connectedComponents/fastSVmpi"
	"connectedComponents/utils"

	"flag"
	"fmt"
)

/*
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
#include "mpi.h"
#include "stdlib.h"
#include <stdint.h>
*/
import "C"

// func runMpiCCSearch(filename string){
// 	utils.
// 	fastSVmpi.BasicMpiCCSearch()
// }

// var (
// 	runmpi *string
// )

// func init() {
// 	runmpi = flag.String("runmpi", "", "graph table, that we want to process with mpi")
// }

func main() {
	var mpirun string
	flag.StringVar(&mpirun, "mpirun", "", "graph table, that we want to process with mpi")
	flag.Parse()
	if mpirun != "" {
		fmt.Println(mpirun)
		fastSVmpi.BasicMpiCCSearch(mpirun)
		return
	}

	// testfile := "tests/graphs/graph10.csv"
	// fastSVmpi.AdapterForMpiBasicCCSearch(testfile)

	utils.CompareManyTests(
		fastSVmpi.AdapterForMpiBasicCCSearch,
		fastSV.BasicCCSearch,
		"tests/graphs/",
	)

	utils.CompareOneTest(
		fastSVmpi.AdapterForMpiBasicCCSearch,
		fastSV.BasicCCSearch,
		"tests/graphs/graph1.csv",
	)
}
