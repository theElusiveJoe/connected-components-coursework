package misc

import (
	"connectedComponents/src/distribution"
	"connectedComponents/src/utils/ioEdges"
	"fmt"
	"reflect"
	"sort"
)

func RWtest() {
	name := "tests/biggraphs/big-1e7"
	csvName := name + ".csv"
	jsonName := name + ".json"

	g := ioEdges.LoadGraph(csvName)

	ioEdges.SaveGraph(g, jsonName)
	g2 := ioEdges.LoadGraph(jsonName)
	fmt.Println(reflect.DeepEqual(g2, g))
}

func DistTest() {
	name := "tests/biggraphs/big-1e6.json"

	gi := ioEdges.LoadGraph(name).ToIterator()

	distribution.FindDistribution(gi, 10, 1000000)
	// fmt.Println(res)
}

func SortTest() {
	mc := []int{4, 6, 2, 6, 9, 3, 1, 4, 7, 9, 2, 5, 5, 542, 8, 8, 99, 2}
	fmt.Println(mc)

	sort.Slice(mc, func(i, j int) bool {
		if mc[i] > mc[j] {
			return true
		}
		return false
	})

	fmt.Println(mc)
}
