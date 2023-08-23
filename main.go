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
	"connectedComponents/algos/fastSV"
	"connectedComponents/algos/fastSVMpi"
	"connectedComponents/algos/fastSVMpiNoDist"

	"connectedComponents/utils/testing"

	"flag"
	"fmt"
)

func GetAlgoAdapter(algo string) func(conf *algos.RunConfig) map[uint32]uint32 {
	adapters := map[string](func(conf *algos.RunConfig) map[uint32]uint32){
		algos.ALGO_MPI_BASIC:            basicMpi.Adapter,
		algos.ALGO_MPI_FASTSV_NO_DIST:   fastSVMpiNoDist.Adapter,
		algos.ALGO_MPI_FASTSV_WITH_DIST: fastSVMpi.Adapter,
		algos.ALGO_NOMPI_BASIC:          basic.Adapter,
		algos.ALGO_NOMPI_FASTSV:         fastSV.Adapter,
	}
	if foo, ok := adapters[algo]; !ok {
		fmt.Println(algo == algos.ALGO_MPI_BASIC)
		panic(fmt.Sprintf("unknow algo: %s", algo))
	} else {
		return foo
	}
}

func GetMPIRun(algo string) func(conf *algos.RunConfig) {
	adapters := map[string](func(conf *algos.RunConfig)){
		algos.ALGO_MPI_BASIC:            basicMpi.Run,
		algos.ALGO_MPI_FASTSV_NO_DIST:   fastSVMpiNoDist.Run,
		algos.ALGO_MPI_FASTSV_WITH_DIST: fastSVMpi.Run,
	}
	if foo, ok := adapters[algo]; !ok {
		panic(fmt.Sprintf("unable to use algo \"%s\" in MPI_LAUNCH mode", algo))
	} else {
		return foo
	}
}

func main() {
	// создаем флаги
	var mode string
	var algo string
	var strconf string
	var file string
	var slaves int
	var routers int
	var hashnum int

	// парсим флаги
	flag.StringVar(&mode, "mode", algos.MODE_FIND_CC, "launch mode")
	flag.StringVar(&algo, "algo", algos.ALGO_MPI_FASTSV_NO_DIST, "launch mode")
	flag.StringVar(&strconf, "conf", "noConf", "json dumps of RunConfig")
	flag.StringVar(&file, "file", "noFile", "file with graph")
	flag.IntVar(&slaves, "slaves", 4, "num of slaves")
	flag.IntVar(&routers, "routers", 3, "num of routers")
	flag.IntVar(&hashnum, "hashnum", 1000, "hashnum for mpi algo with distribution")
	flag.Parse()

	switch mode {
	// для проверки работы алгоритма на тестах
	case algos.MODE_COMPARE_TO_STANDART:
		algoAdapter := GetAlgoAdapter(algo)
		testing.RunTestsInFolder("tests/mygraphs", algoAdapter, basic.Adapter)
	// запукаем алгоритм на конкретном файле
	case algos.MODE_FIND_CC:
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("<FLAG>    %s: %s\n", f.Name, f.Value)
		})
		config := algos.RunConfig{
			TestFile:   file,
			ResultDir:  "outputs/",
			RoutersNum: routers,
			Slavesnum:  slaves,
			HashNum:    hashnum,
		}
		algoAdapter := GetAlgoAdapter(algo)
		algoAdapter(&config)
		// result := algoAdapter(&config)
		// fmt.Println(result)
	// для запуска MPI
	case algos.MODE_MPI_LAUNCH:
		fmt.Println("CONFIG:", strconf)
		config := algos.StrToConfig(strconf[1 : len(strconf)-1])
		mpiRunFunc := GetMPIRun(algo)
		mpiRunFunc(config)
	default:
		panic("unknown mode:" + mode)
	}
}
