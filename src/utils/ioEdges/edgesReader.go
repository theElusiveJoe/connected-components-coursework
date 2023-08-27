package ioEdges

import (
	"connectedComponents/src/graph"
	"log"
)

func LoadGraph(filename string) *graph.Graph {
	// fmt.Printf("-> {edges reader}: opening \"%s\"\n", filename)

	if filename[len(filename)-4:] == "json" {
		return readJsonGraph(filename)
	}

	if filename[len(filename)-3:] == "csv" {
		return readCsvGraph(filename)
	}

	log.Fatal("UNKNOW FILE FORMAT:", filename)
	return nil
}
