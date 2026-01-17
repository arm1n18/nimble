package commands

import (
	"cache/logger"
	"cache/storage"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type ZItem struct {
	Member string
	Score  float64
}

type ZSet struct {
	Items map[string]float64
	Order []ZItem
}

func parseZSet(s string) (*ZSet, bool) {
	var z ZSet
	if err := json.Unmarshal([]byte(s), &z); err != nil {
		return nil, false
	}
	if z.Items == nil {
		z.Items = make(map[string]float64)
	}

	return &z, true
}

func serializeZSet(z ZSet) string {
	b, _ := json.Marshal(z)
	return string(b)
}

func addZOrder(s []ZItem, i ZItem) []ZItem {
	for id, v := range s {
		if i.Score > v.Score {
			s = append(s[:id], append([]ZItem{i}, s[id:]...)...)
			return s
		}
	}

	return append(s, i)
}

func sortZOrder(m map[string]float64) []ZItem {
	s := make([]ZItem, 0, len(m))

	for k, v := range m {
		s = append(s, ZItem{
			Member: k,
			Score:  v,
		})
	}

	sort.Slice(s, func(a, b int) bool {
		return s[a].Score > s[b].Score
	})

	return s
}

func removeZOrder(o []ZItem, t string) []ZItem {
	for i, v := range o {
		if v.Member == t {
			return append(o[:i], o[i+1:]...)
		}
	}

	return o
}

func ZADD(c *storage.Cache, z, v, s string) {
	score, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Error("Index must be a number: %s", s)
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			m := map[string]float64{
				v: score,
			}

			c.SetUnsafe(z, &storage.CacheData{
				Value: serializeZSet(ZSet{
					Items: m,
					Order: []ZItem{{Member: v, Score: score}},
				}),
				Type:      storage.ZSet,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				zs.Items[v] = score

				zs.Order = addZOrder(zs.Order, ZItem{
					Member: v,
					Score:  score,
				})

				cd.Value = serializeZSet(*zs)
			} else {
				logger.Error("%s isn't a zset", z)
				return
			}
		}

		logger.Success("OK")
	})
}

func ZREM(c *storage.Cache, z, v string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("Can`t find %v in memory", z)
			return
		} else {
			cd.Requests++

			if m, ok := parseZSet(cd.Value); ok {
				delete(m.Items, v)

				m.Order = removeZOrder(m.Order, v)
				cd.Value = serializeZSet(*m)
			} else {
				logger.Error("%s isn't a set", z)
				return
			}

		}
	})

	logger.Success("OK")
}

func ZRANGEBYSCORE(c *storage.Cache, z, s, e string) {
	var slice []ZItem

	if (s == "max" && e == "min") || (e == "max" && s == "min") {
		c.WithLock(func() {
			cd, exists := c.GetUnsafe(z)
			if !exists {
				logger.Error("can`t find %v in memory", z)
				return
			}

			cd.Requests++

			m, ok := parseZSet(cd.Value)
			if !ok {
				logger.Error("%s isn't a zset", z)
				return
			}

			if s == "max" && e == "min" {
				for _, k := range m.Order {
					slice = append(slice, k)
				}
			} else if e == "max" && s == "min" {
				for i := len(m.Order) - 1; i >= 0; i-- {
					slice = append(slice, m.Order[i])
				}
			}

			fmt.Println(slice)
		})
	} else {
		logger.Error("Invalid syntax")
	}
}

func SCORE(c *storage.Cache, z, k string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("Can`t find %v in memory", z)
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				if s, ok := zs.Items[k]; ok {
					fmt.Println(s)
				} else {
					fmt.Println(-1)
				}
			} else {
				logger.Error("%s isn't a zset", z)
				return
			}
		}
	})
}

func LSCORE(c *storage.Cache, z string, ks []string) {
	res := make([]float64, 0, len(ks))

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			logger.Error("Can`t find %v in memory", z)
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				for _, k := range ks {
					if s, ok := zs.Items[k]; ok {
						res = append(res, s)
					} else {
						res = append(res, -1)
					}
				}
			} else {
				logger.Error("%s isn't a zset", z)
				return
			}
		}

		fmt.Println(res)
	})
}
