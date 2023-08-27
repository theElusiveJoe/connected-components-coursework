package fastSVMpi

import (
	"connectedComponents/src/algos"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func Adapter(conf *algos.RunConfig) map[uint32]uint32 {
	conf = conf.GetCopy()
	conf.Id = uuid.New().String()[:8]

	cmd := exec.Command(
		"mpiexec", []string{
			"-n", fmt.Sprintf("%d", conf.RoutersNum+conf.Slavesnum+1),
			"-oversubscribe",

			"main",
			"-mode=" + algos.MODE_MPI_LAUNCH,
			"-algo=" + algos.ALGO_MPI_FASTSV_WITH_DIST,
			"-conf=" + "'" + conf.ConfigToStr() + "'",
		}...,
	)

	// fmt.Println("->", cmd)
	if resb, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ошибка в алгоритме MPI FASTSV WITH DIST")
		fmt.Println(string(resb))
		panic(err)
	}

	res := make(map[uint32]uint32)
	pattern := conf.ResultDir + "/" + conf.Id + "*.mapbin"
	files, _ := filepath.Glob(pattern)
	for _, fn := range files {
		r := make(map[uint32]uint32)
		file, _ := os.Open(fn)
		defer file.Close()
		decoder := gob.NewDecoder(file)
		decoder.Decode(&r)
		for k, v := range r {
			res[k] = v
		}
		os.Remove(fn)
	}
	return res

}
