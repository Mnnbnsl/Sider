package core

import "time"

func sampleKeys(sampleSize int) []string {
    keys := make([]string, 0, sampleSize)

    for key, obj := range store {
        if obj.ExpiresAt != -1 {
            keys = append(keys, key)
        }

        if len(keys) == sampleSize {
            break
        }
    }

    return keys
}

func Cleanup() {
    const sampleSize = 20

    for {
        keys := sampleKeys(sampleSize)

        if len(keys) == 0 {
            return
        }

        expiredKeys := 0
        now := time.Now().UnixMilli()

        for _, key := range keys {
            obj, exists := store[key]

            if !exists {
                continue
            }

            if obj.ExpiresAt <= now {
                delete(store, key)
                expiredKeys++
            }
        }

        if expiredKeys*100 <= len(keys)*25 {
            return
        }
    }
}
