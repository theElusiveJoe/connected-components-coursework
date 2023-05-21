package distribution

import (
	"connectedComponents/fastSV"
	"connectedComponents/utils"
	"fmt"
	"sort"
)

type multiEdge struct {
	v1 uint32
	v2 uint32
}

type multiComponent struct {
	hashes []uint32
	weight uint32
}

type Distributor struct {
	multiEdges  map[multiEdge]struct{}
	hashNum     uint32
	nodesWeight []uint32
}

func (dist *Distributor) H(node uint32) uint32 {
	return node % dist.hashNum
}

func (dist *Distributor) FindDistributionFromFile(filename string, numSlaves uint32, hashNum uint32) []uint32 {
	_, _, _, edges1, edges2 := utils.GetEdges(filename)
	return dist.FindDistributionFromEdges(edges1, edges2, numSlaves, hashNum)
}

func (dist *Distributor) FindDistributionFromEdges(edges1 []uint32, edges2 []uint32, numSlaves uint32, hashNum uint32) []uint32 {
	dist.multiEdges = make(map[multiEdge]struct{})
	dist.hashNum = hashNum
	dist.nodesWeight = make([]uint32, hashNum)
	fmt.Printf("-> {dist}: numslaves is %d, hashnum is %d\n", numSlaves, hashNum)

	// создаем мультиграф
	for i := 0; i < len(edges1); i++ {
		dist.addEdge(edges1[i], edges2[i])
	}
	fmt.Printf("-> {dist}: detected %d multiedges\n", len(dist.multiEdges))

	// находим связные компоненты мультиграфа
	_, multiEdges1, multiEdges2 := dist.convertToRegularRepresentation()
	fmt.Printf("-> {dist}: detected %d multinodes\n", hashNum)
	starForest := fastSV.FastSVCCSearchAdapter(hashNum, multiEdges1, multiEdges2)
	multiComponents := utils.StarForestToComponents(starForest)
	fmt.Printf("-> {dist}: found %d multicomponents\n", len(multiComponents))
	// fmt.Println(multiComponents)
	// находим веса связных компонент
	componentsWeight := make([]uint32, len(multiComponents))
	for i, _ := range multiComponents {
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

func (dist *Distributor) addEdge(v1 uint32, v2 uint32) {
	h1, h2 := dist.H(v1), dist.H(v2)
	dist.nodesWeight[h1]++
	dist.nodesWeight[h2]++
	if h2 < h1 {
		h1, h2 = h2, h1
	} else if h1 == h2 {
		return
	}
	newMEdge1 := multiEdge{h1, h2}
	if _, ok := dist.multiEdges[newMEdge1]; ok {
		return
	}
	dist.multiEdges[multiEdge{h1, h2}] = struct{}{}
}

func (dist *Distributor) convertToRegularRepresentation() (uint32, []uint32, []uint32) {
	edges1, edges2 := make([]uint32, len(dist.multiEdges)), make([]uint32, len(dist.multiEdges))
	i := 0
	for multiEdge, _ := range dist.multiEdges {
		edges1[i], edges2[i] = multiEdge.v1, multiEdge.v2
		i++
	}
	return dist.hashNum, edges1, edges2
}
