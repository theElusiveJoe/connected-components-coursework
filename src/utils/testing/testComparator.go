package testing

import (
	"connectedComponents/src/algos"
	"fmt"
	"reflect"
	"runtime"
)

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func CompareOneTest(
	foo1 func(*algos.RunConfig) map[uint32]uint32,
	foo2 func(*algos.RunConfig) map[uint32]uint32,
	conf *algos.RunConfig,
) bool {
	res1 := foo1(conf)
	res2 := foo2(conf)
	// fmt.Println(getFunctionName(foo1), res1)
	// fmt.Println(getFunctionName(foo2), res2)
	fmt.Println(res1)
	return reflect.DeepEqual(res1, res2)
}

func CompareManyTests(
	foo1 func(*algos.RunConfig) map[uint32]uint32,
	foo2 func(*algos.RunConfig) map[uint32]uint32,
	conf *algos.RunConfig,
	test_files []string,
) (bool, map[string]bool) {
	resMap := make(map[string]bool, 0)
	resTotal := true

	for i, fn := range test_files {
		fmt.Printf("----- %d / %d -----", i, len(test_files))
		fmt.Println(fn)
		conf.TestFile = fn
		res := CompareOneTest(foo1, foo2, conf)
		resMap[fn] = res
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
