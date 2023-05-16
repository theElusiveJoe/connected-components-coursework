package utils

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
)

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func CompareOneTest(
	foo1 func(string) []uint32,
	foo2 func(string) []uint32,
	filename string,
) bool {
	res1 := foo1(filename)
	res2 := foo2(filename)

	fmt.Println(getFunctionName(foo1), res1)
	fmt.Println(getFunctionName(foo2), res2)
	return reflect.DeepEqual(res1, res2)
}

func CompareManyTests(
	foo1 func(string) []uint32,
	foo2 func(string) []uint32,
	dir string,
) (bool, map[string]bool) {
	tests, _ := os.ReadDir(dir)
	resMap := make(map[string]bool, 0)
	resTotal := true
	for i, fn := range tests {
		fmt.Printf("----- %d / %d -----", i, len(tests))
		filename := fn.Name()
		fmt.Println(dir + filename)
		res := CompareOneTest(foo1, foo2, dir+filename)
		resMap[filename] = res
		if res {
			fmt.Println("PASSED")
		} else {
			fmt.Println("NOT PASSED")
		}
		resTotal = resTotal && res
	}
	if resTotal {
		fmt.Println("\nALL TESTS PASSED!!!")
	}
	return resTotal, resMap
}
