package commands

import (
	"encoding/json"
	"nimble/protocol"
	"nimble/storage"
	"time"
)

func parseSet(s string) (map[string]struct{}, bool) {
	var m map[string]struct{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, false
	}
	if m == nil {
		m = make(map[string]struct{})
	}

	return m, true
}

func serializeSet(s map[string]struct{}) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func SADD(c *storage.Cache, z string, args ...string) string {
	var result string

	if len(args) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(z)
		if !exists {
			m := make(map[string]struct{})
			for _, v := range args {
				if _, ok := m[v]; !ok {
					m[v] = struct{}{}
					q++
				}
			}

			c.SetUnsafe(z, &storage.CacheData{
				Value:     serializeSet(m),
				Type:      storage.Set,
				Requests:  1,
				CreatedAt: time.Now(),
			})
			result = protocol.Number(q)
		} else {
			cd.Requests++
			if m, ok := parseSet(cd.Value); ok {
				for _, v := range args {
					if _, ok := m[v]; !ok {
						m[v] = struct{}{}
						q++
					}
				}

				cd.Value = serializeSet(m)
				result = protocol.Number(q)
			} else {
				// result = protocol.ErrorMessage("%s isn't a set", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

func SREM(c *storage.Cache, z string, args ...string) string {
	var result string

	if len(args) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", z)
			result = protocol.Failure()
			return
		} else {
			cd.Requests++

			if m, ok := parseSet(cd.Value); ok {
				for _, v := range args {
					if _, exists := m[v]; exists {
						delete(m, v)
						q++
					}
				}
				cd.Value = serializeSet(m)
				result = protocol.Number(q)
			} else {
				// result = protocol.ErrorMessage("%s isn't a set", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

func SCONTAINS(c *storage.Cache, z string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", z)
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse set: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			for _, v := range args {
				if _, ok := m[v]; ok {
					q++
				}
			}

			result = protocol.Number(q)
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse zset: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			for _, v := range args {
				if _, ok := m.Items[v]; ok {
					q++
				}
			}

			result = protocol.Number(q)
		default:
			// result = protocol.ErrorMessage("%s isn't a set", z)
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

func LSCONTAINS(c *storage.Cache, z string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(args))

		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", z)
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse set: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			for i, v := range args {
				if _, ok := m[v]; ok {
					arr[i] = "1"
				} else {
					arr[i] = "0"
				}
			}

			result = protocol.Array(serializeList(arr))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse zset: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			for i, v := range args {
				if _, ok := m.Items[v]; ok {
					arr[i] = "1"
				} else {
					arr[i] = "0"
				}
			}

			result = protocol.Array(serializeList(arr))
		default:
			// result = protocol.ErrorMessage("%s isn't a set", z)
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

func SLEN(c *storage.Cache, z string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", z)
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse set: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			result = protocol.Number(len(m))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse zset: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			result = protocol.Number(len(m.Items))
		default:
			// result = protocol.ErrorMessage("%s isn't a set", z)
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

func SMEMBERS(c *storage.Cache, z string) string {
	var result string

	c.WithRWLock(func() {
		var arr []string

		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", z)
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse set: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}
			for k := range m {
				arr = append(arr, k)
			}

			result = protocol.Array(serializeList(arr))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("can't parse zset: %s", z)
				result = protocol.ErrMismatchType.Error()
				return
			}

			for _, k := range m.Order {
				arr = append(arr, k.Member)
			}

			result = protocol.Array(serializeList(arr))
		default:
			// result = protocol.ErrorMessage("%s isn't a set", z)
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}
