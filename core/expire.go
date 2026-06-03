package core

import (
	"fmt"
	"time"
)

// Simple Sampling
func expireSample() float32 {
	limit := 20
	expiredCount := 0

	for k, obj := range store {
		if obj.ExpiresAt != -1 {
			limit -= 1
			if time.Now().UnixMilli() > obj.ExpiresAt {
				Delete(k)
				expiredCount++
			}
		}

		if limit == 0 {
			break
		}
	}

	return float32(expiredCount) / float32(20.0)
}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()
		if frac < 0.25 {
			break
		}
	}

	fmt.Println("Total keys: ", len(store))
}
