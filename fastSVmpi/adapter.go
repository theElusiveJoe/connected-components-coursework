package fastSVmpi

import (
	"connectedComponents/utils"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func AdapterForMpiBasicCCSearch(filename string) []uint32 {
	// mpiexec -n 4 -oversubscribe main --mpirun=tests/graphs/graph10.csv
	cmd := exec.Command(
		"mpiexec", []string{"-n", "4", "-oversubscribe", "main", "--mpirun=" + filename + ""}...,
	)
	fmt.Println("->", cmd)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("ошибка в MPI BASIC CC SEARCH")
		panic(err)
	}

	rows := utils.GetEdgesReader("temp.csv")
	res := make([]uint32, len(rows))
	for i := 0; i < len(rows); i++ {
		x, _ := strconv.Atoi(rows[i][0])
		xParent, _ := strconv.Atoi(rows[i][1])
		res[x] = uint32(xParent)
	}
	os.Remove("temp.csv")
	return res
}
