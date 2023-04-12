package prefixTree

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

type TrieNode struct {
	parent   *TrieNode
	counter  int32
	children map[byte]*TrieNode
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

func display_stats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("total: megabytes %f (%d)\n", float64(m.TotalAlloc)/1024/1024, m.TotalAlloc)
}

func open_csv(filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	return records
}

func TestMap(filename string) {
	debug.SetGCPercent(-1)

	records := open_csv(filename)

	mapa := make(map[string]uint32)
	n := 0
	for _, record := range records {
		mapa[record[0]] = uint32(n)
		n += 1

	}
	display_stats()
}

func TestPrefixTree(filename string) {
	debug.SetGCPercent(-1)

	records := open_csv(filename)

	root := newTrieNode()
	l := float64(len(records))
	for i, record := range records {
		i := float64(i)
		root.add_string(record[0])
		fmt.Printf("%f\n", (i/l)*100)
	}

	display_stats()
}
