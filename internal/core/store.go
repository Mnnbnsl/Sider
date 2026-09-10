package core 

import (
	"time"
)

var store map[string]*Obj

type Obj struct {
	Value interface{}
	ExpiresAt int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationMs int64) *Obj {
	var expiresAt int64 = -1
	if durationMs >= 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj {
		Value : value,
		ExpiresAt : expiresAt,
	}
}

func Put(key string, obj *Obj) {
	store[key] = obj
}

func Get(key string) *Obj {
	obj, exists := store[key]

	if !exists {
		return nil 
	}

	// object expired (TTL passed)
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		delete(store, key)
		return nil
	}

	return obj
}	

func Del(args []string) int {
	deleted := 0

	for _, key := range args {
		if _, exists := store[key]; exists {
			delete(store, key)
			deleted++
		}
	}

	return deleted
}

func Expire(key string, durationMs int64) int {
	obj, exists := store[key]
	if !exists {
		return 0
	}  
	if durationMs >= 0 {
		obj.ExpiresAt = time.Now().UnixMilli() + durationMs
	}
	return 1
}
