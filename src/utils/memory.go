package utils

import (
	"fmt"
	"runtime"
)

func ReadMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func CheckHeapAlloc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("allocated: %dMB (%dKB)\n", m.HeapAlloc/1024/1024, m.HeapAlloc/1024)
}
