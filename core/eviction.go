package core

import "github.com/rohit-Jung/func-redis/config"

func evictFirst() {
	// the iteration is random on hashmaps
	for k := range store {
		delete(store, k)
		return
	}
}

func evictAllKeysRandom() {
	evictCount := int(config.EvictionRatio * float64(config.KeysLimit))

	for key, _ := range store {
		Delete(key)
		evictCount--
		if evictCount == 0 {
			break
		}
	}
}

func evict() {
	switch config.EvictionStrategy {
	case "all-keys-random":
		evictAllKeysRandom()
	case "simple-first":
		evictFirst()
	}
}
