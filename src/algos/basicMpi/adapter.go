package basicMpi

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
			"-n", fmt.Sprintf("%d", conf.Slavesnum+1),
			"-oversubscribe",

			"main",
			"-mode=" + algos.MODE_MPI_LAUNCH,
			"-algo=" + algos.ALGO_MPI_BASIC,
			"-conf=" + "'" + conf.ConfigToStr() + "'",
		}...,
	)
	// перенаправляем вывод
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// запускаем
	fmt.Println("\n#### COMMAND ######\n", cmd, "\n###################")
	cmd.Start()
	defer cmd.Process.Kill()

	// создаем команду логгера памяти
	memLogger := exec.Command(
		"python",
		"tools/memStats.py",
		// можем получить PID только после запуска cmd
		fmt.Sprintf("%d", cmd.Process.Pid),
	)
	// запускаем
	memLogger.Stdout = os.Stdout
	memLogger.Stderr = os.Stderr
	memLogger.Start()
	defer memLogger.Process.Kill()
	fmt.Println("\n#### COMMAND ######\n", memLogger, "\n###################")

	// ждем, пока mpi отработает и выводим, что он написал
	if err := cmd.Wait(); err != nil {
		fmt.Println("ошибка в алгоритме MPI BASIC")
		panic(err)
	}

	// теперь можно прервать mem_logger
	if err := memLogger.Process.Kill(); err != nil {
		panic(err)
	} else {
		fmt.Println("-> логгер памяти успешно прерван")
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
