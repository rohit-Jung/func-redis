package core

import (
	"fmt"
	"time"
)

// expires; check for key and the expiry according to time
func hasExpired(obj *Obj) bool {
	exp, ok := expires[obj]
	if !ok {
		return false
	}

	return exp <= uint32(time.Now().UnixMilli())
}

func getExpiry(obj *Obj) (uint32, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

// Simple Sampling
func expireSample() float32 {
	limit := 20
	expiredCount := 0

	for k, obj := range store {
		if hasExpired(obj) {
			Delete(k)
			expiredCount++
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
