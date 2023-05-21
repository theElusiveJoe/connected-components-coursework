package main

// "connectedComponents/fastSV"
// "connectedComponents/fastSVmpi"
// "connectedComponents/utils"

// "flag"
// "fmt"

/*
#cgo linux LDFLAGS: -pthread -L/usr/lib/x86_64-linux-gnu/openmpi/lib -lmpi
#include "mpi.h"
#include "stdlib.h"
#include <stdint.h>
*/
import "C"
import (
	"connectedComponents/distribution"
	"connectedComponents/fastSV"
	"connectedComponents/fastSVmpi"
	"connectedComponents/utils"
	"flag"
	"fmt"
)

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
	// res:= fastSVmpi.BasicMpiCCSearch(testfile)
	// fmt.Pr

	// utils.CompareManyTests(
	// 	fastSVmpi.AdapterForMpiBasicCCSearch,
	// 	fastSV.BasicCCSearch,
	// 	"tests/graphs/",
	// )

	// utils.CompareOneTest(
	// 	fastSVmpi.AdapterForMpiBasicCCSearch,
	// 	fastSV.BasicCCSearch,
	// 	"tests/graphs/graph1.csv",
	// )
	test := "tests/graphs2/synthGraph-14l-190e.csv"
	var dist distribution.Distributor
	dist.FindDistributionFromFile(test, 10, 2294)

	forest := fastSV.BasicCCSearch(test)
	components := utils.StarForestToComponents(forest)
	fmt.Println(len(components), "COMPONENTS")
}
