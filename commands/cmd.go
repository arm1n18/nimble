package commands

import (
	"cache/logger"
	"cache/storage"
	"cache/utils"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

func getKeysByPattern(m map[string]*storage.CacheData, pattern string, lim int) []string {
	var kvs []string

	symbol, ok := utils.GetPatternSymbol(pattern)
	if !ok {
		return kvs
	}

	switch symbol {
	case "*":
		{
			for k := range m {
				match, err := path.Match(pattern, k)
				if err != nil {
					continue
				}

				if match {
					kvs = append(kvs, k)
					if lim > 0 && len(kvs) >= lim {
						break
					}
				}
			}
		}
	case "?":
		{
			for k := range m {
				match, err := path.Match(pattern, k)
				if err != nil {
					continue
				}

				if match && len(pattern) == len(k) {
					kvs = append(kvs, k)
					if lim > 0 && len(kvs) >= lim {
						break
					}
				}
			}
		}
	}

	return kvs
}

// Store data of string type in the cache
func SET(c *storage.Cache, k, v string) {
	c.WithLock(func() {
		if k == "" {
			logger.Error("Key cannot be empty")
			return
		}

		if k == "*" {
			logger.Error("Can`t use * as key")
			return
		}

		c.SetUnsafe(k, &storage.CacheData{
			Value:     v,
			Type:      storage.String,
			Requests:  1,
			CreatedAt: time.Now(),
		})
	})

	logger.Success("OK")
}

// Store data of string type in the cache
func MSET(c *storage.Cache, kvs ...string) {
	if len(kvs)%2 == 0 && len(kvs) != 0 {
		// if ok := removeQuotes(&s, 1, 1); !ok {
		// 	return
		// }

		c.WithLock(func() {
			for i := 0; i < len(kvs); i += 2 {
				k, v := kvs[i], kvs[i+1]

				if k == "" {
					logger.Error("Key cannot be empty")
					return
				}

				if k == "*" {
					logger.Error("Can`t use * as key")
					return
				}

				c.SetUnsafe(k, &storage.CacheData{
					Value:     v,
					Type:      storage.String,
					Requests:  1,
					CreatedAt: time.Now(),
				})
			}
		})
	} else {
		logger.Error("Not enough values")
		return
	}

	logger.Success("OK")
}

// Get string type of data from the cache
func GET(c *storage.Cache, k string) {
	c.WithLock(func() {
		if cd, exists := c.GetUnsafe(k); exists {
			cd.Requests++

			var cV interface{}

			// m, ok := cd.Value.(map[string]struct{})
			// if ok {
			// 	tS := make([]string, 0, len(m))

			// 	for k := range m {
			// 		tS = append(tS, k)
			// 	}

			// 	cV = tS
			// } else {
			// 	cV = cd.Value
			// }

			cV = cd.Value

			fmt.Println(cV)
			return
		} else {
			fmt.Println(nil)
			return
		}
	})
}

// Get string type of data from the cache
func MGET(c *storage.Cache, kvs ...string) {
	var res []interface{}

	if len(kvs) != 1 {
		res = make([]interface{}, 0, len(kvs))
	}

	c.WithLock(func() {
		for _, k := range kvs {
			if cd, exists := c.GetUnsafe(k); exists {
				cd.Requests++

				var cV interface{}

				// m, ok := cd.Value.(map[string]string{})
				// if ok {
				// 	tS := make([]string, 0, len(m))

				// 	for k := range m {
				// 		tS = append(tS, k)
				// 	}

				// 	cV = tS
				// } else {
				// 	cV = cd.Value
				// }

				cV = cd.Value

				if len(kvs) == 1 {
					fmt.Println(cV)
					return
				} else {
					res = append(res, cV)
				}
			} else {
				if len(k) == 1 {
					fmt.Println(nil)
					return
				} else {
					res = append(res, nil)
				}
			}
		}
		fmt.Println(res)
	})
}

// Remove any type of data from the cache
func DEL(c *storage.Cache, k string) {
	if k == "*" {
		c.ResetCache()
		logger.Success("OK")
		return
	}

	c.WithLock(func() {
		if _, exists := c.GetUnsafe(k); exists {
			delete(c.GetData(), k)
		} else {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		logger.Success("OK")
	})
}

// Copy data from one structure to another
func COPY(c *storage.Cache, k1, k2 string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k1)
		if !exists {
			logger.Error("Can`t find %v in memory", k1)
			return
		}

		c.SetPartialUnsafe(k2, storage.CacheDataUpdate{Value: &cd.Value, Type: &cd.Type})

		logger.Success("OK")
	})
}

// Show all the keys
func LIST(c *storage.Cache) {
	i := 1

	c.WithRWLock(func() {
		cd := c.GetData()

		for v, k := range cd {
			fmt.Printf("%v) [ %s ] = %v\n", i, v, k.Value)
			i++
		}
	})
}

// Show the number of keys
func LISTLEN(c *storage.Cache) {
	c.WithRWLock(func() {
		fmt.Println(len(c.GetData()))
	})
}

