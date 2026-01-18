package commands

import (
	"encoding/json"
	"nimble/formatter"
	"nimble/storage"
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
func HSET(c *storage.Cache, h string, kv ...string) string {
	var result string

	if len(kv)%2 == 0 && len(kv) != 0 {

		// if ok := removeQuotes(&s, 0, 1); !ok {
		// 	return
		// }

		c.WithLock(func() {
			hash := make(map[string]string)

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

					result = formatter.Ok()
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

			result = formatter.Ok()
		})
	} else {
		return formatter.ErrNotEnoughValues.Error()
	}

	return result
}

// Get hashset data from the cache by key
func HGET(c *storage.Cache, h string, ks ...string) string {
	var result string

	c.WithLock(func() {
		arr := make([]string, len(ks))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			// todocmd
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		for i, k := range ks {
			if v, exists := m[k]; exists {
				arr[i] = v
			} else {
				arr[i] = formatter.Nil()
			}
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
}

// Get hashset data keys
func HKEYS(c *storage.Cache, h string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		s := make([]string, 0, len(m))
		for k := range m {
			s = append(s, k)
		}

		result = formatter.Array(serializeList(s))
	})

	return result
}

// Get hashset data values
func HVALUES(c *storage.Cache, h string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			// result = formatter.ErrorMessage("can`t find %v in memory", h)
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		s := make([]string, 0, len(m))
		for _, v := range m {
			s = append(s, v)
		}

		result = formatter.Array(serializeList(s))
	})

	return result
}

// Get hashset data from the cache by key
func HDEL(c *storage.Cache, h string, kv ...string) string {
	var result string

	if len(kv) == 0 {
		return formatter.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			// result = formatter.ErrorMessage("can`t find %v in memory", h)
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		for _, k := range kv {
			if _, exists := m[k]; exists {
				delete(m, k)
				q++
			}
		}

		c.SetUnsafe(h, &storage.CacheData{
			Value:    serializeHash(m),
			Requests: cd.Requests + 1,
		})

		result = formatter.Number(q)
	})

	return result
}

// Check if the keys exist in the hashset and return their quantity
func HCONTAINS(c *storage.Cache, h string, kv ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = formatter.ErrorMessage("can`t find %v in memory", h)
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		for _, k := range kv {
			if _, exists := m[k]; exists {
				q++
			}
		}

		result = formatter.Number(q)
	})

	return result
}

// Check if the keys exist in the hashset and return array
func LHCONTAINS(c *storage.Cache, h string, kv ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(kv))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		for i, k := range kv {
			if _, exists := m[k]; exists {
				arr[i] = "1"
			} else {
				arr[i] = "0"
			}
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
}

// Get hashset length
func HLEN(c *storage.Cache, h string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = formatter.Nil()
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a hash", h)
			result = formatter.ErrMismatchType.Error()
			return
		}

		result = formatter.Number(len(m))
	})

	return result
}
