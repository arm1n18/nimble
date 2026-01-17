package commands

import (
	"cache/logger"
	"cache/storage"
	"encoding/json"
	"fmt"
	"time"
)

func parseHash(s string) (map[string]string, bool) {
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, false
	}
	if m == nil {
		m = make(map[string]string)
	}

	return m, true
}

func serializeHash(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// Store hashset in the cache
func HSET(c *storage.Cache, h string, kv ...string) {
	hash := make(map[string]string)

	if len(kv)%2 == 0 && len(kv) != 0 {

		// if ok := removeQuotes(&s, 0, 1); !ok {
		// 	return
		// }

		c.WithLock(func() {
			if cd, exists := c.GetUnsafe(h); exists {
				cd.Requests++

				if m, ok := parseHash(cd.Value); ok {
					for i := 0; i < len(kv); i += 2 {
						m[kv[i]] = kv[i+1]
					}

					c.SetUnsafe(h, &storage.CacheData{
						Value:    serializeHash(m),
						Type:     storage.Hash,
						Requests: cd.Requests + 1,
					})

					logger.Success("OK")
					return
				}
			}

			for i := 0; i < len(kv); i += 2 {
				hash[kv[i]] = kv[i+1]
			}

			c.SetUnsafe(h, &storage.CacheData{
				Value:     serializeHash(hash),
				Type:      storage.Hash,
				Requests:  1,
				CreatedAt: time.Now(),
			})

			logger.Success("OK")
		})
	} else {
		logger.Error("Not enough values")
		return
	}
}

// Get hashset data from the cache by key
func HGET(c *storage.Cache, h string, ks ...string) {
	res := make([]string, len(ks))

	cd, exists := c.GetSafe(h)
	if !exists {
		if len(ks) == 1 {
			fmt.Println(nil)
		} else {
			res = append(res, "nil")
		}
		return
	}

	cd.Requests++

	m, ok := parseHash(cd.Value)
	if !ok {
		logger.Error("%s isn't a hash", h)
		for i := range res {
			res[i] = "nil"
		}

		fmt.Println(res)
	}

	for i, k := range ks {
		if v, exists := m[k]; exists {
			res[i] = v
		} else {
			res[i] = "nil"
		}
	}

	fmt.Println(res)
}

// Get hashset data keys
func HKEYS(c *storage.Cache, h string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			logger.Error("can`t find %v in memory", h)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			logger.Error("%s isn't a hash", h)
			return
		}

		s := make([]string, 0, len(m))
		for k := range m {
			s = append(s, k)
		}

		fmt.Println(s)
	})
}

// Get hashset data values
func HVALUES(c *storage.Cache, h string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			logger.Error("can`t find %v in memory", h)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			logger.Error("%s isn't a hash", h)
			return
		}

		s := make([]string, 0, len(m))
		for _, v := range m {
			s = append(s, v)
		}

		fmt.Println(s)
	})
}

// Get hashset data from the cache by key
func HDEL(c *storage.Cache, h string, kv ...string) {
	if len(kv) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			logger.Error("can`t find %v in memory", h)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			logger.Error("%s isn't a hash", h)
			return
		}

		for _, k := range kv {
			if _, exists := m[k]; exists {
				delete(m, k)
			} else {
				logger.Error("key %s doesn`t exist in hash %s", k, h)
				return
			}
		}

		c.SetUnsafe(h, &storage.CacheData{
			Value:    serializeHash(m),
			Requests: cd.Requests + 1,
		})

		logger.Success("OK")
	})
}

// Check if the keys exist in the hashset and return their quantity
func HCONTAINS(c *storage.Cache, h string, kv ...string) {
	var res int

	cd, exists := c.GetSafe(h)
	if !exists {
		logger.Error("can`t find %v in memory", h)
		return
	}

	cd.Requests++

	m, ok := parseHash(cd.Value)
	if !ok {
		logger.Error("%s isn't a hash", h)
		fmt.Println(0)
	}

	for _, k := range kv {
		if _, exists := m[k]; exists {
			res++
		}
	}

	fmt.Println(res)
}

// Check if the keys exist in the hashset and return array
func LHCONTAINS(c *storage.Cache, h string, kv ...string) {
	res := make([]int, len(kv))

	cd, exists := c.GetSafe(h)
	if !exists {
		logger.Error("can`t find %v in memory", h)
		return
	}

	cd.Requests++

	m, ok := parseHash(cd.Value)
	if !ok {
		logger.Error("%s isn't a hash", h)
		fmt.Println(0)
	}

	for i, k := range kv {
		if _, exists := m[k]; exists {
			res[i] = 1
		} else {
			res[i] = 0
		}
	}

	fmt.Println(res)
}

// Get hashset length
func HLEN(c *storage.Cache, h string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			logger.Error("can`t find %v in memory", h)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			logger.Error("%s isn't a hash", h)
			return
		}

		fmt.Println(len(m))
	})
}
