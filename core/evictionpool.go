package core

import "sort"

type PoolItem struct {
	key            string
	lastAccessedAt uint32
}

type ByIdleTime []*PoolItem

type EvictionPool struct {
	ePool  []*PoolItem
	keyset map[string]*PoolItem // to track the Poolitem for key
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		ePool:  make([]*PoolItem, size),
		keyset: make(map[string]*PoolItem),
	}
}

var ePool = newEvictionPool(0)
var ePoolSizeLimit = 10

// for sorting purposes
func (i ByIdleTime) Len() int {
	return len(i)
}

func (i ByIdleTime) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

// the idle time of both should be compared
func (i ByIdleTime) Less(a, b int) bool {
	return getIdleTime(i[a].lastAccessedAt) < getIdleTime(i[b].lastAccessedAt)
}

// PUSH
// if already exist return; check ePool size; append and sort
// if top lastAccessedAt is greater then the key's then
func (e *EvictionPool) Push(key string, lastAccessedAt uint32) {
	_, ok := ePool.keyset[key]
	if ok {
		return
	}

	ePoolItem := &PoolItem{
		key:            key,
		lastAccessedAt: lastAccessedAt,
	}

	if len(e.ePool) < ePoolSizeLimit {
		e.keyset[key] = ePoolItem
		e.ePool = append(e.ePool, ePoolItem)

		// TODO: sort it, this is extremely inefficient try to optimize it
		sort.Sort(ByIdleTime(e.ePool))
	} else if lastAccessedAt > e.ePool[0].lastAccessedAt {
		e.ePool = e.ePool[1:]
		e.keyset[key] = ePoolItem
		e.ePool = append(e.ePool, ePoolItem)
	}
}

// Pop the first one and return the item
// also delete it from keyset
func (e *EvictionPool) Pop() *PoolItem {
	if len(e.ePool) < 0 {
		return nil
	}

	item := e.ePool[0]
	e.ePool = e.ePool[1:]
	delete(e.keyset, item.key)
	return item
}
