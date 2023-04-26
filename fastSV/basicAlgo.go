package fastSV

// максимально базовый алгоритм
func basic(nodesNum uint32, edges1 []uint32, edges2 []uint32) []uint32 {
	f := make([]uint32, nodesNum)
	for i := 0; i < len(f); i++ {
		f[i] = uint32(i)
	}

	changed := true
	for changed {
		changed = false

		for i := 0; i < len(edges1); i++ {
			u, v := edges1[i], edges2[i]
			pu, pv := f[u], f[v]
			if pu < pv {
				f[v] = pu
				changed = true
			} else if pu > pv {
				f[u] = pv
				changed = true
			}
		}
	}

	return f
}

func BasicCCSearch(nodesNum uint32, edges1 []uint32, edges2 []uint32) [][]uint32 {
	return StarForestToComponents(
		basic(uint32(nodesNum), edges1, edges2),
	)
}
