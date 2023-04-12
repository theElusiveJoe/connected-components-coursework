package fastSV

import (
	"encoding/csv"
	"fmt"
	"os"
)

func fastSV(nodesNum uint32, edges1 []uint32, edges2 []uint32) []uint32 {

	// Step 0
	// заполняем лес пнями
	f := make([]uint32, nodesNum)
	f_next := make([]uint32, nodesNum)
	for i := 0; i < len(f); i++ {
		f[i] = uint32(i)
	}
	copy(f_next, f)

	changed := true
	var v, u, pu, gpv, gpu uint32
	edgesNum := len(edges1)

	for changed {
		changed = false

		// STEP 1
		// Stochastic hooking
		for i := 0; i < edgesNum; i++ {
			u, v = edges1[i], edges2[i]

			pu = f[u]
			gpv = f[f[v]]
			gpu = f_next[pu]

			if gpv < gpu {
				f_next[pu] = gpv
				changed = true
			}
		}

		// STEP 2
		// Agressive hooking
		for i := 0; i < edgesNum; i++ {
			u, v = edges1[i], edges2[i]

			pu = f_next[u]
			gpv = f[f[v]]

			if gpv < pu {
				f_next[u] = gpv
				changed = true
			}
		}

		// STEP 3
		// Shortcutting
		for i := uint32(0); i < nodesNum; i++ {
			u = i

			gpu = f[f[u]]
			pu = f_next[u]

			if gpu < pu {
				f_next[u] = gpu
				changed = true
			}
		}

		copy(f, f_next)
	}

	return f_next
}

func getPairsReader(filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	return records
}

func enumerateNodes(rows [][]string) (uint32, map[string]uint32) {
	mapa := make(map[string]uint32)
	n := uint32(0)
	for _, row := range rows {
		if _, err := mapa[row[0]]; !err {
			mapa[row[0]] = n
			n++
		}
		if _, err := mapa[row[1]]; !err {
			mapa[row[1]] = n
			n++
		}
	}
	return n, mapa
}

func FindCC(filename string) map[uint32]map[uint32]struct{} {
	// читаем две колонки узлов из файла
	pairs := getPairsReader(filename)

	// нумеруем излы
	nodesNum, mapa := enumerateNodes(pairs)

	// создаем списки узлов
	edges1, edges2 := make([]uint32, nodesNum), make([]uint32, nodesNum)
	for i, row := range pairs {
		edges1[i], edges2[i] = mapa[row[0]], mapa[row[0]]
	}

	// ищем компонеты
	startsForest := fastSV(uint32(nodesNum), edges1, edges2)

	// распределяем узлы по множествам
	// ccs := make(map[string]struct{})
	sets := make(map[uint32]map[uint32]struct{})
	exists := struct{}{}
	for i, compNum := range startsForest {
		if _, err := sets[uint32(compNum)]; !err {
			sets[uint32(compNum)] = make(map[uint32]struct{})
		}
		sets[uint32(compNum)][uint32(i)] = exists
	}

	return sets
}
