package fastSVMpi

import (
	"connectedComponents/graph"
	"connectedComponents/utils/ioEdges"
	"fmt"
	"os"
	"os/exec"
)

// адаптер для mpi версий алгоритма запускает сторонние процессы, а их результаkт собирает из файла
func AdapterForFastSVMPI(filename string) *graph.Graph {
	// mpiexec -n 4 -oversubscribe main --mpirun=tests/graphs/graph10.csv
	cmd := exec.Command(
		"mpiexec", []string{"-n", "6", "-oversubscribe", "main", "--mpirun=" + filename + ""}...,
	)
	fmt.Println("->", cmd)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("ошибка в MPI BASIC CC SEARCH")
		panic(err)
	}

	g := ioEdges.LoadGraph("temp.csv")
	os.Remove("temp.csv")
	return g
}
