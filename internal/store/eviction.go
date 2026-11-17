package store

import (
	"math/rand"
	"time"

	"github.com/mnnbnsl/sider/config"
)

const lruSampleSize = 5

func lruClock() uint32 {
	return uint32(time.Now().Unix())
}

func lruIdleTime(now, lastAccess uint32) uint32 {
	return now - lastAccess
}

// Evict removes enough keys to bring the store back under its configured limit.
// The selected policy determines which keys are eligible for removal.
func Evict() []string {
	if config.KeysLimit < 0 || len(store) <= config.KeysLimit {
		return nil
	}

	evictionCount := int(float64(len(store)) * config.EvictionRatio)
	if evictionCount < 1 {
		evictionCount = 1
	}
	if evictionCount > len(store)-config.KeysLimit {
		evictionCount = len(store) - config.KeysLimit
	}

	switch config.EvictionStrategy {
	case "allkeys-random":
		return evictAllKeysRandom(evictionCount)
	case "allkeys-lru":
		return evictAllKeysLRU(evictionCount)
	case "noeviction":
		return nil
	default:
		return nil
	}
}

func evictAllKeysRandom(count int) []string {
	keys := make([]string, 0, len(store))
	for key := range store {
		keys = append(keys, key)
	}

	evicted := make([]string, 0, count)
	for len(evicted) < count && len(keys) > 0 {
		index := rand.Intn(len(keys))
		key := keys[index]
		delete(store, key)
		evicted = append(evicted, key)
		keys[index] = keys[len(keys)-1]
		keys = keys[:len(keys)-1]
	}
	return evicted
}

func evictAllKeysLRU(count int) []string {
	evicted := make([]string, 0, count)
	now := lruClock()

	for len(evicted) < count && len(store) > 0 {
		keys := make([]string, 0, len(store))
		for key := range store {
			keys = append(keys, key)
		}

		sampleSize := lruSampleSize
		if sampleSize > len(keys) {
			sampleSize = len(keys)
		}

		for i := 0; i < sampleSize; i++ {
			index := i + rand.Intn(len(keys)-i)
			keys[i], keys[index] = keys[index], keys[i]
		}

		var oldestKey string
		var oldestIdle uint32
		var haveOldest bool
		for _, key := range keys[:sampleSize] {
			obj := store[key]
			idle := lruIdleTime(now, obj.LRU)
			if !haveOldest || idle > oldestIdle {
				oldestKey = key
				oldestIdle = idle
				haveOldest = true
			}
		}

		delete(store, oldestKey)
		evicted = append(evicted, oldestKey)
	}

	return evicted
}
