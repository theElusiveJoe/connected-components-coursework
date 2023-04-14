package fastSV

import (
	"fmt"
	"io/ioutil"
)

func checkFastSV(filename string) bool {
	nodesNum, _, _, edges1, edges2 := GetPairs(filename)

	// fmt.Printf("Graph parsed\ngot %d nodes:\n", nodesNum)
	// fmt.Println(edges1)
	// fmt.Println(edges2)

	// phl()
	baseRes := BasicCCSearch(nodesNum, edges1, edges2)
	// fmt.Println("Basic result\n", baseRes)
	// phl()

	fastSVRes := FastSVCCSearch(nodesNum, edges1, edges2)
	// fmt.Println("My result\n", fastSVRes)
	// phl()

	for i := 0; i < len(baseRes); i++ {
		for j := 0; j < len(baseRes[i]); j++ {
			if baseRes[i][j] != fastSVRes[i][j] {
				// fmt.Println("fastSV реализован неверно")
				return false
			}
		}
	}
	fmt.Println(fastSVRes)
	return true
}

func TestAll(dir string) (bool, map[string]bool) {
	tests, _ := ioutil.ReadDir(dir)

	res := make(map[string]bool, 0)
	resTotal := true
	for _, fn := range tests {
		println("-----")
		filename := fn.Name()
		r := checkFastSV(dir + filename)
		res[filename] = r
		fmt.Println(filename, "<-", r)
		resTotal = resTotal && r
	}
	if resTotal {
		fmt.Println("\nALL TESTS PASSED!!!\n")
	}
	return resTotal, res
}
