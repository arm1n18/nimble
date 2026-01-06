package commands

import (
	"cache/logger"
	"cache/storage"
	"fmt"
	"time"
)

func SADD(c *storage.Cache, z string, vs []string) {
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
				Value:     m,
				Requests:  1,
				TimeStamp: time.Now(),
			})
		} else {
			cd.Requests++
			if m, ok := cd.Value.(map[string]struct{}); ok {
				for _, v := range vs {
					if _, ok := m[v]; !ok {
						m[v] = struct{}{}
					}
				}
			} else {
				logger.Error("%s isn't a set", z)
				return
			}
		}
	})

	logger.Success("OK")
}

func SREM(c *storage.Cache, z string, vs []string) {
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

			if m, ok := cd.Value.(map[string]struct{}); ok {
				for _, v := range vs {
					delete(m, v)
				}
			} else {
				logger.Error("%s isn't a set", z)
				return
			}
		}
	})

	logger.Success("OK")
}

func SCONTAINS(c *storage.Cache, z string, vs []string) {
	var res int

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		// CHECK
		cd.Requests++

		// m, ok := cd.Value.(map[string]struct{})
		// if !ok {
		// 	logger.Error("%s isn`t set", m)
		// 	return
		// }

		switch m := cd.Value.(type) {
		case map[string]struct{}:
			for _, v := range vs {
				if _, ok := m[v]; ok {
					res++
				}
			}
		case ZSet:
			for _, v := range vs {
				if _, ok := m.Items[v]; ok {
					res++
				}
			}
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})

	fmt.Println(res)
}

func LSCONTAINS(c *storage.Cache, z string, vs []string) {
	res := make([]int, len(vs))

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		// CHECK
		cd.Requests++

		switch m := cd.Value.(type) {
		case map[string]struct{}:
			for i, v := range vs {
				if _, ok := m[v]; ok {
					res[i] = 1
				} else {
					res[i] = 0
				}
			}
		case ZSet:
			for i, v := range vs {
				if _, ok := m.Items[v]; ok {
					res[i] = 1
				} else {
					res[i] = 0
				}
			}
		default:
			logger.Error("%s isn't a set", z)
			return
		}
	})

	fmt.Println(res)
}

func SLEN(c *storage.Cache, z string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("can`t find %v in memory", z)
			return
		}

		// CHECK
		cd.Requests++

		// m, ok := cachedData.Value.(map[string]struct{})
		// if !ok {
		// 	logger.Error("%s isn't a set", z)
		// 	return
		// }

		switch m := cd.Value.(type) {
		case map[string]struct{}:
			fmt.Println(len(m))
		case ZSet:
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

		// CHECK
		cd.Requests++

		switch m := cd.Value.(type) {
		case map[string]struct{}:
			for k := range m {
				slice = append(slice, k)
			}
		case ZSet:
			for _, k := range m.Order {
				slice = append(slice, k.Member)
			}
		default:
			logger.Error("%s isn't a set", z)
			return
		}

		fmt.Println(slice)
	})
}
