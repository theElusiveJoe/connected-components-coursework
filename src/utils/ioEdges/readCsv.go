package ioEdges

import (
	"connectedComponents/src/graph"
	"encoding/csv"
	"log"
	"os"
)

func getRecords(filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	return records
}

func createIndexAndPairsLists(rows [][]string) *graph.Graph {
	mapa := make(map[string]uint32)
	edges := [][]uint32{make([]uint32, 0), make([]uint32, 0)}
	n := uint32(0)

	for _, row := range rows {
		for i := 0; i <= 1; i++ {
			if _, ok := mapa[row[i]]; !ok {
				mapa[row[i]] = n
				n++
			}
			edges[i] = append(edges[i], mapa[row[i]])
		}
	}

	// fmt.Printf("-> {csv reader}: detected %d nodes and %d edges\n", n, len(edges[0]))

	g := graph.Graph{
		NodesNum: n,
		Edges1:   edges[0],
		Edges2:   edges[1],
		Mapa:     map[string]uint32{}, //mapa,
	}
	return &g
}

func readCsvGraph(filename string) *graph.Graph {
	rows := getRecords(filename)

	return createIndexAndPairsLists(rows)
}
