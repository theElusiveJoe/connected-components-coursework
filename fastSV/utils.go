package fastSV

import (
	"encoding/csv"
	"log"
	"os"
)

// принимает имя csv-файла, возвращает массив его строк
func getEdgesReader(filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	return records
}

// принимает массив пар узлов,
// возвращает количество уникальных улов, словарь-индексацию и реерс этого словаря
func createIndexAndPairsLists(rows [][]string) (uint32, map[string]uint32, map[uint32]string, []uint32, []uint32) {
	mapa := make(map[string]uint32)
	mapaRev := make(map[uint32]string)
	edges := [][]uint32{make([]uint32, 0), make([]uint32, 0)}

	n := uint32(0)
	for _, row := range rows {
		for i := 0; i <= 1; i++ {
			if _, err := mapa[row[i]]; !err {
				mapa[row[i]] = n
				mapaRev[n] = row[i]
				n++
			}
			edges[i] = append(edges[i], mapa[row[i]])
		}
	}
	return n, mapa, mapaRev, edges[0], edges[1]
}

func GetEdges(filename string) (uint32, map[string]uint32, map[uint32]string, []uint32, []uint32) {
	return createIndexAndPairsLists(getEdgesReader(filename))
}

func StarForestToComponents(startsForest []uint32) [][]uint32 {
	representorToNum := make(map[uint32]uint32)
	n := uint32(0)
	for _, x := range startsForest {
		if _, err := representorToNum[x]; !err {
			representorToNum[x] = n
			n++
		}
	}

	components := make([][]uint32, n)
	var componentIndex uint32
	for nodeIndex, representor := range startsForest {
		componentIndex = representorToNum[representor]
		components[componentIndex] = append(components[componentIndex], uint32(nodeIndex))
	}
	return components
}
