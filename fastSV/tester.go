package fastSV

import (
	"fmt"
	"os"
)

func compareResults(baseRes [][]uint32, fastSVRes [][]uint32) bool {
	if len(baseRes) != len(fastSVRes) {
		return false
	}
	for i := 0; i < len(baseRes); i++ {
		if len(baseRes[i]) != len(fastSVRes[i]) {
			return false
		}
		for j := 0; j < len(baseRes[i]); j++ {
			if baseRes[i][j] != fastSVRes[i][j] {
				return false
			}
		}
	}
	// fmt.Println(fastSVRes)
	return true
}

func CompareOneTest(filename string) bool {
	nodesNum, _, _, edges1, edges2 := GetEdges(filename)
	baseRes := BasicCCSearch(nodesNum, edges1, edges2)
	fastSVRes := FastSVCCSearch(nodesNum, edges1, edges2)
	return compareResults(baseRes, fastSVRes)
}

func CompareManyTests(dir string) (bool, map[string]bool) {
	tests, _ := os.ReadDir(dir)

	resMap := make(map[string]bool, 0)
	resTotal := true
	for _, fn := range tests {
		println("-----")
		filename := fn.Name()
		fmt.Println(filename)
		res := CompareOneTest(dir + filename)
		resMap[filename] = res
		if res {
			fmt.Println("PASSED")
		} else {
			fmt.Println("ERROR")
		}
		resTotal = resTotal && res
	}
	if resTotal {
		fmt.Println("\nALL TESTS PASSED!!!")
	}
	return resTotal, resMap
}

func TestOne(filename string) [][]uint32 {
	nodesNum, _, _, edges1, edges2 := GetEdges(filename)
	return FastSVCCSearch(nodesNum, edges1, edges2)
}
