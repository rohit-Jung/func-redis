package core

var KeymapsStat [4]map[string]int

func UpdateDBStat(num int, metric string, value int) {
	if KeymapsStat[num] == nil {
		KeymapsStat[num] = make(map[string]int)
	}

	KeymapsStat[num][metric] = value
}

func IncrementDbStat(num int, metric string) {
	if KeymapsStat[num] == nil {
		KeymapsStat[num] = make(map[string]int)
	}

	KeymapsStat[num][metric]++
}

func DecrementDbStat(num int, metric string) {
	KeymapsStat[num][metric]--
}
