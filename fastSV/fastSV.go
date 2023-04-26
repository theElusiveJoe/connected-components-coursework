package fastSV

import "fmt"
import "connectedComponents/utils"

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

	n := 0

	for changed {
		changed = false
		n++
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

		for i := 0; i < edgesNum; i++ {
			v, u = edges1[i], edges2[i]

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

		for i := 0; i < edgesNum; i++ {
			v, u = edges1[i], edges2[i]

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
	fmt.Println(n, "iterations")
	return f_next
}

func FastSVCCSearch(nodesNum uint32, edges1 []uint32, edges2 []uint32) [][]uint32 {
	return utils.StarForestToComponents(
		fastSV(uint32(nodesNum), edges1, edges2),
	)
}
