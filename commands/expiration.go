package commands

import (
	"strconv"
	"time"

	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
)

/*
Set key expiration (Time to Kill)

Description:

	Sets a lifespan for the specified key. The key will be automatically deleted from the cache after the given number of seconds.

Example:

  - Pattern: TTL session:123 360

  - Explanation: Sets the key "session:123" to expire in 360 seconds (6 minutes)
*/
func TTK(c *storage.Cache, k, v string) protocol.Response {
	result := protocol.Response{
		Success: true,
	}

	t, err := strconv.Atoi(v)
	if err != nil || t < -1 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrInvalidTTL.Error(),
		}
	}

	var expiresAt *time.Time
	if t != -1 {
		et := time.Now().Add(time.Duration(t) * time.Second)
		expiresAt = &et
	}

	c.WithLock(func() {
		if k == "*" {
			for key := range c.GetUnsafeData() {
				c.SetPartialUnsafe(key, storage.CacheDataUpdate{
					ExpiresAt: expiresAt,
				})
			}
			result.Output = protocol.Success()
			return
		}

		_, exists := c.GetUnsafe(k)
		if !exists {
			result = protocol.Response{
				Success: false,
				Output:  protocol.Failure(),
			}
			return
		}

		c.SetPartialUnsafe(k, storage.CacheDataUpdate{
			ExpiresAt: expiresAt,
		})

		result.Output = protocol.Success()
	})

	return result
}

/*
Time left before data is deleted (Time to Live)

Description:

	Returns the remaining time (in seconds) before the specified key is automatically deleted from the cache.

Example:

  - Pattern: TTL KEY_1

  - Result: 120

  - Explanation: 120 seconds left before the key expires

Notes:
  - If the key exists but has no expiration, returns -1
  - If the key doesn`t exist, returns -2
*/
func TTL(c *storage.Cache, k string) protocol.Response {
	result := protocol.Response{
		Success: true,
	}

	c.WithRWLock(func() {
		if dataCache, exists := c.GetSafe(k); exists {
			if dataCache.ExpiresAt == nil {
				result.Output = "-1"
				return
			}
			result.Output = protocol.Number(int(time.Until(*dataCache.ExpiresAt).Seconds()))
		} else {
			result.Output = "-2"
			return
		}
	})

	return result
}
