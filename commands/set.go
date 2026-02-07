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

/*
Add one or more members to a set.

Description:

	Adds the specified values to the set. Duplicate values are ignored.

Example:
  - Pattern: SADD SET_NAME "VALUE_1" "VALUE_2"

Notes:
  - Returns the number of values that were actually added to the set.
*/
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
				result = protocol.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

/*
Remove one or more members from a set.

Description:

	Removes the specified values from the set.

Example:
  - Pattern: SREM SET_NAME "VALUE_1" "VALUE_2"

Notes:
  - Returns the number of values that were actually removes from the set.
*/
func SREM(c *storage.Cache, z string, args ...string) string {
	var result string

	if len(args) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(z)
		if !exists {
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
				result = protocol.ErrMismatchType.Error()
				return
			}
		}
	})

	return result
}

/*
Check if a values exists in a set.

Description:

	Checks if the specified value is a member of the set.

Example:
  - Pattern: SCONTAINS SET_NAME "VALUE_1" "VALUE_2"

Notes:
  - Returns the number of specified values that exist in the set.
*/
func SCONTAINS(c *storage.Cache, z string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int
		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
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
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

/*
Check if values exist in a set.

Description:

	Checks if the specified values exist in the set.

Example:

  - Pattern: LSCONTAINS SET_NAME "VALUE_1" "VALUE_2"

  - Result: [1, 0]

Notes:
  - Returns an array of 1s and 0s, where 1 indicates the value exists and 0 indicates it does not.
*/
func LSCONTAINS(c *storage.Cache, z string, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(args))

		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
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
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

/*
Get the number of stored values in set.

Description:

	Returns a number of all stored values in set.

Example:

  - Pattern: SLEN SET_NAME

  - Result: 2

Notes:
  - Returns the number of all values that exist in the set.
*/
func SLEN(c *storage.Cache, z string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
				result = protocol.ErrMismatchType.Error()
				return
			}
			result = protocol.Number(len(m))
		case storage.ZSet:
			m, ok := parseZSet(cd.Value)
			if !ok {
				result = protocol.ErrMismatchType.Error()
				return
			}
			result = protocol.Number(len(m.Items))
		default:
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}

/*
Get all members from a set.

Description:

	Returns a list of all values stored in the set.

Example:
  - Pattern: SMEMBERS SET_NAME

Notes:
  - Returns an array of all values that stored in the set.
*/
func SMEMBERS(c *storage.Cache, z string) string {
	var result string

	c.WithRWLock(func() {
		var arr []string

		cd, exists := c.GetUnsafe(z)
		if !exists {
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		switch cd.Type {
		case storage.Set:
			m, ok := parseSet(cd.Value)
			if !ok {
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
				result = protocol.ErrMismatchType.Error()
				return
			}

			for _, k := range m.Order {
				arr = append(arr, k.Member)
			}

			result = protocol.Array(serializeList(arr))
		default:
			result = protocol.ErrMismatchType.Error()
			return
		}
	})

	return result
}
