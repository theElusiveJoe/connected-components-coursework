package utils

import (
	"encoding/csv"
	"os"
)

func GetEdgesWriter(filename string) *csv.Writer {
	file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0660)
	writer := csv.NewWriter(file)
	return writer
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
