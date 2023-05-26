package ioEdges 

import (
    "encoding/json"
    "io/ioutil"
	"log"
	"connectedComponents/graph"
)



func readJsonGraph(filename string) *graph.Graph {
	content, err := ioutil.ReadFile(filename)
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }
 
    var g graph.Graph
    err = json.Unmarshal(content, &g)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

	return &g
}

func SaveGraph(g *graph.Graph, filename string){
	gbyte, err := json.Marshal(&g)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filename, gbyte, 0644)
	if err != nil {
		log.Fatal(err)
	}
}