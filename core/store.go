package core

import (
	"time"

	"github.com/rohit-Jung/func-redis/config"
)

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObject(v any, durationMs int64, oType uint8, oEncoding uint8) *Obj {
	var expiry int64 = -1
	if durationMs > 0 {
		expiry = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		TypeEncoding: oType | oEncoding,
		Value:        v,
		ExpiresAt:    int64(expiry),
	}
}

func Put(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}

	store[k] = obj
}

func Get(k string) *Obj {
	// Active Delete
	if obj, ok := store[k]; ok &&
		obj.ExpiresAt != -1 &&
		time.Now().UnixMilli() > obj.ExpiresAt {
		Delete(k)
	}

	return store[k]
}

func Delete(k string) bool {
	if _, ok := store[k]; ok {
		delete(store, k)
		return true
	}

	return false
}
