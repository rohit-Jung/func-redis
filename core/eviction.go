package core

func evictFirst() {
	// the iteration is random on hashmaps
	for k := range store {
		delete(store, k)
		return
	}
}

func evict() {
	evictFirst()
}
