package main

import (
	"connectedComponents/prefixTree"
)

func main() {
	prefixTree.TestMap("someData/500k.csv")
	// display_stats()
	// fmt.Println(len(records), "nodes")
	// fmt.Println(unsafe.Sizeof(root.children))
}
