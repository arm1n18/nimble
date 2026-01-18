package commands

import (
	"fmt"
	"nimble/formatter"
	"nimble/storage"
	"nimble/utils"
	"path"
	"strconv"
	"strings"
	"time"
)

func geteysByPattern(m map[string]*storage.CacheData, pattern string, lim int) []string {
	var args []string

	symbol, ok := utils.GetPatternSymbol(pattern)
	if !ok {
		return args
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
					args = append(args, k)
					if lim > 0 && len(args) >= lim {
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
					args = append(args, k)
					if lim > 0 && len(args) >= lim {
						break
					}
				}
			}
		}
	}

	return args
}

// Store data of string type in the cache
func SET(c *storage.Cache, k, v string) string {
	var result string

	c.WithLock(func() {
		if k == "" || k == "*" {
			result = formatter.ErrWrongKey.Error()
			return
		}

		c.SetUnsafe(k, &storage.CacheData{
			Value:     v,
			Type:      storage.String,
			Requests:  1,
			CreatedAt: time.Now(),
		})

		result = formatter.Ok()
	})

	return result
}

// Store data of string type in the cache
func MSET(c *storage.Cache, args ...string) string {
	var result string

	if len(args)%2 == 0 && len(args) != 0 {
		// if ok := removeQuotes(&s, 1, 1); !ok {
		// 	return
		// }

		for i := 0; i < len(args); i += 2 {
			k, _ := args[i], args[i+1]

			if k == "" || k == "*" {
				return formatter.ErrWrongKey.Error()
			}
		}

		c.WithLock(func() {

			for i := 0; i < len(args); i += 2 {
				k, v := args[i], args[i+1]

				c.SetUnsafe(k, &storage.CacheData{
					Value:     v,
					Type:      storage.String,
					Requests:  1,
					CreatedAt: time.Now(),
				})
			}

			result = formatter.Ok()
		})
	} else {
		result = formatter.ErrNotEnoughValues.Error()
		return result
	}

	return result
}

// Get string type of data from the cache
func GET(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {
		if cd, exists := c.GetUnsafe(k); exists {
			cd.Requests++

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

			result = formatter.String(cd.Value)
			return
		} else {
			result = formatter.Nil()
			return
		}
	})

	return result
}

// Get string type of data from the cache
func MGET(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var arr []string

		for _, k := range args {
			if cd, exists := c.GetUnsafe(k); exists {
				cd.Requests++

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

				arr = append(arr, cd.Value)
			} else {
				arr = append(arr, formatter.Nil())
			}
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
}

// Remove any type of data from the cache
func DEL(c *storage.Cache, args ...string) string {
	var result string
	var q int

	if len(args) == 0 {
		return formatter.ErrNotEnoughValues.Error()
	}

	if len(args) == 1 && args[0] == "*" {
		c.WithRWLock(func() {
			cd := c.GetData()
			q = len(cd)
		})

		c.ResetCache()
		return formatter.Number(q)
	}

	c.WithLock(func() {
		for _, k := range args {
			if _, exists := c.GetUnsafe(k); exists {
				delete(c.GetData(), k)
				q++
			}
		}

		result = formatter.Number(q)
	})

	return result
}

// Copy data from one structure to another
func COPY(c *storage.Cache, f, t string) string {
	var result string

	if len(f) == 0 || len(t) == 0 || f == "*" || t == "*" {
		return formatter.ErrWrongKey.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(f)
		if !exists {
			// result = formatter.ErrorMessage("Can`t find %v in memory", f)
			result = formatter.Failure()
			return
		}

		c.SetPartialUnsafe(t, storage.CacheDataUpdate{Value: &cd.Value, Type: &cd.Type})

		result = formatter.Success()
	})

	return result
}

// Show all the keys
func LIST(c *storage.Cache) string {
	var result string

	c.WithRWLock(func() {
		cd := c.GetData()
		arr := make([]string, 0, len(cd))

		for k := range cd {
			arr = append(arr, k)
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
}

// Show the number of keys
func LISTLEN(c *storage.Cache) string {
	var result string

	c.WithRWLock(func() {
		result = formatter.Number(len(c.GetData()))
	})

	return result
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
func TTK(c *storage.Cache, k, v string) string {
	var result string

	t, err := strconv.Atoi(v)
	if err != nil {
		return formatter.ErrInvalidTTL.Error()
	}

	if t < -1 {
		return formatter.ErrInvalidTTL.Error()
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
			result = formatter.Success()
			return
		}

		_, exists := c.GetUnsafe(k)
		if !exists {
			result = formatter.Failure()
			return
		}

		c.SetPartialUnsafe(k, storage.CacheDataUpdate{
			ExpiresAt: expiresAt,
		})

		result = formatter.Success()
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
func TTL(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {

		if dataCache, exists := c.GetSafe(k); exists {
			if dataCache.ExpiresAt == nil {
				result = "-1"
				return
			}
			// fmt.Println(int(time.Since(dataCache.TTL).Seconds()) * -1)
			result = formatter.Number(int(time.Until(*dataCache.ExpiresAt).Seconds()))
		} else {
			result = "-2"
			return
		}
	})

	return result
}

/*
Count how many keys exist in the cache

Description:
Checargs one or more keys in the cache and returns the total number of kesy that exist.

Example:

  - Pattern: EXISTS KEY_1 KEY_2 KEY_0

  - Result: 2

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func EXISTS(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int
		cd := c.GetData()

		for _, k := range args {
			if _, exists := cd[k]; exists {
				q++
			}
		}

		result = formatter.Number(q)
	})

	return result
}

/*
Check if the keys exist and return array.

Description:
Checargs one or more keys in the cache and returns an array of integers.

Behavior:
  - For each key provided:
    1 if the key exists or
    0 if the key does not exist

Example:

  - Pattern: LEXISTS KEY_1 KEY_2 KEY_0

  - Result: [ 1, 1, 0 ]

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func LEXISTS(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(args))
		cd := c.GetData()

		for i, k := range args {
			if _, exists := cd[k]; exists {
				arr[i] = "1"
			} else {
				arr[i] = "0"
			}
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
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
func KEYS(c *storage.Cache, args ...string) string {
	var result string

	if len(args) == 0 || len(args) == 2 || len(args) > 3 {
		return formatter.ErrInvalidSyntax.Error()
	}

	d := -1

	if len(args) > 2 {
		if strings.ToLower(args[1]) == "count" {
			var err error
			d, err = strconv.Atoi(args[2])
			if err != nil {
				return formatter.ErrNotANumber.Error()
			}
		} else {
			return formatter.ErrInvalidSyntax.Error()
		}
	}

	if utils.IsPatternCmd(args[0]) {
		c.WithRWLock(func() {
			result = formatter.Array(serializeList(geteysByPattern(c.GetData(), args[0], d)))
		})
	}

	return result
}

func RENAME(c *storage.Cache, f, t string) string {
	var result string

	c.WithLock(func() {
		sCd, exists := c.GetUnsafe(f)
		if !exists {
			// result = formatter.ErrorMessage("can`t find %v in memory", f)
			result = formatter.Failure()
			return
		}

		_, exists = c.GetUnsafe(t)
		if exists {
			// result = formatter.ErrorMessage("%v already exists", t)
			result = formatter.Failure()
			return
		}

		d := c.GetData()

		delete(d, f)
		c.SetUnsafe(t, sCd)

		result = formatter.Success()
	})

	return result
}

func INFO(c *storage.Cache, k string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			// formatter.ErrorMessage("Can`t find %v in memory", k)
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
