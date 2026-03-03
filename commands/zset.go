package commands

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
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

func updateZOrder(s []ZItem, i ZItem) []ZItem {
	found := false

	for index, v := range s {
		if v.Member == i.Member {
			s[index].Score = i.Score
			found = true
			break
		}
	}

	if !found {
		s = append(s, i)
	}

	sort.Slice(s, func(a, b int) bool {
		return s[a].Score > s[b].Score
	})

	return s
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

/*
Add one member to a sorted set.

Description:

	Adds the specified values to the zset with score. Duplicate values are ignored.

Example:
  - Pattern: ZADD ZSET_NAME "VALUE" 120

Notes:
  - Returns 1 on success, 0 on failure.
  - If the member already exists, its score is updated.
*/
func ZADD(c *storage.Cache, z, v, s string) string {
	var result string

	score, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return protocol.ErrNotANumber.Error()
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
			result = protocol.Success()
			return
		}

		cd.Requests++
		zs, ok := parseZSet(cd.Value)
		if !ok {
			result = protocol.ErrMismatchType.Error()
			return
		}

		if _, existsInSet := zs.Items[v]; existsInSet {
			zs.Items[v] = score
			zs.Order = updateZOrder(zs.Order, ZItem{
				Member: v,
				Score:  score,
			})
			cd.Value = serializeZSet(*zs)

			result = protocol.Success()
		} else {
			zs.Items[v] = score
			zs.Order = addZOrder(zs.Order, ZItem{
				Member: v,
				Score:  score,
			})
			cd.Value = serializeZSet(*zs)

			result = protocol.Success()
		}
	})

	return result
}

/*
Remove one member from a sorted set.

Description:

	Removes the specified MEMBER from the sorted set stored at ZSET_NAME.

Example:
  - Pattern: ZREM ZSET_NAME "VALUE"

Notes:
  - Returns 1 on success, 0 on failure.
*/
func ZREM(c *storage.Cache, z, v string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Failure()
			return
		} else {
			cd.Requests++

			zs, ok := parseZSet(cd.Value)
			if !ok {
				result = protocol.ErrMismatchType.Error()
				return
			}

			if _, existsInSet := zs.Items[v]; existsInSet {
				delete(zs.Items, v)

				zs.Order = removeZOrder(zs.Order, v)
				cd.Value = serializeZSet(*zs)
				result = protocol.Success()
			} else {
				result = protocol.Failure()
			}

		}
	})

	return result
}

/*
Return a range of members from a sorted set.

Description:

	Returns the members of the sorted set stored at ZSET_NAME from position MIN to MAX.

Example:
  - Pattern: ZRANGE ZSET_NAME MIN MAX

Notes:
  - Returns an array of members in the specified range.
  - IN FUTURE WILL BE ADDED ZRANGEBYSCORE
*/
func ZRANGE(c *storage.Cache, z, s, e string) string {
	var result string

	if (s == "max" && e == "min") || (e == "max" && s == "min") {
		c.WithRWLock(func() {
			var arr []ZItem

			cd, exists := c.GetUnsafe(z)
			if !exists {
				result = protocol.Array("[]")
				return
			}

			cd.Requests++

			m, ok := parseZSet(cd.Value)
			if !ok {
				result = protocol.ErrMismatchType.Error()
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

			result = protocol.Array(serializeZItems(arr...))
		})
	} else {
		return protocol.ErrInvalidSyntax.Error()
	}

	return result
}

/*
Get the score of a member in a sorted set.

Description:

	Returns the score associated with the specified MEMBER in the sorted set stored at ZSET_NAME.

Example:
  - Pattern: SCORE ZSET_NAME "VALUE"

Notes:
  - Returns the score as a string or null if the member does not exist.
*/
func SCORE(c *storage.Cache, z, k string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Nil()
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				if s, ok := zs.Items[k]; ok {
					result = protocol.Number(s)
				} else {
					result = protocol.Number(-1)
				}
			} else {
				result = protocol.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

/*
Get the scores of specific members in a sorted set.

Description:

	Returns an array of scores corresponding to the specified MEMBERS in the list stored at ZSET_NAME.

Example:
  - Pattern: LSCORE ZSET_NAME "VALUE_1" "VALUE_2"

Notes:
  - Returns an array of numbers or nulls.
*/
func LSCORE(c *storage.Cache, z string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, 0, len(args))

		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Array("[]")
			return
		} else {
			cd.Requests++
			if zs, ok := parseZSet(cd.Value); ok {
				for _, k := range args {
					if s, ok := zs.Items[k]; ok {
						arr = append(arr, serializeFloat(s))
					} else {
						arr = append(arr, "-nil")
					}
				}
			} else {
				result = protocol.ErrMismatchType.Error()
				return
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}
