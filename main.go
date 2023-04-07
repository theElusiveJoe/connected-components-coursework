package main

import (
	"fmt"

	"runtime"

	"github.com/google/uuid"
)

type TrieNode struct {
	parent   *TrieNode
	counter  int32
	children map[byte]*TrieNode
}

type TrieNodeer interface {
	add_string(string)
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		parent:   nil,
		counter:  0,
		children: make(map[byte]*TrieNode),
	}
}

func (receiver *TrieNode) print_graph() {
	fmt.Println("digraph {")
	receiver.print_graph_loop()
	fmt.Println("}")
}

func (receiver *TrieNode) print_graph_loop() {
	fmt.Printf("	\"%p\" [label=%d]", receiver, receiver.counter)
	for val, child := range receiver.children {
		fmt.Printf("	\"%p\" -> \"%p\"\n [label=\"%s\"]", receiver, child, string(val))
		child.print_graph_loop()
	}
}

// func (receiver *TrieNode) get_size_in_bytes() uint64 {
// 	// receiver.print_graph_loop()
// 	// var sum uint64 = 0
// 	// for _, x := range receiver.children {
// 	// 	sum += x.get_size_in_bytes()
// 	// }
// 	return uint64(unsafe.Sizeof(int32(0))) +
// }

func (receiver *TrieNode) add_string(s string) int {
	// если строка пустая, то мы успешно завершились
	if len(s) == 0 {
		return 0
	}

	sym, rest := s[0], s[1:]

	child_node, already_exists := receiver.children[sym]

	if !already_exists {
		child_node = newTrieNode()
		receiver.children[sym] = child_node
	}
	receiver.counter++

	return child_node.add_string(rest)
}

func main() {
	root := newTrieNode()
	for i := 0; i < 999999; i++ {
		root.add_string(uuid.NewString())
	}
	root.add_string("aboba")
	root.add_string("abobus")
	root.add_string("aboob")

	// root.print_graph()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println("total:", m.TotalAlloc)
}
