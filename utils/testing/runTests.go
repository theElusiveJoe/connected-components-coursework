package testing

import (
	"connectedComponents/algos"

	"fmt"
	"os"
	"path"
)

func RunTestsInFolder(
	folder string,
	foo1 func(*algos.RunConfig) map[uint32]uint32,
	foo2 func(*algos.RunConfig) map[uint32]uint32,
) {
	files := make([]string, 0)
	fs, _ := os.ReadDir(folder)
	for _, f := range fs {
		files = append(files, path.Join(folder, f.Name()))
	}
	fmt.Println("TESTS:", files)

	hns := []uint32{1, 2, 3, 10, 20, 40, 100, 1000, 10000, 1000000}
	rts := []int{1, 2, 3, 5, 10, 20}
	sls := []int{1, 2, 3, 5, 10, 20}
	total := len(hns) * len(rts) * len(sls)
	completed := 0
	for _, hashNum := range hns {
		for _, routersNum := range rts {
			for _, slavesNum := range sls {
				conf := algos.RunConfig{
					TestFile:   "",
					ResultDir:  "outputs/",
					RoutersNum: routersNum,
					Slavesnum:  slavesNum,
					HashNum:    int(hashNum),
					Id:         "",
				}
				res, resmap := CompareManyTests(foo1, foo2, &conf, files)
				fmt.Println(res)
				if !res {
					fmt.Println(resmap)
					panic("ТЕСТ НЕ ПРОШЕЛ!")
				}
				completed++
				fmt.Printf("COMPLETED %d of %d test cycles\n\n\n", completed, total)
			}
		}
	}
}
