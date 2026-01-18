package commands

import (
	"encoding/json"
	"nimble/formatter"
	"nimble/storage"
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

func serializeZItems(z ...ZItem) string {
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

func ZADD(c *storage.Cache, z, v, s string) string {
	var result string

	score, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// return formatter.ErrorMessage("Index must be a number: %s", s)
		return formatter.ErrNotANumber.Error()
	}

	if score < 0 {
		return formatter.ErrInvalidScore.Error()
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
			result = formatter.Success()
			return
		}

		cd.Requests++
		zs, ok := parseZSet(cd.Value)
		if !ok {
			// result = formatter.ErrorMessage("%s isn't a zset", z)
			result = formatter.ErrMismatchType.Error()
			return
		}

		if _, existsInSet := zs.Items[v]; existsInSet {
			result = formatter.Failure()
		} else {
			zs.Items[v] = score
			zs.Order = addZOrder(zs.Order, ZItem{
				Member: v,
				Score:  score,
			})
			cd.Value = serializeZSet(*zs)

			result = formatter.Success()
		}
	})

	return result
}

func ZREM(c *storage.Cache, z, v string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = formatter.ErrorMessage("Can`t find %v in memory", z)
			result = formatter.Failure()
			return
		} else {
			cd.Requests++

			zs, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("%s isn't a set", z)
				result = formatter.ErrMismatchType.Error()
				return
			}

			if _, existsInSet := zs.Items[v]; existsInSet {
				delete(zs.Items, v)

				zs.Order = removeZOrder(zs.Order, v)
				cd.Value = serializeZSet(*zs)
				result = formatter.Success()
			} else {
				result = formatter.Failure()
			}

		}
	})

	return result
}

func ZRANGEBYSCORE(c *storage.Cache, z, s, e string) string {
	var result string

	if (s == "max" && e == "min") || (e == "max" && s == "min") {
		c.WithRWLock(func() {
			var arr []ZItem

			cd, exists := c.GetUnsafe(z)
			if !exists {
				// result = formatter.ErrorMessage("can`t find %v in memory", z)
				result = formatter.Array("[]")
				return
			}

			cd.Requests++

			m, ok := parseZSet(cd.Value)
			if !ok {
				// result = formatter.ErrorMessage("%s isn't a zset", z)
				result = formatter.ErrMismatchType.Error()
				return
			}

			if s == "max" && e == "min" {
				for _, k := range m.Order {
					arr = append(arr, k)
				}
			} else if e == "max" && s == "min" {
				for i := len(m.Order) - 1; i >= 0; i-- {
					arr = append(arr, m.Order[i])
				}
			}

			result = formatter.Array(serializeZItems(arr...))
		})
	} else {
		return formatter.ErrInvalidSyntax.Error()
	}

	return result
}

func SCORE(c *storage.Cache, z, k string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			// formatter.ErrorMessage("Can`t find %v in memory", z)
			result = formatter.Number(-1)
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				if s, ok := zs.Items[k]; ok {
					result = formatter.Number(s)
				} else {
					result = formatter.Number(-1)
				}
			} else {
				// formatter.ErrorMessage("%s isn't a zset", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

func LSCORE(c *storage.Cache, z string, ks []string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, 0, len(ks))

		cd, exists := c.GetUnsafe(z)
		if !exists {
			// result = formatter.ErrorMessage("Can`t find %v in memory", z)
			result = formatter.Array("[]")
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				for _, k := range ks {
					if s, ok := zs.Items[k]; ok {
						arr = append(arr, serializeFloat(s))
					} else {
						arr = append(arr, "-1")
					}
				}
			} else {
				// formatter.ErrorMessage("%s isn't a zset", z)
				result = formatter.ErrMismatchType.Error()
				return
			}
		}

		result = formatter.Array(serializeList(arr))
	})

	return result
}