/*
Set key expiration (Time to Kill)

Description:
Sets a lifespan for the specified key. The key will be automatically deleted from the cache
after the given number of seconds.

Example:

  - Pattern: TTL session:123 360

  - Explanation: Sets the key "session:123" to expire in 360 seconds (6 minutes)
*/
func TTK(c *storage.Cache, k, v string) {
	t, err := strconv.Atoi(v)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	if t < -1 {
		logger.Error("TTL must be >= -1")
		return
	}

	var expiresAt *time.Time
	if t != -1 {
		et := time.Now().Add(time.Duration(t) * time.Second)
		expiresAt = &et
	}

	c.WithLock(func() {

		if k == "*" {
			for key := range c.GetData() {
				c.SetPartialUnsafe(key, storage.CacheDataUpdate{
					ExpiresAt: expiresAt,
				})
			}
			logger.Success("OK")
			return
		}

		_, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		c.SetPartialUnsafe(k, storage.CacheDataUpdate{
			ExpiresAt: expiresAt,
		})

		logger.Success("OK")
	})
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
*/
func TTL(c *storage.Cache, k string) {
	c.WithRWLock(func() {

		if dataCache, exists := c.GetSafe(k); exists {
			if dataCache.ExpiresAt == nil {
				fmt.Println(-1) // -1 means that data has no TTK
				return
			}
			// fmt.Println(int(time.Since(dataCache.TTL).Seconds()) * -1)
			fmt.Println(int(time.Until(*dataCache.ExpiresAt).Seconds()))
		} else {
			logger.Error("Can`t find %v in memory", k)
			return
		}
	})
}

/*
Count how many keys exist in the cache

Description:
Checkvs one or more keys in the cache and returns the total number of kesy that exist.

Example:

  - Pattern: EXISTS KEY_1 KEY_2 KEY_0

  - Result: 2

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func EXISTS(c *storage.Cache, kvs ...string) {
	var res int

	c.WithRWLock(func() {
		cd := c.GetData()

		for _, k := range kvs {
			if _, exists := cd[k]; exists {
				res++
			}
		}

		fmt.Println(res)
	})
}

/*
Check if the keys exist and return array.

Description:
Checkvs one or more keys in the cache and returns an array of integers.

Behavior:
  - For each key provided:
    1 if the key exists or
    0 if the key does not exist

Example:

  - Pattern: LEXISTS KEY_1 KEY_2 KEY_0

  - Result: [ 1, 1, 0 ]

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func LEXISTS(c *storage.Cache, kvs ...string) {
	res := make([]int, len(kvs))

	c.WithRWLock(func() {
		cd := c.GetData()

		for i, k := range kvs {
			if _, exists := cd[k]; exists {
				res[i] = 1
			} else {
				res[i] = 0
			}
		}

		fmt.Println(res)
	})
}

/*
Get keys from cache by pattern

1. Pattern '*'

  - Return all keys that start with prefix before '*'

  - Pattern: KEYS user:*

  - Result: [ user:1, user:123, user:ABC ]

2. Pattern '?'

  - Return all keys that start with prefix before '?' and have a length equal to the number of '?' after prefix

  - Pattern: KEYS user:???

  - Result: [ user:123, user:256, user:ABC ]
*/
func KEYS(c *storage.Cache, args ...string) {
	if len(args) == 0 || len(args) == 2 || len(args) > 3 {
		logger.Error("Invalid syntax")
		return
	}

	d := -1

	if len(args) > 2 {
		if strings.ToLower(args[1]) == "count" {
			var err error
			d, err = strconv.Atoi(args[2])
			if err != nil {
				logger.Error("Can`t parse number")
				return
			}
		} else {
			logger.Error("Invalid syntax")
			return
		}
	}

	if utils.IsPatternCmd(args[0]) {
		c.WithLock(func() {
			fmt.Println(getKeysByPattern(c.GetData(), args[0], d))
		})
	}
}

func RENAME(c *storage.Cache, sK, tK string) {
	c.WithLock(func() {
		sCd, exists := c.GetUnsafe(sK)
		if !exists {
			logger.Error("can`t find %v in memory", sK)
			return
		}

		_, exists = c.GetUnsafe(tK)
		if exists {
			logger.Error("%v already exists", tK)
			return
		}

		d := c.GetData()

		delete(d, sK)
		c.SetUnsafe(tK, sCd)

		logger.Success("OK")
	})
}

func INFO(c *storage.Cache, k string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		fmt.Println("[INFO]:")
		fmt.Println("	{")
		fmt.Println("		key:", k)
		fmt.Println("		value:", cd.Value)
		fmt.Println("		requests:", cd.Requests)
		fmt.Println("		type:", cd.Type)
		fmt.Println("		created_at:", cd.CreatedAt)

		if cd.ExpiresAt != nil {
			ttl := time.Until(*cd.ExpiresAt)
			if ttl < 0 {
				ttl = 0
			}

			fmt.Println("		ttl:", int(ttl.Seconds()))
			fmt.Println("		expires_at:", cd.ExpiresAt.Format(time.RFC3339))
		} else {
			fmt.Println("		ttl: -1")
		}

		fmt.Println("	}")
	})
}
