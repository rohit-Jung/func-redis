package core

import (
	"time"

	"github.com/rohit-Jung/func-redis/config"
)

var store map[string]*Obj
var expires map[*Obj]uint32

func init() {
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint32)
}

// sets expiry; requires duration in Ms
func setExpiry(obj *Obj, durationMs uint32) {
	expires[obj] = uint32(time.Now().UnixMilli()) + durationMs
}

func NewObject(v any, durationMs int64, oType uint8, oEncoding uint8) *Obj {
	obj := &Obj{
		TypeEncoding:   oType | oEncoding,
		Value:          v,
		lastAccessedAt: getCurrentClock(),
	}

	if durationMs > 0 {
		setExpiry(obj, uint32(durationMs))
	}

	return obj
}

func Put(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}

	obj.lastAccessedAt = getCurrentClock()
	store[k] = obj
	IncrementDbStat(0, "keys")
}

func Get(k string) *Obj {
	// Active Delete
	obj, ok := store[k]
	// fault: if it exist then only you can access lastAccessedAt
	if ok {
		if hasExpired(obj) {
			Delete(k)
			return nil
		}

		obj.lastAccessedAt = getCurrentClock()
	}

	return obj
}

func Delete(k string) bool {
	if obj, ok := store[k]; ok {
		delete(store, k)
		delete(expires, obj)
		DecrementDbStat(0, "keys")
		return true
	}

	return false
}
