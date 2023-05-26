package distribution

import (
	"connectedComponents/utils"
	"connectedComponents/graph"
	"connectedComponents/algos/basic"
	"fmt"
	"sort"
)

func (dist *Distributor) FindDistribution(g *graph.Graph, numSlaves uint32, hashNum uint32) []uint32 {
	// инициализируем объект распределителя
	dist.multiEdges = make(map[multiEdge]struct{}) 
	dist.hashNum = hashNum		
	dist.nodesWeight = make([]uint32, hashNum)	
	fmt.Printf("-> {dist}: numslaves is %d, hashnum is %d\n", numSlaves, hashNum)

	// задаем мультиграф
	for i := uint32(0); i < g.Len(); i++ {
		dist.addEdge(g.GetEdge(i))
	}
	fmt.Printf("-> {dist}: detected %d multiedges\n", len(dist.multiEdges))

	// находим связные компоненты мультиграфа
	multiG := dist.toGraph()
	fmt.Printf("-> {dist}: detected %d multinodes\n", hashNum)
	starForest := basic.BasicCCSearch(multiG)
	multiComponents := utils.StarForestToComponents(starForest)
	// fmt.Println(multiComponents)
	fmt.Printf("-> {dist}: found %d multicomponents\n", len(multiComponents))
	
	// находим веса связных компонент
	componentsWeight := make([]uint32, len(multiComponents))
	for i := range multiComponents {
		for _, multinodeNum := range multiComponents[i] {
			componentsWeight[i] += dist.nodesWeight[multinodeNum]
		}
	}
	// и сортируем компоненты по весам
	sort.Slice(multiComponents, func(i, j int) bool {
		if componentsWeight[i] > componentsWeight[j] {
			componentsWeight[i], componentsWeight[j] = componentsWeight[j], componentsWeight[i]
			return true
		}
		return false
	})

	// создае массив распределений хешей по слейвам
	slavesHashes := make([][]uint32, hashNum)
	for j := uint32(0); j < uint32(numSlaves); j++ {
		slavesHashes[j] = make([]uint32, 0)
	}
	// "змейкой" распределяем мультикомпоненты по слейвам
	i := 0
	for i < len(multiComponents) {
		for j := uint32(0); j < uint32(numSlaves) && i < len(multiComponents); j++ {
			slavesHashes[j] = append(slavesHashes[j], multiComponents[i]...)
			i++
		}
		for j := numSlaves - 1; j >= uint32(0) && j < 4294967243 && i < len(multiComponents); j-- {
			slavesHashes[j] = append(slavesHashes[j], multiComponents[i]...)
			i++
		}
	}

	// возвращаем массив ret[h] = b, где h - хэш, b - номер слейва
	ret := make([]uint32, dist.hashNum)
	for slaveNum, hashes := range slavesHashes {
		for _, hash := range hashes {
			ret[hash] = uint32(slaveNum) + 1
		}
	}

	return ret
}
