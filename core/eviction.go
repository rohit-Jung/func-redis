package core

import (
	"time"

	"github.com/rohit-Jung/func-redis/config"
)

// gives the 24 bit masked current time
func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF
}

// getIdleTime gives the idle time for a key
func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	// c < LAT - c - LAT
	if c < lastAccessedAt {
		return c - lastAccessedAt
	}

	// MAX - LAT + Clock
	return (0x00FFFFFF - lastAccessedAt) + c
}

func evictFirst() {
	// the iteration is random on hashmaps
	for k := range store {
		delete(store, k)
		return
	}
}

func evictAllKeysRandom() {
	evictCount := int(config.EvictionRatio * float64(config.KeysLimit))

	for key := range store {
		Delete(key)
		evictCount--
		if evictCount == 0 {
			break
		}
	}
}

func populateSamplePool() {
	sampleKeySize := 5
	for k, v := range store {
		ePool.Push(k, v.lastAccessedAt)
		sampleKeySize--
		if sampleKeySize == 0 {
			break
		}
	}
}

// populates the pool and evicts according to the eviction ratio
// TODO: do not populate on each call
func evictAllKeysLru() {
	populateSamplePool()
	evictionCount := int(config.EvictionRatio * float64(config.KeysLimit))
	for {
		item := ePool.Pop()
		if item == nil {
			break
		}

		// delete it from store too;
		Delete(item.key)
		if evictionCount == 0 {
			break
		}

		evictionCount--
	}

}

func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "all-keys-random":
		evictAllKeysRandom()
	case "all-keys-lru":
		evictAllKeysLru()
	}
}
