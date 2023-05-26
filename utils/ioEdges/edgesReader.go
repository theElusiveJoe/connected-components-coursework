package ioEdges

import (
	"connectedComponents/graph"
	"log"
)

func GetGraph(filename string) *graph.Graph {
	if filename[len(filename)-4:] == "json" {
		return readJsonGraph(filename)
	}

	if filename[len(filename)-4:] == "csv" {
		return readCsvGraph(filename)
	}

	log.Fatal("UNKNOW FILE FORMAT:", filename)
	return nil
}
