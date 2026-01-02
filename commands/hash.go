package commands

import (
	"cache/logger"
	"cache/storage"
	"fmt"
	"time"
)

// Store hashset in the cache
func HSET(c *storage.Cache, hN string, s []string) {
	hash := make(map[string]string)

	if len(s)%2 == 0 && len(s) != 0 {

		// if ok := removeQuotes(&s, 0, 1); !ok {
		// 	return
		// }

		c.WithLock(func() {
			if cachedData, exists := c.GetUnsafe(hN); exists {
				cachedData.Requests++

				if m, ok := cachedData.Value.(map[string]string); ok {
					for i := 0; i < len(s); i += 2 {
						m[s[i]] = s[i+1]
					}

					logger.Success("OK")
					return
				}
			}

			for i := 0; i < len(s); i += 2 {
				hash[s[i]] = s[i+1]
			}

			c.SetUnsafe(hN, &storage.CacheData{
				Value:     hash,
				Requests:  1,
				TimeStamp: time.Now(),
			})

			logger.Success("OK")
		})
	} else {
		logger.Error("Not enough values")
		return
	}
}

// Get hashset data from the cache by key
func HGET(c *storage.Cache, hN string, s []string) {
	res := make([]string, len(s))

	cachedData, exists := c.GetSafe(hN)
	if !exists {
		if len(s) == 1 {
			fmt.Println(nil)
		} else {
			res = append(res, "nil")
		}
		return
	}

	// CHECK
	cachedData.Requests++

	m, ok := cachedData.Value.(map[string]string)
	if !ok {
		logger.Error("%s isn't a hash", hN)
		for i := range res {
			res[i] = "nil"
		}

		fmt.Println(res)
	}

	for i, k := range s {
		if v, exists := m[k]; exists {
			res[i] = v
		} else {
			res[i] = "nil"
		}
	}

	fmt.Println(res)
}

// Get hashset data keys
func HKEYS(c *storage.Cache, hN string) {
	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(hN)
		if !exists {
			logger.Error("can`t find %v in memory", hN)
			return
		}

		// CHECK
		cachedData.Requests++

		m, ok := cachedData.Value.(map[string]string)
		if !ok {
			logger.Error("%s isn't a hash", hN)
			return
		}

		slice := make([]string, 0, len(m))
		for k := range m {
			slice = append(slice, k)
		}

		fmt.Println(slice)
	})
}

// Get hashset data values
func HVALUES(c *storage.Cache, hN string) {
	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(hN)
		if !exists {
			logger.Error("can`t find %v in memory", hN)
			return
		}

		// CHECK
		cachedData.Requests++

		m, ok := cachedData.Value.(map[string]string)
		if !ok {
			logger.Error("%s isn't a hash", hN)
			return
		}

		slice := make([]string, 0, len(m))
		for _, v := range m {
			slice = append(slice, v)
		}

		fmt.Println(slice)
	})
}

// Get hashset data from the cache by key
func HDEL(c *storage.Cache, hN string, ks []string) {
	if len(ks) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(hN)
		if !exists {
			logger.Error("can`t find %v in memory", hN)
			return
		}

		// CHECK
		cachedData.Requests++

		m, ok := cachedData.Value.(map[string]string)
		if !ok {
			logger.Error("%s isn't a hash", hN)
			return
		}

		for _, k := range ks {
			if _, exists := m[k]; exists {
				delete(m, k)
			} else {
				logger.Error("key %s doesn`t exist in hash %s", k, hN)
				return
			}
		}

		logger.Success("OK")
	})
}

// Check if the keys exist in the hashset and return their quantity
func HCONTAINS(c *storage.Cache, hN string, ks []string) {
	var res int

	cachedData, exists := c.GetSafe(hN)
	if !exists {
		logger.Error("can`t find %v in memory", hN)
		return
	}

	// CHECK
	cachedData.Requests++

	m, ok := cachedData.Value.(map[string]string)
	if !ok {
		logger.Error("%s isn't a hash", hN)
		fmt.Println(0)
	}

	for _, k := range ks {
		if _, exists := m[k]; exists {
			res++
		}
	}

	fmt.Println(res)
}

// Check if the keys exist in the hashset and return array
func LHCONTAINS(c *storage.Cache, hN string, ks []string) {
	res := make([]int, len(ks))

	cachedData, exists := c.GetSafe(hN)
	if !exists {
		logger.Error("can`t find %v in memory", hN)
		return
	}

	// CHECK
	cachedData.Requests++

	m, ok := cachedData.Value.(map[string]string)
	if !ok {
		logger.Error("%s isn't a hash", hN)
		fmt.Println(0)
	}

	for i, k := range ks {
		if _, exists := m[k]; exists {
			res[i] = 1
		} else {
			res[i] = 0
		}
	}

	fmt.Println(res)
}

// Get hashset length
func HLEN(c *storage.Cache, hN string) {
	c.WithRWLock(func() {
		cachedData, exists := c.GetUnsafe(hN)
		if !exists {
			logger.Error("can`t find %v in memory", hN)
			return
		}

		// CHECK
		cachedData.Requests++

		m, ok := cachedData.Value.(map[string]string)
		if !ok {
			logger.Error("%s isn't a hash", hN)
			return
		}

		fmt.Println(len(m))
	})
}
