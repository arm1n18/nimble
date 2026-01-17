package commands

import (
	"cache/logger"
	"cache/storage"
	"encoding/json"
	"fmt"
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

func SADD(c *storage.Cache, z string, vs ...string) {
	if len(vs) == 0 {
		logger.Error("Not enough values")
		return
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
			logger.Success("OK")
		} else {
			cd.Requests++
			if m, ok := parseSet(cd.Value); ok {
				for _, v := range vs {
					if _, ok := m[v]; !ok {
						m[v] = struct{}{}
					}
				}

				cd.Value = serializeSet(m)
				logger.Success("OK")
			} else {
				logger.Error("%s isn't a set", z)
				return
			}
		}
	})
}

func SREM(c *storage.Cache, z string, vs ...string) {
	if len(vs) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("Can`t find %v in memory", z)
			return
		} else {
			cd.Requests++

			if m, ok := parseSet(cd.Value); ok {
				for _, v := range vs {
					delete(m, v)
				}
				cd.Value = serializeSet(m)
				logger.Success("OK")
			} else {
				logger.Error("%s isn't a set", z)
				return
			}
		}
	})
}

func SCONTAINS(c *storage.Cache, z string, vs ...string) {
	var res int

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				logger.Error("can't parse set: %s", z)
				return
			}
			for _, v := range vs {
				if _, ok := m[v]; ok {
					res++
				}
			}
			fmt.Println(res)
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				logger.Error("can't parse zset: %s", z)
				return
			}
			for _, v := range vs {
				if _, ok := m.Items[v]; ok {
					res++
				}
			}
			fmt.Println(res)
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})

}

func LSCONTAINS(c *storage.Cache, z string, vs ...string) {
	res := make([]int, len(vs))

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				logger.Error("can't parse set: %s", z)
				return
			}
			for i, v := range vs {
				if _, ok := m[v]; ok {
					res[i] = 1
				} else {
					res[i] = 0
				}
			}

			fmt.Println(res)
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				logger.Error("can't parse zset: %s", z)
				return
			}
			for i, v := range vs {
				if _, ok := m.Items[v]; ok {
					res[i] = 1
				} else {
					res[i] = 0
				}
			}

			fmt.Println(res)
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})
}

func SLEN(c *storage.Cache, z string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				logger.Error("can't parse set: %s", z)
				return
			}
			fmt.Println(len(m))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				logger.Error("can't parse zset: %s", z)
				return
			}
			fmt.Println(len(m.Items))
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})
}

func SMEMBERS(c *storage.Cache, z string) {
	var slice []string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				logger.Error("can't parse set: %s", z)
				return
			}
			for k := range m {
				slice = append(slice, k)
			}

			fmt.Println(slice)
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				logger.Error("can't parse zset: %s", z)
				return
			}
			for _, k := range m.Order {
				slice = append(slice, k.Member)
			}

			fmt.Println(slice)
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})
}
