package commands

import (
	"encoding/json"
	"nimble/formatter"
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

func SADD(c *storage.Cache, z string, vs ...string) string {
	var result string

	if len(vs) == 0 {
		return formatter.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			m := make(map[string]struct{})
			for _, v := range vs {
				if _, ok := m[v]; !ok {
					m[v] = struct{}{}
				}
			}

			c.SetUnsafe(z, &storage.CacheData{
				Value:     serializeSet(m),
				Type:      storage.Set,
				Requests:  1,
				CreatedAt: time.Now(),
			})
			result = formatter.Success()
		} else {
			cd.Requests++
			if m, ok := parseSet(cd.Value); ok {
				for _, v := range vs {
					if _, ok := m[v]; !ok {
						m[v] = struct{}{}
					}
				}

				cd.Value = serializeSet(m)
				result = formatter.Success()
			} else {
				// result = formatter.ErrorMessage("%s isn't a set", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

func SREM(c *storage.Cache, z string, vs ...string) string {
	var result string

	if len(vs) == 0 {
		return formatter.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = formatter.ErrorMessage("Can`t find %v in memory", z)
			result = formatter.Failure()
			return
		} else {
			cd.Requests++

			if m, ok := parseSet(cd.Value); ok {
				for _, v := range vs {
					delete(m, v)
				}
				cd.Value = serializeSet(m)
				result = formatter.Success()
			} else {
				// result = formatter.ErrorMessage("%s isn't a set", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

func SCONTAINS(c *storage.Cache, z string, vs ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = formatter.ErrorMessage("can`t find %v in memory", z)
			result = formatter.Nil()
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse set: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			for _, v := range vs {
				if _, ok := m[v]; ok {
					q++
				}
			}

			result = formatter.Number(q)
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse zset: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			for _, v := range vs {
				if _, ok := m.Items[v]; ok {
					q++
				}
			}

			result = formatter.Number(q)
		default:
			// result = formatter.ErrorMessage("%s isn't a set", z)
			result = formatter.ErrMismatchType.Error()
			return
		}
	})

	return result
}

func LSCONTAINS(c *storage.Cache, z string, vs ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(vs))

		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = formatter.ErrorMessage("can`t find %v in memory", z)
			result = formatter.Nil()
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse set: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			for i, v := range vs {
				if _, ok := m[v]; ok {
					arr[i] = "1"
				} else {
					arr[i] = "0"
				}
			}

			result = formatter.Array(serializeList(arr))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse zset: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			for i, v := range vs {
				if _, ok := m.Items[v]; ok {
					arr[i] = "1"
				} else {
					arr[i] = "0"
				}
			}

			result = formatter.Array(serializeList(arr))
		default:
			// result = formatter.ErrorMessage("%s isn't a set", z)
			result = formatter.ErrMismatchType.Error()
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
			// result = formatter.ErrorMessage("can`t find %v in memory", z)
			result = formatter.Number(-1)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse set: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			result = formatter.Number(len(m))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse zset: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			result = formatter.Number(len(m.Items))
		default:
			// result = formatter.ErrorMessage("%s isn't a set", z)
			result = formatter.ErrMismatchType.Error()
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
			// result = formatter.ErrorMessage("can`t find %v in memory", z)
			// result = formatter.Nil()
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse set: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
			for k := range m {
				arr = append(arr, k)
			}

			result = formatter.Array(serializeList(arr))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("can't parse zset: %s", z)
				result = formatter.ErrMismatchType.Error()
				return
			}

			for _, k := range m.Order {
				arr = append(arr, k.Member)
			}

			result = formatter.Array(serializeList(arr))
		default:
			// result = formatter.ErrorMessage("%s isn't a set", z)
			result = formatter.ErrMismatchType.Error()
			return
		}
	})

	return result
}
