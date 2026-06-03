package core

import (
	"fmt"
	"time"
)

type Obj struct {
	Value     any
	ExpiresAt int64
}

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObject(v any, durationMs int64) *Obj {
	var expiry int64 = -1
	if durationMs > 0 {
		expiry = time.Now().UnixMilli() + durationMs
	}

	fmt.Print("Setting expiry", expiry, durationMs)

	return &Obj{
		Value:     v,
		ExpiresAt: int64(expiry),
	}
}

func Put(k string, obj *Obj) {
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
