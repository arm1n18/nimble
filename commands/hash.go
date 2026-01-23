package commands

import (
	"encoding/json"
	"nimble/protocol"
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
func HSET(c *storage.Cache, h string, args ...string) string {
	var result string

	if len(args)%2 == 0 && len(args) != 0 {

		// if ok := removeQuotes(&s, 0, 1); !ok {
		// 	return
		// }

		c.WithLock(func() {
			hash := make(map[string]string)

			if cd, exists := c.GetUnsafe(h); exists {
				cd.Requests++

				if m, ok := parseHash(cd.Value); ok {
					for i := 0; i < len(args); i += 2 {
						m[args[i]] = args[i+1]
					}

					c.SetUnsafe(h, &storage.CacheData{
						Value:    serializeHash(m),
						Type:     storage.Hash,
						Requests: cd.Requests + 1,
					})

					result = protocol.Ok()
					return
				}
			}

			for i := 0; i < len(args); i += 2 {
				hash[args[i]] = args[i+1]
			}

			c.SetUnsafe(h, &storage.CacheData{
				Value:     serializeHash(hash),
				Type:      storage.Hash,
				Requests:  1,
				CreatedAt: time.Now(),
			})

			result = protocol.Ok()
		})
	} else {
		return protocol.ErrNotEnoughValues.Error()
	}

	return result
}

// Get hashset data from the cache by key
func HGET(c *storage.Cache, h string, ks ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(ks))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			// todocmd
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		for i, k := range ks {
			if v, exists := m[k]; exists {
				arr[i] = v
			} else {
				arr[i] = protocol.Nil()
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

// Get hashset data keys
func HKEYS(c *storage.Cache, h string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		s := make([]string, 0, len(m))
		for k := range m {
			s = append(s, k)
		}

		result = protocol.Array(serializeList(s))
	})

	return result
}

// Get hashset data values
func HVALUES(c *storage.Cache, h string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", h)
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		s := make([]string, 0, len(m))
		for _, v := range m {
			s = append(s, v)
		}

		result = protocol.Array(serializeList(s))
	})

	return result
}

// Get hashset data from the cache by key
func HDEL(c *storage.Cache, h string, args ...string) string {
	var result string

	if len(args) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", h)
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		for _, k := range args {
			if _, exists := m[k]; exists {
				delete(m, k)
				q++
			}
		}

		c.SetUnsafe(h, &storage.CacheData{
			Value:    serializeHash(m),
			Requests: cd.Requests + 1,
		})

		result = protocol.Number(q)
	})

	return result
}

// Check if the keys exist in the hashset and return their quantity
func HCONTAINS(c *storage.Cache, h string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = protocol.ErrorMessage("can`t find %v in memory", h)
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		for _, k := range args {
			if _, exists := m[k]; exists {
				q++
			}
		}

		result = protocol.Number(q)
	})

	return result
}

// Check if the keys exist in the hashset and return array
func LHCONTAINS(c *storage.Cache, h string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(args))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		for i, k := range args {
			if _, exists := m[k]; exists {
				arr[i] = "1"
			} else {
				arr[i] = "0"
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

// Get hashset length
func HLEN(c *storage.Cache, h string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn't a hash", h)
			result = protocol.ErrMismatchType.Error()
			return
		}

		result = protocol.Number(len(m))
	})

	return result
}
