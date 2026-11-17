package store

import (
	"math/rand"

	"github.com/mnnbnsl/sider/config"
)

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
