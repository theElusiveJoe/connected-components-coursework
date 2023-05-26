package ioEdges

import (
	"connectedComponents/graph"
	"encoding/csv"
	"fmt"
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

func createIndexAndPairsLists(rows [][]string, needMapa bool) *graph.Graph {
	var mapa map[string]uint32
	if needMapa{
		mapa = make(map[string]uint32)
	} else {
		mapa = map[string]uint32{}
	}

	edges := [][]uint32{make([]uint32, 0), make([]uint32, 0)}
	n := uint32(0)

	for _, row := range rows {
		for i := 0; i <= 1; i++ {
			if _, err := mapa[row[i]]; !err {
				if needMapa{
					mapa[row[i]] = n
				}
				n++
			}
			edges[i] = append(edges[i], mapa[row[i]])
		}
	}

	fmt.Printf("-> {edges reader}: detected %d nodes and %d edges\n", n, len(edges[0]))
	
	g := graph.Graph{
		n,
		edges[0],
		edges[1],
		mapa,
	}
	return &g
}

func readCsvGraph(filename string, needMapa ...bool) *graph.Graph {
	rows := getRecords(filename)
	
	nm := false
	if len(needMapa) > 0 && needMapa[0]{
		nm = true
	}

	return createIndexAndPairsLists(rows, nm)
}
